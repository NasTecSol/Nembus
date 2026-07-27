package usecase

import (
	"context"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// UomPackagingTemplateOutput represents a packaging template with its levels.
type UomPackagingTemplateOutput struct {
	ID             int32                             `json:"id"`
	OrganizationID int32                             `json:"organization_id"`
	Name           string                            `json:"name"`
	Code           string                            `json:"code"`
	IsActive       pgtype.Bool                       `json:"is_active"`
	CreatedAt      pgtype.Timestamp                  `json:"created_at"`
	UpdatedAt      pgtype.Timestamp                  `json:"updated_at"`
	Levels         []UomPackagingTemplateLevelOutput `json:"levels,omitempty"`
}

// UomPackagingTemplateLevelOutput represents a level within a packaging template.
type UomPackagingTemplateLevelOutput struct {
	ID         int32          `json:"id"`
	TemplateID int32          `json:"template_id"`
	LevelOrder int32          `json:"level_order"`
	UomID      int32          `json:"uom_id"`
	Multiplier pgtype.Numeric `json:"multiplier"`
}

func templateToOutput(t repository.UomPackagingTemplate) UomPackagingTemplateOutput {
	return UomPackagingTemplateOutput{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		Name:           t.Name,
		Code:           t.Code,
		IsActive:       t.IsActive,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func levelToOutput(l repository.UomPackagingTemplateLevel) UomPackagingTemplateLevelOutput {
	return UomPackagingTemplateLevelOutput{
		ID:         l.ID,
		TemplateID: l.TemplateID,
		LevelOrder: l.LevelOrder,
		UomID:      l.UomID,
		Multiplier: l.Multiplier,
	}
}

type UomPackagingTemplateUseCase struct {
	repo *repository.Queries
}

func NewUomPackagingTemplateUseCase() *UomPackagingTemplateUseCase {
	return &UomPackagingTemplateUseCase{}
}

func (uc *UomPackagingTemplateUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *UomPackagingTemplateUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateTemplate creates a new UOM packaging template.
func (uc *UomPackagingTemplateUseCase) CreateTemplate(
	ctx context.Context,
	organizationID int32,
	name string,
	code string,
	isActive bool,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}

	active := pgtype.Bool{Bool: isActive, Valid: true}

	row, err := uc.repo.CreateUomPackagingTemplate(ctx, repository.CreateUomPackagingTemplateParams{
		OrganizationID: organizationID,
		Name:           name,
		Code:           code,
		IsActive:       active,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "packaging template created", templateToOutput(row))
}

// GetTemplate retrieves a template by ID.
func (uc *UomPackagingTemplateUseCase) GetTemplate(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetUomPackagingTemplate(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "packaging template not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "packaging template fetched", templateToOutput(row))
}

// ListTemplates lists templates for an organization.
func (uc *UomPackagingTemplateUseCase) ListTemplates(ctx context.Context, organizationID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListUomPackagingTemplates(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]UomPackagingTemplateOutput, len(rows))
	for i := range rows {
		out[i] = templateToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "packaging templates listed", out)
}

// UpdateTemplate updates a template.
func (uc *UomPackagingTemplateUseCase) UpdateTemplate(
	ctx context.Context,
	id string,
	name string,
	code string,
	isActive bool,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.UpdateUomPackagingTemplate(ctx, repository.UpdateUomPackagingTemplateParams{
		ID:       int32(parsed),
		Name:     name,
		Code:     code,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "packaging template updated", templateToOutput(row))
}

// DeleteTemplate deletes a template.
func (uc *UomPackagingTemplateUseCase) DeleteTemplate(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteUomPackagingTemplate(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "packaging template deleted", nil)
}

// CreateLevel creates a level for a template.
func (uc *UomPackagingTemplateUseCase) CreateLevel(
	ctx context.Context,
	templateID int32,
	levelOrder int32,
	uomID int32,
	multiplier string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var mult pgtype.Numeric
	if err := mult.Scan(multiplier); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid multiplier", nil)
	}

	row, err := uc.repo.CreateUomPackagingTemplateLevel(ctx, repository.CreateUomPackagingTemplateLevelParams{
		TemplateID: templateID,
		LevelOrder: levelOrder,
		UomID:      uomID,
		Multiplier: mult,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "template level created", levelToOutput(row))
}

// ListLevels lists levels for a template.
func (uc *UomPackagingTemplateUseCase) ListLevels(ctx context.Context, templateID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListUomPackagingTemplateLevels(ctx, templateID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]UomPackagingTemplateLevelOutput, len(rows))
	for i := range rows {
		out[i] = levelToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "template levels listed", out)
}

// UpdateLevel updates a template level.
func (uc *UomPackagingTemplateUseCase) UpdateLevel(
	ctx context.Context,
	id string,
	levelOrder int32,
	uomID int32,
	multiplier string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	var mult pgtype.Numeric
	if err := mult.Scan(multiplier); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid multiplier", nil)
	}

	row, err := uc.repo.UpdateUomPackagingTemplateLevel(ctx, repository.UpdateUomPackagingTemplateLevelParams{
		ID:         int32(parsed),
		LevelOrder: levelOrder,
		UomID:      uomID,
		Multiplier: mult,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "template level updated", levelToOutput(row))
}

// DeleteLevel deletes a template level.
func (uc *UomPackagingTemplateUseCase) DeleteLevel(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteUomPackagingTemplateLevel(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "template level deleted", nil)
}

// GetTemplateWithLevels retrieves a template and its levels.
func (uc *UomPackagingTemplateUseCase) GetTemplateWithLevels(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	rows, err := uc.repo.GetPackagingTemplateWithLevels(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if len(rows) == 0 {
		return utils.NewResponse(utils.CodeNotFound, "packaging template not found", nil)
	}

	// Group rows by template
	first := rows[0]
	out := UomPackagingTemplateOutput{
		ID:             first.TemplateID,
		OrganizationID: first.OrganizationID,
		Name:           first.Name,
		Code:           first.Code,
		IsActive:       first.IsActive,
		Levels:         []UomPackagingTemplateLevelOutput{},
	}

	for _, row := range rows {
		if row.LevelID.Valid {
			out.Levels = append(out.Levels, UomPackagingTemplateLevelOutput{
				ID:         row.LevelID.Int32,
				TemplateID: row.TemplateID,
				LevelOrder: row.LevelOrder.Int32,
				UomID:      row.UomID.Int32,
				Multiplier: row.Multiplier,
			})
		}
	}

	return utils.NewResponse(utils.CodeOK, "packaging template with levels fetched", out)
}
