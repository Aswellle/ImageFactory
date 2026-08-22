package batchimage

import "time"

// Account is a minimal stub of Sub2API's service.Account.
// In the full port, this wraps or aliases ent.Account.
type Account struct {
	ID       int64
	Platform string
	Credentials map[string]string
}

func (a *Account) GetCredential(key string) string {
	if a == nil {
		return ""
	}
	return a.Credentials[key]
}

// BatchImageJob is a minimal stub of Sub2API's service.BatchImageJob.
// In the full port, this mirrors ent.BatchImageJob fields.
type BatchImageJob struct {
	ID              int64
	BatchID         string
	UserID          int64
	AccountID       *int64
	Provider        string
	Model           string
	TaskName        string
	AspectRatio     string
	ImageSize       string
	Status          string
	ProviderJobName *string
	ProviderInputRef *string
	ProviderOutputRef *string
	ItemCount       int
	SuccessCount    int
	FailCount       int
	CancelledCount  int
	EstimatedCost   float64
	HoldAmount      *float64
	ActualCost      *float64
	Currency        string
	HoldID          *string
	IdempotencyKey  *string
	RetryCount      int
	Version         int
	LastErrorCode   *string
	LastErrorMessage *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SubmittedAt     *time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	SettledAt       *time.Time
}
