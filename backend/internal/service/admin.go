// Package service implements business logic for ImageForge.
package service

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/apikey"
	"github.com/imageforge/imageforge/ent/asset"
	"github.com/imageforge/imageforge/ent/generationjob"
	"github.com/imageforge/imageforge/ent/user"
)

// --- Admin DTOs ---

type AdminUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminJob struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	ImageCount   int    `json:"image_count"`
	RetryCount   int    `json:"retry_count"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type AdminAPIKey struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	KeyPrefix string `json:"key_prefix"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	TotalJobs     int64 `json:"total_jobs"`
	CompletedJobs int64 `json:"completed_jobs"`
	FailedJobs    int64 `json:"failed_jobs"`
	TotalAssets   int64 `json:"total_assets"`
	TotalAPIKeys  int64 `json:"total_api_keys"`
}

type DashboardPayload struct {
	DashboardStats
	RecentActivity []ActivityItem `json:"recent_activity"`
}

type ActivityItem struct {
	Type      string    `json:"type"`
	JobID     int64     `json:"job_id"`
	UserID    int64     `json:"user_id"`
	Status    string    `json:"status"`
	Prompt    string    `json:"prompt"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserInput struct {
	Role   *string `json:"role,omitempty"`
	Status *string `json:"status,omitempty"`
	Name   *string `json:"name,omitempty"`
}

// --- Admin Service ---

type AdminService interface {
	Dashboard(ctx context.Context) (*DashboardPayload, error)
	UserList(ctx context.Context, page, pageSize int, status, role, search string) ([]AdminUser, int, error)
	UpdateUser(ctx context.Context, userID int64, input UpdateUserInput) (*AdminUser, error)
	SetUserStatus(ctx context.Context, userID int64, status string) error
	JobList(ctx context.Context, page, pageSize int, status, search string) ([]AdminJob, int, error)
	RetryJob(ctx context.Context, jobID int64) error
	APIKeyList(ctx context.Context, page, pageSize int, status, search string) ([]AdminAPIKey, int, error)
	RevokeAPIKey(ctx context.Context, keyID int64) error
}

type adminServiceImpl struct {
	db *ent.Client
}

func NewAdminService(db *ent.Client) AdminService {
	return &adminServiceImpl{db: db}
}

func (s *adminServiceImpl) Dashboard(ctx context.Context) (*DashboardPayload, error) {
	stats := &DashboardStats{}

	if c, err := s.db.User.Query().Count(ctx); err == nil {
		stats.TotalUsers = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.User.Query().Where(user.StatusEQ("active")).Count(ctx); err == nil {
		stats.ActiveUsers = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.GenerationJob.Query().Count(ctx); err == nil {
		stats.TotalJobs = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.GenerationJob.Query().Where(generationjob.StatusEQ(generationjob.StatusCompleted)).Count(ctx); err == nil {
		stats.CompletedJobs = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.GenerationJob.Query().Where(generationjob.StatusEQ(generationjob.StatusFailed)).Count(ctx); err == nil {
		stats.FailedJobs = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.Asset.Query().Where(asset.StatusEQ("active")).Count(ctx); err == nil {
		stats.TotalAssets = int64(c)
	} else {
		return nil, err
	}
	if c, err := s.db.APIKey.Query().Count(ctx); err == nil {
		stats.TotalAPIKeys = int64(c)
	} else {
		return nil, err
	}

	activity, err := s.recentActivity(ctx, 20)
	if err != nil {
		return nil, err
	}

	return &DashboardPayload{
		DashboardStats: *stats,
		RecentActivity: activity,
	}, nil
}

func (s *adminServiceImpl) recentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	jobs, err := s.db.GenerationJob.Query().
		Order(generationjob.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ActivityItem, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, ActivityItem{
			Type:      "job",
			JobID:     j.ID,
			UserID:    j.UserID,
			Status:    string(j.Status),
			Prompt:    truncate(j.Prompt, 120),
			Model:     j.Model,
			CreatedAt: j.CreatedAt,
		})
	}
	return items, nil
}

func (s *adminServiceImpl) UserList(ctx context.Context, page, pageSize int, status, role, search string) ([]AdminUser, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	q := s.db.User.Query()
	if status != "" {
		q = q.Where(user.StatusEQ(user.Status(status)))
	}
	if role != "" {
		q = q.Where(user.RoleEQ(user.Role(role)))
	}
	if search != "" {
		term := strings.ToLower(search)
		q = q.Where(user.Or(
			user.EmailContainsFold(term),
			user.NameContainsFold(term),
		))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.
		Order(user.ByID(sql.OrderDesc())).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	users := make([]AdminUser, 0, len(rows))
	for _, u := range rows {
		users = append(users, AdminUser{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      string(u.Role),
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt,
		})
	}
	return users, total, nil
}

func (s *adminServiceImpl) UpdateUser(ctx context.Context, userID int64, input UpdateUserInput) (*AdminUser, error) {
	up := s.db.User.UpdateOneID(userID)
	if input.Role != nil {
		up.SetRole(user.Role(*input.Role))
	}
	if input.Status != nil {
		up.SetStatus(user.Status(*input.Status))
	}
	if input.Name != nil {
		up.SetName(*input.Name)
	}
	up.SetUpdatedAt(time.Now())
	u, err := up.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &AdminUser{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *adminServiceImpl) SetUserStatus(ctx context.Context, userID int64, status string) error {
	st := user.Status(status)
	if st != user.StatusActive && st != "suspended" {
		st = "suspended"
	}
	_, err := s.db.User.UpdateOneID(userID).
		SetStatus(st).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (s *adminServiceImpl) JobList(ctx context.Context, page, pageSize int, status, search string) ([]AdminJob, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	q := s.db.GenerationJob.Query()
	if status != "" {
		q = q.Where(generationjob.StatusEQ(generationjob.Status(status)))
	}
	if search != "" {
		term := strings.ToLower(search)
		q = q.Where(generationjob.Or(
			generationjob.PromptContainsFold(term),
			generationjob.ErrorMessageContainsFold(term),
		))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.
		Order(generationjob.ByID(sql.OrderDesc())).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	jobs := make([]AdminJob, 0, len(rows))
	for _, j := range rows {
		jobs = append(jobs, AdminJob{
			ID:           j.ID,
			UserID:       j.UserID,
			Type:         string(j.Type),
			Status:       string(j.Status),
			Provider:     j.Provider,
			Model:        j.Model,
			Prompt:       truncate(j.Prompt, 100),
			ImageCount:   j.ImageCount,
			RetryCount:   j.RetryCount,
			ErrorCode:    j.ErrorCode,
			ErrorMessage: j.ErrorMessage,
		})
	}
	return jobs, total, nil
}

func (s *adminServiceImpl) RetryJob(ctx context.Context, jobID int64) error {
	j, err := s.db.GenerationJob.Query().Where(generationjob.IDEQ(jobID)).Only(ctx)
	if err != nil {
		return err
	}
	if j.Status != generationjob.StatusFailed {
		return ErrNotRetryable
	}
	_, err = s.db.GenerationJob.UpdateOneID(jobID).
		SetStatus(generationjob.StatusPending).
		SetRetryCount(j.RetryCount + 1).
		ClearErrorCode().
		ClearErrorMessage().
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (s *adminServiceImpl) APIKeyList(ctx context.Context, page, pageSize int, status, search string) ([]AdminAPIKey, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	q := s.db.APIKey.Query()
	if status != "" {
		q = q.Where(apikey.StatusEQ(apikey.Status(status)))
	}
	if search != "" {
		term := strings.ToLower(search)
		q = q.Where(apikey.Or(
			apikey.NameContainsFold(term),
			apikey.KeyPrefixContainsFold(term),
		))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.
		Order(apikey.ByID(sql.OrderDesc())).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	keys := make([]AdminAPIKey, 0, len(rows))
	for _, k := range rows {
		keys = append(keys, AdminAPIKey{
			ID:        k.ID,
			UserID:    k.UserID,
			KeyPrefix: k.KeyPrefix,
			Name:      k.Name,
			Status:    string(k.Status),
		})
	}
	return keys, total, nil
}

func (s *adminServiceImpl) RevokeAPIKey(ctx context.Context, keyID int64) error {
	_, err := s.db.APIKey.UpdateOneID(keyID).
		SetStatus(apikey.StatusRevoked).
		Save(ctx)
	return err
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// ErrNotRetryable is returned when a job cannot be retried.
var ErrNotRetryable = &adminError{msg: "job is not in failed state"}

type adminError struct{ msg string }

func (e *adminError) Error() string { return e.msg }
