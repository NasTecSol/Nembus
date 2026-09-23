-- =====================================================================
-- Migration: Seed Standard Currencies and Relax Business Partner Currency FK
-- =====================================================================

INSERT INTO currencies (code, name, symbol, decimal_places, is_active)
VALUES 
    ('SAR', 'Saudi Riyal', 'ر.س', 2, true),
    ('USD', 'US Dollar', '$', 2, true),
    ('EUR', 'Euro', '€', 2, true),
    ('AED', 'UAE Dirham', 'د.إ', 2, true),
    ('GBP', 'British Pound', '£', 2, true),
    ('KWD', 'Kuwaiti Dinar', 'د.ك', 3, true),
    ('BHD', 'Bahraini Dinar', 'ب.د', 3, true),
    ('OMR', 'Omani Rial', 'ر.ع', 3, true),
    ('QAR', 'Qatari Riyal', 'ر.ق', 2, true)
ON CONFLICT(code) DO NOTHING;

-- Make currency_code nullable or foreign key with ON UPDATE CASCADE ON DELETE SET NULL
ALTER TABLE business_partners DROP CONSTRAINT IF EXISTS business_partners_currency_code_fkey;
ALTER TABLE business_partners ADD CONSTRAINT business_partners_currency_code_fkey 
    FOREIGN KEY (currency_code) REFERENCES currencies(code) ON UPDATE CASCADE ON DELETE SET NULL;
