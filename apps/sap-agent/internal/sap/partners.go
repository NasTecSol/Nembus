package sap

import (
	"context"
	"fmt"
)

type SAPBusinessPartner struct {
	CardCode     string `json:"CardCode"`
	CardName     string `json:"CardName"`
	CardType     string `json:"CardType"` // 'cCustomer', 'cSupplier'
	Phone1       string `json:"Phone1,omitempty"`
	EmailAddress string `json:"EmailAddress,omitempty"`
}

func (c *Client) PostBusinessPartner(ctx context.Context, bp *SAPBusinessPartner) (*SAPBusinessPartner, error) {
	var resp SAPBusinessPartner
	if err := c.DoRequest(ctx, "POST", "BusinessPartners", bp, &resp); err != nil {
		return nil, fmt.Errorf("failed to create business partner in SAP: %w", err)
	}
	return &resp, nil
}
