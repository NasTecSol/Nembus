package schema

// SAP Business One Table Constants
const (
	TableOWHS = "OWHS" // Warehouses / Stores
	TableOBIN = "OBIN" // Storage Locations / Bins
	TableOUSR = "OUSR" // Users
	TableOSLP = "OSLP" // Sales Employees / Cashiers
	TableOUOM = "OUOM" // Units of Measure
	TableOUGP = "OUGP" // UoM Groups
	TableUGP1 = "UGP1" // UoM Group Definitions
	TableOITB = "OITB" // Item Groups / Categories
	TableOMRC = "OMRC" // Manufacturers / Brands
	TableOITM = "OITM" // Item Master Data / Products
	TableOBCD = "OBCD" // Bar Codes
	TableITM1 = "ITM1" // Item Price Lists
	TableOPLN = "OPLN" // Price Lists
	TableOITW = "OITW" // Item Warehouse Info / Inventory Stock
	TableOCRD = "OCRD" // Business Partners (Customers / Suppliers)
	TableCRD1 = "CRD1" // BP Addresses
	TableORDR = "ORDR" // Sales Orders Header
	TableRDR1 = "RDR1" // Sales Orders Lines
	TableOINV = "OINV" // A/R Invoices Header
	TableINV1 = "INV1" // A/R Invoices Lines
	TableOADM = "OADM" // Company Administration / Info
)

// SQL Queries for SAP Business One Inspection and Extraction

const QuerySAPVersion = `
SELECT 
    ISNULL(CompnyName, '') AS CompnyName,
    ISNULL(Version, '') AS Version,
    ISNULL(RevNum, '') AS RevNum,
    ISNULL(CompnyAddr, '') AS CompnyAddr
FROM OADM;
`

const QueryTableCounts = `
SELECT 
    'OWHS' AS TableName, COUNT(1) AS TotalCount FROM OWHS
UNION ALL
SELECT 
    'OBIN' AS TableName, COUNT(1) AS TotalCount FROM OBIN
UNION ALL
SELECT 
    'OUSR' AS TableName, COUNT(1) AS TotalCount FROM OUSR
UNION ALL
SELECT 
    'OSLP' AS TableName, COUNT(1) AS TotalCount FROM OSLP
UNION ALL
SELECT 
    'OUOM' AS TableName, COUNT(1) AS TotalCount FROM OUOM
UNION ALL
SELECT 
    'OUGP' AS TableName, COUNT(1) AS TotalCount FROM OUGP
UNION ALL
SELECT 
    'OITB' AS TableName, COUNT(1) AS TotalCount FROM OITB
UNION ALL
SELECT 
    'OMRC' AS TableName, COUNT(1) AS TotalCount FROM OMRC
UNION ALL
SELECT 
    'OITM' AS TableName, COUNT(1) AS TotalCount FROM OITM
UNION ALL
SELECT 
    'OBCD' AS TableName, COUNT(1) AS TotalCount FROM OBCD
UNION ALL
SELECT 
    'OITW' AS TableName, COUNT(1) AS TotalCount FROM OITW
UNION ALL
SELECT 
    'OCRD_C' AS TableName, COUNT(1) AS TotalCount FROM OCRD WHERE CardType = 'C'
UNION ALL
SELECT 
    'OCRD_S' AS TableName, COUNT(1) AS TotalCount FROM OCRD WHERE CardType = 'S'
UNION ALL
SELECT 
    'ORDR' AS TableName, COUNT(1) AS TotalCount FROM ORDR
UNION ALL
SELECT 
    'OINV' AS TableName, COUNT(1) AS TotalCount FROM OINV;
`

const QueryStores = `
SELECT 
    WhsCode,
    ISNULL(WhsName, '') AS WhsName,
    ISNULL(Locked, 'N') AS Locked,
    ISNULL(Street, '') AS Street,
    ISNULL(City, '') AS City,
    ISNULL(Country, '') AS Country,
    ISNULL(ZipCode, '') AS ZipCode
FROM OWHS
ORDER BY WhsCode;
`

const QueryStorageLocations = `
SELECT 
    AbsEntry,
    BinCode,
    WhsCode,
    ISNULL(Descr, '') AS Descr,
    ISNULL(Disabled, 'N') AS Disabled
FROM OBIN
ORDER BY AbsEntry;
`

const QueryUsers = `
SELECT 
    USERID,
    ISNULL(USER_CODE, '') AS USER_CODE,
    ISNULL(U_NAME, '') AS U_NAME,
    ISNULL(E_Mail, '') AS E_Mail,
    ISNULL(Locked, 'N') AS Locked
FROM OUSR
ORDER BY USERID;
`

