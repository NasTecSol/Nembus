package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type ReceiptTemplateRecord struct {
	ID             int64           `json:"id"`
	OrganizationID int32           `json:"organization_id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Language       string          `json:"language"`
	Config         json.RawMessage `json:"config"`
	IsActive       bool            `json:"is_active"`
	IsDefault      bool            `json:"is_default"`
	UpdatedAt      string          `json:"updated_at"`
}

func (q *Queries) ListReceiptTemplates(ctx context.Context, organizationID int32) ([]ReceiptTemplateRecord, error) {
	rows, err := q.db.Query(ctx, `SELECT id,organization_id,name,type,language,config,is_active,is_default,updated_at::text FROM printing_templates WHERE organization_id=$1 ORDER BY type,language,name`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []ReceiptTemplateRecord{}
	for rows.Next() {
		var v ReceiptTemplateRecord
		if err := rows.Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Type, &v.Language, &v.Config, &v.IsActive, &v.IsDefault, &v.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (q *Queries) SaveReceiptTemplate(ctx context.Context, organizationID int32, id int64, name, templateType, language string, config json.RawMessage, active bool) (ReceiptTemplateRecord, error) {
	var v ReceiptTemplateRecord
	var err error
	if id == 0 {
		err = q.db.QueryRow(ctx, `INSERT INTO printing_templates(organization_id,name,type,language,config,is_active) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,organization_id,name,type,language,config,is_active,is_default,updated_at::text`, organizationID, name, templateType, language, config, active).Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Type, &v.Language, &v.Config, &v.IsActive, &v.IsDefault, &v.UpdatedAt)
		return v, err
	}
	err = q.db.QueryRow(ctx, `UPDATE printing_templates SET name=$3,type=$4,language=$5,config=$6,is_active=$7,updated_at=now() WHERE organization_id=$1 AND id=$2 AND ((type=$4 AND language=$5) OR (NOT is_default AND NOT EXISTS(SELECT 1 FROM store_receipt_settings WHERE organization_id=$1 AND default_template_id=$2))) AND ($7 OR (NOT is_default AND NOT EXISTS(SELECT 1 FROM store_receipt_settings WHERE organization_id=$1 AND default_template_id=$2))) RETURNING id,organization_id,name,type,language,config,is_active,is_default,updated_at::text`, organizationID, id, name, templateType, language, config, active).Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Type, &v.Language, &v.Config, &v.IsActive, &v.IsDefault, &v.UpdatedAt)
	return v, err
}

func (q *Queries) SetOrganizationReceiptDefault(ctx context.Context, orgID int32, templateID int64) error {
	beginner, ok := q.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return fmt.Errorf("organization default update requires a database connection")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var templateType, language string
	err = tx.QueryRow(ctx, `SELECT type,language FROM printing_templates WHERE organization_id=$1 AND id=$2 AND is_active`, orgID, templateID).Scan(&templateType, &language)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE printing_templates SET is_default=false,updated_at=now() WHERE organization_id=$1 AND type=$2 AND language=$3 AND is_default`, orgID, templateType, language); err != nil {
		return err
	}
	var selectedID int64
	err = tx.QueryRow(ctx, `UPDATE printing_templates SET is_default=true,updated_at=now() WHERE organization_id=$1 AND id=$2 AND is_active RETURNING id`, orgID, templateID).Scan(&selectedID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (q *Queries) SetStoreReceiptDefault(ctx context.Context, storeID int64, templateID *int64, templateType, language string) error {
	templateType = strings.ToLower(strings.TrimSpace(templateType))
	language = strings.ToLower(strings.TrimSpace(language))
	var id int64
	if templateID == nil {
		if templateType == "" || language == "" {
			return fmt.Errorf("template type and language are required when inheriting")
		}
		err := q.db.QueryRow(ctx, `INSERT INTO store_receipt_settings(store_id,organization_id,type,language,default_template_id) SELECT s.id,s.organization_id,$2,$3,NULL FROM stores s WHERE s.id=$1 ON CONFLICT(store_id,type,language) DO UPDATE SET default_template_id=NULL,organization_id=EXCLUDED.organization_id,updated_at=now() RETURNING store_id`, storeID, templateType, language).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return err
	}
	err := q.db.QueryRow(ctx, `INSERT INTO store_receipt_settings(store_id,organization_id,type,language,default_template_id) SELECT s.id,s.organization_id,t.type,t.language,t.id FROM stores s JOIN printing_templates t ON t.organization_id=s.organization_id WHERE s.id=$1 AND t.id=$2 AND t.is_active ON CONFLICT(store_id,type,language) DO UPDATE SET default_template_id=EXCLUDED.default_template_id,organization_id=EXCLUDED.organization_id,updated_at=now() RETURNING store_id`, storeID, *templateID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}
	return err
}

func (q *Queries) ResolveReceiptTemplate(ctx context.Context, orgID int32, storeID int64, templateType, language string) (ReceiptTemplateRecord, string, error) {
	var v ReceiptTemplateRecord
	var source string
	err := q.db.QueryRow(ctx, `SELECT id,organization_id,name,type,language,config,is_active,is_default,updated_at::text,scope FROM (
 SELECT t.id,t.organization_id,t.name,t.type,t.language,t.config,t.is_active,t.is_default,t.updated_at,'store'::text scope,1 rank
 FROM store_receipt_settings s JOIN printing_templates t ON t.id=s.default_template_id AND t.organization_id=s.organization_id AND t.type=s.type AND t.language=s.language AND t.is_active
 WHERE s.organization_id=$1 AND s.store_id=$2 AND s.type=$3 AND s.language=$4
 UNION ALL
 SELECT t.id,t.organization_id,t.name,t.type,t.language,t.config,t.is_active,t.is_default,t.updated_at,'organization'::text scope,2 rank
 FROM printing_templates t WHERE t.organization_id=$1 AND t.type=$3 AND t.language=$4 AND t.is_active AND t.is_default
 UNION ALL
 SELECT t.id,t.organization_id,t.name,t.type,t.language,t.config,t.is_active,t.is_default,t.updated_at,'organization_fallback'::text scope,3 rank
 FROM printing_templates t WHERE t.organization_id=$1 AND t.type='receipt' AND t.language='en' AND t.is_active AND t.is_default
) effective ORDER BY rank LIMIT 1`, orgID, storeID, templateType, language).Scan(&v.ID, &v.OrganizationID, &v.Name, &v.Type, &v.Language, &v.Config, &v.IsActive, &v.IsDefault, &v.UpdatedAt, &source)
	return v, source, err
}

func (q *Queries) GetStoreOrganizationID(ctx context.Context, storeID int64) (int32, error) {
	var orgID int32
	err := q.db.QueryRow(ctx, `SELECT organization_id FROM stores WHERE id=$1`, storeID).Scan(&orgID)
	return orgID, err
}
