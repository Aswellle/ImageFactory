package service

import (
	"context"
	"strings"

	"github.com/imageforge/imageforge/ent"
	"github.com/imageforge/imageforge/ent/prompttemplate"
	"github.com/imageforge/imageforge/internal/pkg/errors"
)

type PromptTemplateService struct {
	db *ent.Client
}

// NewPromptTemplateService builds a PromptTemplateService.
func NewPromptTemplateService(db *ent.Client) *PromptTemplateService {
	return &PromptTemplateService{db: db}
}

// TemplateCreateInput is the input for creating a template.
type TemplateCreateInput struct {
	UserID      int64
	Name        string
	Description string
	Content     string
	Variables   []string
	Category    string
}

// TemplateUpdateInput is the input for updating a template (zero-values mean "no change").
type TemplateUpdateInput struct {
	Name        *string
	Description *string
	Content     *string
	Variables   []string
	Category    *string
	// ClearVariables forces the variables field to be emptied when true.
	ClearVariables bool
}

// BuiltInTemplate is a system-provided template not bound to any user.
type BuiltInTemplate struct {
	ID          int64                        `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Content     string                       `json:"content"`
	Variables   []string                     `json:"variables"`
	Category    string                       `json:"category"`
	Examples    map[string]map[string]string `json:"examples,omitempty"`
}

// Create persists a new prompt template.
func (s *PromptTemplateService) Create(ctx context.Context, input TemplateCreateInput) (*ent.PromptTemplate, error) {
	if input.Name == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "template name is required")
	}
	if input.Content == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "template content is required")
	}

	builder := s.db.PromptTemplate.Create().
		SetUserID(input.UserID).
		SetName(input.Name).
		SetContent(input.Content)

	if input.Description != "" {
		builder = builder.SetDescription(input.Description)
	}
	if len(input.Variables) > 0 {
		builder = builder.SetVariables(joinVariables(input.Variables))
	}
	if input.Category != "" {
		c := prompttemplate.Category(input.Category)
		if err := prompttemplate.CategoryValidator(c); err != nil {
			return nil, errors.New(errors.ErrInvalidRequest, "invalid category")
		}
		builder = builder.SetCategory(c)
	}

	t, err := builder.Save(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create prompt template", err)
	}
	return t, nil
}

// Get returns a template by ID, enforcing ownership.
func (s *PromptTemplateService) Get(ctx context.Context, templateID, userID int64) (*ent.PromptTemplate, error) {
	t, err := s.db.PromptTemplate.Query().
		Where(prompttemplate.ID(templateID), prompttemplate.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "prompt template not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to fetch prompt template", err)
	}
	return t, nil
}

// List returns all templates for a user, newest first.
func (s *PromptTemplateService) List(ctx context.Context, userID int64) ([]*ent.PromptTemplate, error) {
	templates, err := s.db.PromptTemplate.Query().
		Where(prompttemplate.UserID(userID)).
		Order(ent.Desc(prompttemplate.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to list prompt templates", err)
	}
	return templates, nil
}

// Update modifies a template. Pointers distinguish "leave unchanged" from "set to empty".
func (s *PromptTemplateService) Update(ctx context.Context, templateID, userID int64, input TemplateUpdateInput) (*ent.PromptTemplate, error) {
	builder := s.db.PromptTemplate.UpdateOneID(templateID)

	if input.Name != nil {
		builder = builder.SetName(*input.Name)
	}
	if input.Description != nil {
		builder = builder.SetDescription(*input.Description)
	}
	if input.Content != nil {
		builder = builder.SetContent(*input.Content)
	}
	if input.ClearVariables {
		builder = builder.ClearVariables()
	} else if input.Variables != nil {
		builder = builder.SetVariables(joinVariables(input.Variables))
	}
	if input.Category != nil {
		c := prompttemplate.Category(*input.Category)
		if err := prompttemplate.CategoryValidator(c); err != nil {
			return nil, errors.New(errors.ErrInvalidRequest, "invalid category")
		}
		builder = builder.SetCategory(c)
	}

	t, err := builder.Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrNotFound, "prompt template not found")
		}
		return nil, errors.Wrap(errors.ErrInternal, "failed to update prompt template", err)
	}
	// Enforce ownership: the update above did not filter by user, so verify.
	if t.UserID != userID {
		return nil, errors.New(errors.ErrNotFound, "prompt template not found")
	}
	return t, nil
}

// Delete removes a template owned by the user.
func (s *PromptTemplateService) Delete(ctx context.Context, templateID, userID int64) error {
	n, err := s.db.PromptTemplate.Delete().
		Where(prompttemplate.ID(templateID), prompttemplate.UserID(userID)).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to delete prompt template", err)
	}
	if n == 0 {
		return errors.New(errors.ErrNotFound, "prompt template not found")
	}
	return nil
}

// Apply fills a template's {{variable}} placeholders with the given values.
// The template is loaded and ownership is enforced before substitution.
func (s *PromptTemplateService) Apply(ctx context.Context, templateID, userID int64, variables map[string]string) (string, error) {
	t, err := s.Get(ctx, templateID, userID)
	if err != nil {
		return "", err
	}
	return ApplyVariables(t.Content, variables), nil
}

// ApplyVariables performs {{variable}} substitution on an arbitrary template string.
// Unmatched placeholders are left intact so the user can see what was not filled.
func ApplyVariables(content string, variables map[string]string) string {
	prompt := content
	for k, v := range variables {
		prompt = strings.ReplaceAll(prompt, "{{"+k+"}}", v)
	}
	return prompt
}

// GetBuiltIn returns the system-provided prompt templates.
func GetBuiltIn() []BuiltInTemplate {
	return []BuiltInTemplate{
		{
			ID:          -1,
			Name:        "Product Photography — Studio",
			Description: "Clean studio product shot on a seamless background.",
			Content:     "A professional product photograph of {{product}}, centered on a seamless white background. {{style}} style, {{lighting}} lighting. {{requirements}}",
			Variables:   []string{"product", "style", "lighting", "requirements"},
			Category:    "product",
			Examples: map[string]map[string]string{
				"perfume": {
					"product":      "a minimalist glass perfume bottle with gold cap",
					"style":        "luxury commercial",
					"lighting":     "soft diffused key light with gentle rim light",
					"requirements": "shallow depth of field, reflective surface, high-end catalog look",
				},
			},
		},
		{
			ID:          -2,
			Name:        "Product Photography — Lifestyle",
			Description: "Product shown in a realistic lifestyle setting.",
			Content:     "A lifestyle product photo of {{product}} placed in {{scene}}. Captured in {{style}} style with {{lighting}} lighting. {{requirements}}",
			Variables:   []string{"product", "scene", "style", "lighting", "requirements"},
			Category:    "product",
		},
		{
			ID:          -3,
			Name:        "Social Media — Instagram Post",
			Description: "Eye-catching square image optimized for social feeds.",
			Content:     "An eye-catching social media image featuring {{product}} in {{scene}}. {{style}} aesthetic, {{lighting}} mood. Composition optimized for a square crop. {{requirements}}",
			Variables:   []string{"product", "scene", "style", "lighting", "requirements"},
			Category:    "scene",
		},
		{
			ID:          -4,
			Name:        "Social Media — Cover Banner",
			Description: "Wide banner image for channel covers and headers.",
			Content:     "A wide cinematic banner showcasing {{product}} within {{scene}}. {{style}} look, {{lighting}} atmosphere. Leave negative space on the right for text overlay. {{requirements}}",
			Variables:   []string{"product", "scene", "style", "lighting", "requirements"},
			Category:    "scene",
		},
		{
			ID:          -5,
			Name:        "Artistic Style — Cinematic",
			Description: "Dramatic cinematic frame with film-like color grading.",
			Content:     "A cinematic shot of {{product}} in {{scene}}. {{style}} mood, {{lighting}}. Film-like color grading, anamorphic vibe. {{requirements}}",
			Variables:   []string{"product", "scene", "style", "lighting", "requirements"},
			Category:    "style",
		},
		{
			ID:          -6,
			Name:        "Artistic Style — Minimalist",
			Description: "Clean, minimal composition with lots of negative space.",
			Content:     "A minimalist composition featuring {{product}} against {{scene}}. {{style}} aesthetic, {{lighting}}. Plenty of negative space, restrained palette. {{requirements}}",
			Variables:   []string{"product", "scene", "style", "lighting", "requirements"},
			Category:    "style",
		},
	}
}

// joinVariables serializes variable names to a comma-separated string.
// Empty entries are skipped; "product,scene," becomes "product,scene".
func joinVariables(vars []string) string {
	out := make([]string, 0, len(vars))
	for _, v := range vars {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, ",")
}