const QueryCashiers = `
SELECT 
    SlpCode,
    ISNULL(SlpName, '') AS SlpName,
    ISNULL(Memo, '') AS Memo,
    ISNULL(Active, 'Y') AS Active,
    ISNULL(Email, '') AS Email,
    ISNULL(Telephone, '') AS Telephone
FROM OSLP
ORDER BY SlpCode;
`

const QueryUOM = `
WITH AllUoms AS (
    SELECT 
        UomEntry,
        LTRIM(RTRIM(UomCode)) AS UomCode,
        ISNULL(UomName, UomCode) AS UomName,
        ISNULL(Locked, 'N') AS Locked
    FROM OUOM
    WHERE UomCode IS NOT NULL AND LTRIM(RTRIM(UomCode)) <> ''
    UNION
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(InvntryUom)) AS UomCode,
        LTRIM(RTRIM(InvntryUom)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE InvntryUom IS NOT NULL AND LTRIM(RTRIM(InvntryUom)) <> ''
    UNION
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(SalUnitMsr)) AS UomCode,
        LTRIM(RTRIM(SalUnitMsr)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE SalUnitMsr IS NOT NULL AND LTRIM(RTRIM(SalUnitMsr)) <> ''
    UNION
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(BuyUnitMsr)) AS UomCode,
        LTRIM(RTRIM(BuyUnitMsr)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE BuyUnitMsr IS NOT NULL AND LTRIM(RTRIM(BuyUnitMsr)) <> ''
)
SELECT 
    UomEntry,
    UomCode,
    UomName,
    Locked
FROM (
    SELECT 
        UomEntry,
        UomCode,
        UomName,
        Locked,
        ROW_NUMBER() OVER (PARTITION BY UomCode ORDER BY CASE WHEN UomEntry > 0 THEN 0 ELSE 1 END) AS rn
    FROM AllUoms
) t
WHERE rn = 1
ORDER BY UomCode;
`

const QueryUOMFallback = `
WITH ItemUoms AS (
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(InvntryUom)) AS UomCode,
        LTRIM(RTRIM(InvntryUom)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE InvntryUom IS NOT NULL AND LTRIM(RTRIM(InvntryUom)) <> ''
    UNION
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(SalUnitMsr)) AS UomCode,
        LTRIM(RTRIM(SalUnitMsr)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE SalUnitMsr IS NOT NULL AND LTRIM(RTRIM(SalUnitMsr)) <> ''
    UNION
    SELECT 
        -1 AS UomEntry,
        LTRIM(RTRIM(BuyUnitMsr)) AS UomCode,
        LTRIM(RTRIM(BuyUnitMsr)) AS UomName,
        'N' AS Locked
    FROM OITM
    WHERE BuyUnitMsr IS NOT NULL AND LTRIM(RTRIM(BuyUnitMsr)) <> ''
)
SELECT DISTINCT
    UomEntry,
    UomCode,
    UomName,
    Locked
FROM ItemUoms
WHERE UomCode <> ''
ORDER BY UomCode;
`

const QueryUOMGroups = `
SELECT 
    g.UgpEntry,
    g.UgpCode,
    ISNULL(g.UgpName, g.UgpCode) AS UgpName,
    g.BaseUom AS BaseUomEntry,
    bu.UomCode AS BaseUomCode,
    ISNULL(d.UomEntry, g.BaseUom) AS AltUomEntry,
    ISNULL(au.UomCode, bu.UomCode) AS AltUomCode,
    ISNULL(d.AltQty, 1.0) AS AltQty,
    ISNULL(d.BaseQty, 1.0) AS BaseQty
FROM OUGP g
INNER JOIN OUOM bu ON g.BaseUom = bu.UomEntry
LEFT JOIN UGP1 d ON g.UgpEntry = d.UgpEntry
LEFT JOIN OUOM au ON d.UomEntry = au.UomEntry
ORDER BY g.UgpEntry, d.LineNum;
`

const QueryCategories = `
SELECT 
    ItmsGrpCod,
    ISNULL(ItmsGrpNam, '') AS ItmsGrpNam
FROM OITB
ORDER BY ItmsGrpCod;
`

const QueryBrands = `
SELECT 
    FirmCode,
    ISNULL(FirmName, '') AS FirmName
FROM OMRC
ORDER BY FirmCode;
`

