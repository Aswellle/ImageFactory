// Package service implements business logic for ImageForge.
package service

import (
	"context"
	"time"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/usagerecord"
)

// UsageStat is one daily bucket in a user's usage history.
type UsageStat struct {
	Date     string  `json:"date"`
	Requests int64   `json:"requests"`
	Images   int64   `json:"images"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// UsageSummary is the payload returned by GET /v1/usage: totals for the
// current billing period plus a per-day breakdown suitable for charts.
type UsageSummary struct {
	TotalRequests int64       `json:"total_requests"`
	TotalImages   int64       `json:"total_images"`
	TotalTokens   int64       `json:"total_tokens"`
	TotalCost     float64     `json:"total_cost"`
	PeriodStart   string      `json:"period_start"`
	PeriodEnd     string      `json:"period_end"`
	Daily         []UsageStat `json:"daily"`
}

// UsageService aggregates usage_records into dashboard-ready summaries.
type UsageService struct {
	db *ent.Client
}

// NewUsageService builds a UsageService.
func NewUsageService(db *ent.Client) *UsageService {
	return &UsageService{db: db}
}

// RecordUsage appends a single usage row. requestType must be one of the
// usagerecord enum values (generation, edit, upload, storage); unknown values
// fall back to generation.
func (s *UsageService) RecordUsage(ctx context.Context, userID int64, requestType string, count int) error {
	if count <= 0 {
		count = 1
	}
	typ := usagerecord.Type(requestType)
	if err := usagerecord.TypeValidator(typ); err != nil {
		typ = usagerecord.TypeGeneration
	}
	builder := s.db.UsageRecord.Create().
		SetUserID(userID).
		SetType(typ).
		SetImageCount(count)
	_, err := builder.Save(ctx)
	return err
}

// GetUserUsage returns aggregate totals for the trailing 30-day period together
// with a per-day breakdown.
func (s *UsageService) GetUserUsage(ctx context.Context, userID int64) (*UsageSummary, error) {
	const periodDays = 30
	start := time.Now().AddDate(0, 0, -periodDays)

	rows, err := s.db.UsageRecord.Query().
		Where(
			usagerecord.UserIDEQ(userID),
			usagerecord.CreatedAtGTE(start),
		).
		Order(usagerecord.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	summary := &UsageSummary{
		PeriodStart: start.Format("2006-01-02"),
		PeriodEnd:   time.Now().Format("2006-01-02"),
	}

	buckets := map[string]*UsageStat{}
	for _, r := range rows {
		summary.TotalRequests++
		summary.TotalImages += int64(r.ImageCount)
		summary.TotalTokens += int64(r.Tokens)
		summary.TotalCost += r.Cost

		day := r.CreatedAt.Format("2006-01-02")
		b, ok := buckets[day]
		if !ok {
			b = &UsageStat{Date: day}
			buckets[day] = b
		}
		b.Requests++
		b.Images += int64(r.ImageCount)
		b.Tokens += int64(r.Tokens)
		b.Cost += r.Cost
	}

	// Emit a contiguous daily series so charts have no gaps.
	summary.Daily = make([]UsageStat, 0, periodDays)
	for i := range periodDays {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		if b, ok := buckets[day]; ok {
			summary.Daily = append(summary.Daily, *b)
		} else {
			summary.Daily = append(summary.Daily, UsageStat{Date: day})
		}
	}

	return summary, nil
}

// GetUsageHistory returns a per-day breakdown for the last `days` days.
func (s *UsageService) GetUsageHistory(ctx context.Context, userID int64, days int) ([]UsageStat, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -days)

	rows, err := s.db.UsageRecord.Query().
		Where(
			usagerecord.UserIDEQ(userID),
			usagerecord.CreatedAtGTE(start),
		).
		Order(usagerecord.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	buckets := map[string]*UsageStat{}
	for _, r := range rows {
		day := r.CreatedAt.Format("2006-01-02")
		b, ok := buckets[day]
		if !ok {
			b = &UsageStat{Date: day}
			buckets[day] = b
		}
		b.Requests++
		b.Images += int64(r.ImageCount)
		b.Tokens += int64(r.Tokens)
		b.Cost += r.Cost
	}

	history := make([]UsageStat, 0, days)
	for i := range days {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		if b, ok := buckets[day]; ok {
			history = append(history, *b)
		} else {
			history = append(history, UsageStat{Date: day})
		}
	}

	return history, nil
}
