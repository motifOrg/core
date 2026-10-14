ALTER TABLE companies DROP CONSTRAINT companies_source_check;

ALTER TABLE companies ADD CONSTRAINT companies_source_check
    CHECK (source IN ('sample', 'csv', 'manual', 'inbound', 'gmail'));