const QueryProducts = `
SELECT 
    ItemCode,
    ISNULL(ItemName, '') AS ItemName,
    ISNULL(UserText, '') AS UserText,
    ISNULL(OITM.ItmsGrpCod, 0) AS ItmsGrpCod,
    ISNULL(g.ItmsGrpNam, '') AS ItmsGrpNam,
    ISNULL(FirmCode, 0) AS FirmCode,
    ISNULL(InvntItem, 'Y') AS InvntItem,
    ISNULL(SellItem, 'Y') AS SellItem,
    ISNULL(PrchseItem, 'Y') AS PrchseItem,
    ISNULL(validFor, 'Y') AS validFor,
    ISNULL(CodeBars, '') AS CodeBars,
    ISNULL(BuyUnitMsr, 'UNIT') AS BuyUnitMsr,
    ISNULL(SalUnitMsr, 'UNIT') AS SalUnitMsr,
    ISNULL(InvntryUom, 'UNIT') AS InvntryUom,
    ISNULL(NumInSale, 1.0) AS NumInSale,
    ISNULL(NumInBuy, 1.0) AS NumInBuy,
    ISNULL(UgpEntry, -1) AS UgpEntry,
    ISNULL(OITM.IUoMEntry, -1) AS IUoMEntry,
    ISNULL(OITM.SUoMEntry, -1) AS SUoMEntry,
    ISNULL(OITM.PUoMEntry, -1) AS PUoMEntry,
    ISNULL(ManSerNum, 'N') AS ManSerNum,
    ISNULL(ManBtchNum, 'N') AS ManBtchNum,
    ISNULL(VatGourpSa, '') AS VatGourpSa
FROM OITM
LEFT JOIN OITB g ON OITM.ItmsGrpCod = g.ItmsGrpCod
ORDER BY ItemCode;
`

const QueryProductBarcodes = `
SELECT 
    b.BcdEntry,
    b.BcdCode,
    b.ItemCode,
    ISNULL(b.UomEntry, -1) AS UomEntry,
    ISNULL(u.UomCode, '') AS UomCode
FROM OBCD b
LEFT JOIN OUOM u ON b.UomEntry = u.UomEntry
ORDER BY b.BcdEntry;
`

const QueryInventoryStock = `
SELECT 
    ItemCode,
    WhsCode,
    ISNULL(OnHand, 0.0) AS OnHand,
    ISNULL(IsCommited, 0.0) AS IsCommited,
    ISNULL(OnOrder, 0.0) AS OnOrder,
    ISNULL(MinStock, 0.0) AS MinStock,
    ISNULL(MaxStock, 0.0) AS MaxStock
FROM OITW
WHERE OnHand <> 0 OR IsCommited <> 0 OR OnOrder <> 0
ORDER BY ItemCode, WhsCode;
`

const QueryBusinessPartners = `
SELECT 
    CardCode,
    ISNULL(CardName, '') AS CardName,
    ISNULL(CardType, 'C') AS CardType,
    ISNULL(LicTradNum, '') AS LicTradNum,
    ISNULL(Phone1, '') AS Phone1,
    ISNULL(E_Mail, '') AS E_Mail,
    ISNULL(Currency, 'USD') AS Currency,
    ISNULL(validFor, 'Y') AS validFor,
    ISNULL(Balance, 0.0) AS Balance
FROM OCRD
ORDER BY CardCode;
`

const QuerySalesOrdersHeader = `
SELECT 
    DocEntry,
    DocNum,
    DocDate,
    DocDueDate,
    ISNULL(CardCode, '') AS CardCode,
    ISNULL(CardName, '') AS CardName,
    ISNULL(DocTotal, 0.0) AS DocTotal,
    ISNULL(VatSum, 0.0) AS VatSum,
    ISNULL(DiscSum, 0.0) AS DiscSum,
    ISNULL(DocStatus, 'O') AS DocStatus,
    ISNULL(SlpCode, -1) AS SlpCode,
    ISNULL(Comments, '') AS Comments
FROM ORDR
WHERE DocDate >= @FromDate AND DocDate <= @ToDate
ORDER BY DocEntry;
`

const QuerySalesOrderLines = `
SELECT 
    DocEntry,
    LineNum,
    ISNULL(ItemCode, '') AS ItemCode,
    ISNULL(Dscription, '') AS Dscription,
    ISNULL(Quantity, 0.0) AS Quantity,
    ISNULL(Price, 0.0) AS Price,
    ISNULL(LineTotal, 0.0) AS LineTotal,
    ISNULL(VatSum, 0.0) AS VatSum,
    ISNULL(WhsCode, '') AS WhsCode,
    ISNULL(unitMsr, 'UNIT') AS unitMsr
FROM RDR1
WHERE DocEntry IN (%s)
ORDER BY DocEntry, LineNum;
`

