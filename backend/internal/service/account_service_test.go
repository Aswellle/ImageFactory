package service

import (
	"testing"

	"github.com/imageforge/imageforge/ent"
)

func TestPickByPriority_Empty(t *testing.T) {
	svc := &AccountService{}
	got := svc.pickByPriority(nil)
	if got != nil {
		t.Fatalf("expected nil for empty input, got %+v", got)
	}
}

func TestPickByPriority_Single(t *testing.T) {
	svc := &AccountService{}
	acc := &ent.Account{ID: 1, Priority: 5}
	got := svc.pickByPriority([]*ent.Account{acc})
	if got.ID != 1 {
		t.Fatalf("expected ID=1, got ID=%d", got.ID)
	}
}

func TestPickByPriority_HighestWins(t *testing.T) {
	svc := &AccountService{}
	accounts := []*ent.Account{
		{ID: 1, Priority: 3},
		{ID: 2, Priority: 1}, // highest (lowest number)
		{ID: 3, Priority: 5},
	}
	got := svc.pickByPriority(accounts)
	if got.ID != 2 {
		t.Fatalf("expected ID=2 (priority 1), got ID=%d (priority %d)", got.ID, got.Priority)
	}
}

func TestPickByPriority_RotationInTier(t *testing.T) {
	svc := &AccountService{}
	accounts := []*ent.Account{
		{ID: 1, Priority: 1},
		{ID: 2, Priority: 1},
		{ID: 3, Priority: 1},
	}

	// Run many times to verify all top-tier accounts get selected.
	counts := map[int64]int{}
	for range 300 {
		got := svc.pickByPriority(accounts)
		counts[got.ID]++
	}

	for _, acc := range accounts {
		if counts[acc.ID] == 0 {
			t.Fatalf("account %d was never selected in 300 trials — rotation is broken", acc.ID)
		}
	}
}

func TestPickByPriority_MixedPriorities(t *testing.T) {
	svc := &AccountService{}
	accounts := []*ent.Account{
		{ID: 1, Priority: 2},
		{ID: 2, Priority: 1}, // top tier
		{ID: 3, Priority: 1}, // top tier
		{ID: 4, Priority: 3},
	}

	// Only priority-1 accounts should ever be selected.
	for range 100 {
		got := svc.pickByPriority(accounts)
		if got.Priority != 1 {
			t.Fatalf("expected priority 1, got priority %d (ID=%d)", got.Priority, got.ID)
		}
	}
}
