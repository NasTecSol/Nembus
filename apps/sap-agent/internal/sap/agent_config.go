package sap

import (
	legacy "github.com/NasTecSol/nembus-sap-agent/config"
	agentcfg "github.com/NasTecSol/nembus-sap-agent/internal/config"
)

// LegacyConfig builds the Service Layer client config from agent_config.json.
func LegacyConfig(ac *agentcfg.AgentConfig) *legacy.Config {
	if ac == nil {
		ac = agentcfg.DefaultConfig()
	}
	companyDB := ac.SAP.CompanyDB
	if companyDB == "" {
		companyDB = ac.SAPCompanyDB
	}
	return &legacy.Config{
		SAPServiceLayerURL:        ac.SAP.URL,
		SAPCompanyDB:              companyDB,
		SAPUsername:               ac.SAP.Username,
		SAPPassword:               ac.SAP.Password,
		SAPInsecureSkipSSL:        ac.SAP.InsecureSkipSSL,
		SAPMSSQLEnabled:           ac.MSSQL.Host != "",
		SAPMSSQLHost:              ac.MSSQL.Host,
		SAPMSSQLPort:              ac.MSSQL.Port,
		SAPMSSQLDatabase:          ac.MSSQL.Database,
		SAPMSSQLUser:              ac.MSSQL.User,
		SAPMSSQLPassword:          ac.MSSQL.Password,
		NembusOrganizationID:      int32(ac.Cloud.OrganizationID),
		NembusPriceListID:         ac.Sync.PriceListID,
		NembusTaxCategoryID:       ac.Sync.TaxCategoryID,
		NembusDefaultCustomer:     ac.Sync.DefaultCustomerCode,
		NembusDefaultWarehouse:    ac.Sync.DefaultWarehouseCode,
		BatchSize:                 ac.BatchSize,
		DryRun:                    ac.Sync.DryRun,
		DownstreamIntervalSeconds: ac.DownstreamIntervalSec,
		UpstreamOutboxPollMs:      ac.Sync.UpstreamOutboxPollMs,
	}
}

func NewClientFromAgent(ac *agentcfg.AgentConfig) (*Client, error) {
	return NewClient(LegacyConfig(ac))
}
