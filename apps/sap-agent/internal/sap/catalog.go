package sap

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

type SAPItemGroup struct {
	Number int    `json:"Number"`
	Name   string `json:"GroupName"`
}

type SAPUoM struct {
	Entry int    `json:"AbsEntry"`
	Code  string `json:"Code"`
	Name  string `json:"Name"`
}

type SAPItemPrice struct {
	PriceList   int     `json:"PriceList"`
	Price       float64 `json:"Price"`
	Currency    string  `json:"Currency"`
	UoMEntry    *int    `json:"UoMEntry,omitempty"`
	BasePriceID *int    `json:"BasePriceList,omitempty"`
}

type SAPBarCode struct {
	AbsEntry int     `json:"AbsEntry"`
	UoMEntry *int    `json:"UoMEntry,omitempty"`
	Barcode  string  `json:"BarCode"`
	ItemNo   string  `json:"ItemNo"`
}

type SAPWarehouseInfo struct {
	WarehouseCode string  `json:"WarehouseCode"`
	InStock       float64 `json:"InStock"`
	Committed     float64 `json:"Committed"`
	Ordered       float64 `json:"Ordered"`
}

type SAPItem struct {
	ItemCode                    string             `json:"ItemCode"`
	ItemName                    string             `json:"ItemName"`
	ForeignName                 string             `json:"ForeignName"`
	ItemsGroupCode              int                `json:"ItemsGroupCode"`
	SalesVATGroup               string             `json:"SalesVATGroup"`
	DefaultSalesUoMEntry        *int               `json:"DefaultSalesUoMEntry"`
	InventoryUOM                string             `json:"InventoryUOM"`
	BarCode                     string             `json:"BarCode"`
	ItemPrices                  []SAPItemPrice     `json:"ItemPrices"`
	ItemBarCodeCollection       []SAPBarCode       `json:"ItemBarCodeCollection"`
	ItemWarehouseInfoCollection []SAPWarehouseInfo `json:"ItemWarehouseInfoCollection"`
	UpdateDate                  *string            `json:"UpdateDate"`
	UpdateTime                  *string            `json:"UpdateTime"`
	Valid                       string             `json:"Valid"`
	Frozen                      string             `json:"Frozen"`
	UserFields                  map[string]any     `json:"-"`
}

type ODataResponse[T any] struct {
	Value    []T    `json:"value"`
	NextLink string `json:"@odata.nextLink,omitempty"`
}

func (c *Client) FetchItemGroups(ctx context.Context) ([]SAPItemGroup, error) {
	var resp ODataResponse[SAPItemGroup]
	endpoint := "ItemGroups?$select=Number,GroupName"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch item groups: %w", err)
	}
	return resp.Value, nil
}

func (c *Client) FetchUnitsOfMeasure(ctx context.Context) ([]SAPUoM, error) {
	var resp ODataResponse[SAPUoM]
	endpoint := "UnitOfMeasurements?$select=AbsEntry,Code,Name"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch units of measure: %w", err)
	}
	return resp.Value, nil
}

func (c *Client) FetchItems(ctx context.Context, updatedSince *time.Time, top, skip int) ([]SAPItem, bool, error) {
	var resp ODataResponse[SAPItem]

	params := url.Values{}
	params.Set("$top", fmt.Sprintf("%d", top))
	params.Set("$skip", fmt.Sprintf("%d", skip))
	params.Set("$select", "ItemCode,ItemName,ForeignName,ItemsGroupCode,SalesVATGroup,DefaultSalesUoMEntry,InventoryUOM,BarCode,ItemPrices,ItemBarCodeCollection,ItemWarehouseInfoCollection,UpdateDate,UpdateTime,Valid,Frozen")

	if updatedSince != nil && !updatedSince.IsZero() {
		dateStr := updatedSince.Format("2006-01-02")
		params.Set("$filter", fmt.Sprintf("UpdateDate ge '%s' or CreateDate ge '%s'", dateStr, dateStr))
	}

	endpoint := fmt.Sprintf("Items?%s", params.Encode())
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, false, fmt.Errorf("failed to fetch items from SAP: %w", err)
	}

	hasNext := resp.NextLink != "" || len(resp.Value) == top
	return resp.Value, hasNext, nil
}
