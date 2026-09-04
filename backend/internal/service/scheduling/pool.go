package scheduling

import (
	"container/heap"
	"math/rand"
	"sort"
	"time"
)

// AccountPool manages a collection of accounts for scheduling.
//
// Ported from Sub2API's account pool management in:
// - internal/service/openai_account_scheduler.go (load balancing, sticky sessions)
// - internal/service/account_service.go (account selection, priority sorting)
//
// The pool provides:
// 1. Priority-based account selection with round-robin within priority tiers
// 2. Sticky session support (session hash → account binding)
// 3. Load-aware candidate scoring and top-K selection
// 4. Threshold-based pre-filtering

// Pool manages accounts for a single platform.
type Pool struct {
	accounts []*Account
	rng      *rand.Rand
}

// NewPool creates a new account pool.
func NewPool() *Pool {
	return &Pool{
		accounts: make([]*Account, 0),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddAccount adds an account to the pool.
func (p *Pool) AddAccount(account *Account) {
	if account != nil {
		p.accounts = append(p.accounts, account)
	}
}

// AddAccounts adds multiple accounts to the pool.
func (p *Pool) AddAccounts(accounts []*Account) {
	for _, acc := range accounts {
		p.AddAccount(acc)
	}
}

// SelectAvailable returns accounts that are currently schedulable.
// Filters out: inactive, rate-limited, overloaded, temp-unschedulable.
//
// Ported from Sub2API: account_service.go filterAvailable + scheduler candidate filtering
func (p *Pool) SelectAvailable(now time.Time) []*Account {
	var available []*Account
	for _, acc := range p.accounts {
		if acc == nil {
			continue
		}
		if !acc.IsActive() {
			continue
		}
		if acc.RateLimitResetAt != nil && acc.RateLimitResetAt.After(now) {
			continue
		}
		if acc.OverloadUntil != nil && acc.OverloadUntil.After(now) {
			continue
		}
		if acc.TempUnschedulableUntil != nil && acc.TempUnschedulableUntil.After(now) {
			continue
		}
		available = append(available, acc)
	}
	return available
}

// SelectByPriority selects an account using priority-based round-robin.
// Accounts are grouped by priority (lower number = higher priority).
// Within the highest-priority group, selection is randomized for round-robin.
//
// Ported from Sub2API: account_service.go pickByPriority
func (p *Pool) SelectByPriority(now time.Time) *Account {
	available := p.SelectAvailable(now)
	if len(available) == 0 {
		return nil
	}

	// Sort by priority (ascending - lower number = higher priority)
	sort.SliceStable(available, func(i, j int) bool {
		return available[i].Priority < available[j].Priority
	})

	// Find all accounts in the highest priority tier
	bestPriority := available[0].Priority
	var bestTier []*Account
	for _, acc := range available {
		if acc.Priority == bestPriority {
			bestTier = append(bestTier, acc)
		}
	}

	// Random selection within the best tier (round-robin)
	if len(bestTier) == 1 {
		return bestTier[0]
	}
	return bestTier[p.rng.Intn(len(bestTier))]
}

// SelectWithThreshold applies scheduling threshold evaluation before selection.
// Accounts that exceed their usage threshold are excluded.
//
// This combines Sub2API's threshold evaluation with account selection.
func (p *Pool) SelectWithThreshold(now time.Time, thresholds map[string]int) *Account {
	available := p.SelectAvailable(now)
	if len(available) == 0 {
		return nil
	}

	// Filter out accounts that exceed their scheduling threshold
	var eligible []*Account
	for _, acc := range available {
		decision := EvaluateAccountSchedulingThreshold(acc, thresholds, now)
		if !decision.ShouldPause {
			eligible = append(eligible, acc)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	// Sort by priority and select
	sort.SliceStable(eligible, func(i, j int) bool {
		return eligible[i].Priority < eligible[j].Priority
	})

	bestPriority := eligible[0].Priority
	var bestTier []*Account
	for _, acc := range eligible {
		if acc.Priority == bestPriority {
			bestTier = append(bestTier, acc)
		}
	}

	if len(bestTier) == 1 {
		return bestTier[0]
	}
	return bestTier[p.rng.Intn(len(bestTier))]
}

// CandidateScore holds a scored account candidate for load-aware selection.
//
// Ported from Sub2API: openai_account_scheduler.go openAIAccountCandidateScore
type CandidateScore struct {
	Account      *Account
	Score        float64
	Priority     int
	ErrorRate    float64
	LoadFactor   float64
	WaitingCount int
}

// CandidateHeap implements a min-heap for top-K candidate selection.
//
// Ported from Sub2API: openai_account_scheduler.go openAIAccountCandidateHeap
type CandidateHeap []CandidateScore

func (h CandidateHeap) Len() int { return len(h) }
func (h CandidateHeap) Less(i, j int) bool {
	if h[i].Score != h[j].Score {
		return h[i].Score < h[j].Score
	}
	return h[i].Priority > h[j].Priority
}
func (h CandidateHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *CandidateHeap) Push(x any) {
	*h = append(*h, x.(CandidateScore))
}

func (h *CandidateHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// SelectTopK selects the top-K candidates by score using a min-heap.
//
// Ported from Sub2API: openai_account_scheduler.go selectTopKOpenAICandidates
func SelectTopK(candidates []CandidateScore, topK int) []CandidateScore {
	if len(candidates) <= topK {
		result := make([]CandidateScore, len(candidates))
		copy(result, candidates)
		sort.Slice(result, func(i, j int) bool {
			return result[i].Score > result[j].Score
		})
		return result
	}

	h := &CandidateHeap{}
	heap.Init(h)

	for _, c := range candidates {
		if h.Len() < topK {
			heap.Push(h, c)
			continue
		}
		if (*h)[0].Score < c.Score {
			heap.Pop(h)
			heap.Push(h, c)
		}
	}

	result := make([]CandidateScore, 0, topK)
	for h.Len() > 0 {
		result = append(result, heap.Pop(h).(CandidateScore))
	}

	// Reverse to get descending order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// ScoreWeights defines the multi-factor scoring weights for account selection.
//
// Ported from Sub2API: openai_account_scheduler.go GatewayOpenAIWSSchedulerScoreWeightsView
type ScoreWeights struct {
	Priority       float64
	Load           float64
	ErrorRate      float64
	Reset          float64
	QuotaHeadroom  float64
	UpstreamCost   float64
	SessionSticky  float64
}

// DefaultScoreWeights returns the default scoring weights.
func DefaultScoreWeights() ScoreWeights {
	return ScoreWeights{
		Priority:      1.0,
		Load:          0.8,
		ErrorRate:     0.6,
		Reset:         0.4,
		QuotaHeadroom: 0.3,
		UpstreamCost:  0.2,
		SessionSticky: 0.5,
	}
}

// ScoreCandidates scores all candidates using the weighted multi-factor formula.
//
// Ported from Sub2API: openai_account_scheduler.go buildOpenAIAccountLoadPlan
func ScoreCandidates(accounts []*Account, weights ScoreWeights, stickyAccountID *int64) []CandidateScore {
	now := nowFunc()
	candidates := make([]CandidateScore, 0, len(accounts))

	for _, acc := range accounts {
		if acc == nil {
			continue
		}

		cs := CandidateScore{
			Account:  acc,
			Priority: acc.Priority,
		}

		// Priority factor (normalized: lower priority number = higher score)
		maxPriority := 100.0
		cs.Score += weights.Priority * Clamp01(1.0-float64(acc.Priority)/maxPriority)

		// Load factor (prefer less-loaded accounts)
		loadFactor := float64(acc.Priority) / 50.0 // simplified load estimation
		cs.LoadFactor = loadFactor
		cs.Score += weights.Load * Clamp01(1.0-loadFactor)

		// Error rate (prefer accounts with fewer recent errors)
		if acc.Status == StatusError {
			cs.ErrorRate = 1.0
		}
		cs.Score += weights.ErrorRate * Clamp01(1.0-cs.ErrorRate)

		// Reset proximity (prefer accounts further from rate limit reset)
		if acc.RateLimitResetAt != nil && acc.RateLimitResetAt.After(now) {
			timeToReset := acc.RateLimitResetAt.Sub(now).Seconds()
			cs.Score += weights.Reset * Clamp01(1.0/(1.0+timeToReset/60.0))
		} else {
			cs.Score += weights.Reset // No rate limit = full score
		}

		// Sticky session bonus
		if stickyAccountID != nil && acc.ID == *stickyAccountID {
			cs.Score += weights.SessionSticky * 10.0 // Strong sticky bonus
		}

		candidates = append(candidates, cs)
	}

	return candidates
}

// SelectLoadAware performs load-aware account selection with weighted scoring.
//
// This is the ImageForge adaptation of Sub2API's 3-layer cascade scheduler:
// 1. Try sticky session account (if available and healthy)
// 2. Fall back to scored weighted selection from top-K candidates
//
// Ported from Sub2API: openai_account_scheduler.go Select
func (p *Pool) SelectLoadAware(weights ScoreWeights, stickyAccountID *int64, now time.Time) *Account {
	available := p.SelectAvailable(now)
	if len(available) == 0 {
		return nil
	}

	// Layer 1: Try sticky session account
	if stickyAccountID != nil {
		for _, acc := range available {
			if acc.ID == *stickyAccountID {
				return acc
			}
		}
	}

	// Layer 2: Score and select from top-K
	candidates := ScoreCandidates(available, weights, nil)
	if len(candidates) == 0 {
		return nil
	}

	topK := SelectTopK(candidates, min(3, len(candidates)))

	// Weighted random selection from top-K
	if len(topK) == 1 {
		return topK[0].Account
	}

	// Simple weighted random: higher score = higher probability
	totalScore := 0.0
	for _, c := range topK {
		if c.Score < 0 {
			c.Score = 0.01 // Ensure non-negative
		}
		totalScore += c.Score
	}

	if totalScore <= 0 {
		return topK[p.rng.Intn(len(topK))].Account
	}

	r := p.rng.Float64() * totalScore
	for _, c := range topK {
		r -= c.Score
		if r <= 0 {
			return c.Account
		}
	}

	return topK[len(topK)-1].Account
}

// min returns the minimum of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Count returns the total number of accounts in the pool.
func (p *Pool) Count() int {
	return len(p.accounts)
}

// CountAvailable returns the number of currently available accounts.
func (p *Pool) CountAvailable(now time.Time) int {
	return len(p.SelectAvailable(now))
}