const QuerySalesOrderLinesByDate = `
SELECT 
    l.DocEntry,
    l.LineNum,
    ISNULL(l.ItemCode, '') AS ItemCode,
    ISNULL(l.Dscription, '') AS Dscription,
    ISNULL(l.Quantity, 0.0) AS Quantity,
    ISNULL(l.Price, 0.0) AS Price,
    ISNULL(l.LineTotal, 0.0) AS LineTotal,
    ISNULL(l.VatSum, 0.0) AS VatSum,
    ISNULL(l.WhsCode, '') AS WhsCode,
    ISNULL(l.unitMsr, 'UNIT') AS unitMsr
FROM RDR1 l
INNER JOIN ORDR h ON l.DocEntry = h.DocEntry
WHERE h.DocDate >= @FromDate AND h.DocDate <= @ToDate
ORDER BY l.DocEntry, l.LineNum;
`

const QueryInvoicesHeader = `
SELECT 
    DocEntry,
    DocNum,
    DocDate,
    DocDueDate,
    ISNULL(CardCode, '') AS CardCode,
    ISNULL(CardName, '') AS CardName,
    ISNULL(DocTotal, 0.0) AS DocTotal,
    ISNULL(PaidToDate, 0.0) AS PaidToDate,
    ISNULL(VatSum, 0.0) AS VatSum,
    ISNULL(DiscSum, 0.0) AS DiscSum,
    ISNULL(DocStatus, 'C') AS DocStatus,
    ISNULL(SlpCode, -1) AS SlpCode,
    ISNULL(Comments, '') AS Comments
FROM OINV
WHERE DocDate >= @FromDate AND DocDate <= @ToDate
ORDER BY DocEntry;
`

const QueryInvoiceLines = `
SELECT 
    DocEntry,
    LineNum,
    ISNULL(ItemCode, '') AS ItemCode,
    ISNULL(Dscription, '') AS Dscription,
    ISNULL(Quantity, 0.0) AS Quantity,
    ISNULL(Price, 0.0) AS Price,
    ISNULL(LineTotal, 0.0) AS LineTotal,
    ISNULL(VatSum, 0.0) AS VatSum,
    ISNULL(WhsCode, '') AS WhsCode,
    ISNULL(unitMsr, 'UNIT') AS unitMsr
FROM INV1
WHERE DocEntry IN (%s)
ORDER BY DocEntry, LineNum;
`

const QueryInvoiceLinesByDate = `
SELECT 
    l.DocEntry,
    l.LineNum,
    ISNULL(l.ItemCode, '') AS ItemCode,
    ISNULL(l.Dscription, '') AS Dscription,
    ISNULL(l.Quantity, 0.0) AS Quantity,
    ISNULL(l.Price, 0.0) AS Price,
    ISNULL(l.LineTotal, 0.0) AS LineTotal,
    ISNULL(l.VatSum, 0.0) AS VatSum,
    ISNULL(l.WhsCode, '') AS WhsCode,
    ISNULL(l.unitMsr, 'UNIT') AS unitMsr
FROM INV1 l
INNER JOIN OINV h ON l.DocEntry = h.DocEntry
WHERE h.DocDate >= @FromDate AND h.DocDate <= @ToDate
ORDER BY l.DocEntry, l.LineNum;
`


const QueryPriceLists = `
SELECT
    ListNum,
    ISNULL(ListName, '') AS ListName,
    ISNULL(Currency, 'USD') AS Currency,
    ISNULL(Factor, 1.0) AS Factor,
    ISNULL(BasedOn, 0) AS BasedOn,
    ISNULL(validFor, 'Y') AS validFor
FROM OPLN
ORDER BY ListNum;
`

const QueryPriceListItems = `
SELECT
    i.ItemCode,
    i.PriceList,
    ISNULL(i.Price, 0.0) AS Price,
    ISNULL(i.Currency, '') AS Currency,
    ISNULL(i.UomEntry, -1) AS UomEntry,
    ISNULL(u.UomCode, '') AS UomCode
FROM ITM1 i
LEFT JOIN OUOM u ON i.UomEntry = u.UomEntry
WHERE i.Price > 0
ORDER BY i.ItemCode, i.PriceList;
`

const QueryBPAddresses = `
SELECT
    CardCode,
    AdresType,
    ISNULL(Address, '') AS Address,
    ISNULL(Street, '') AS Street,
    ISNULL(City, '') AS City,
    ISNULL(Country, '') AS Country,
    ISNULL(ZipCode, '') AS ZipCode,
    ISNULL(State1, '') AS State1,
    ISNULL(Phone1, '') AS Phone1,
    ISNULL(Phone2, '') AS Phone2
FROM CRD1
ORDER BY CardCode, AdresType;
`
