-- Migration: 004_refactor_bank_providers.sql
-- Description: Add base_url to bank_providers and migrate data from api_config

ALTER TABLE public.bank_providers ADD COLUMN IF NOT EXISTS base_url VARCHAR(255);

-- Update base_url from api_config url (assuming url is like http://host:port/path)
-- We extract the part before the last slash if it contains /validate
UPDATE public.bank_providers 
SET base_url = 'http://mock-bank:8081'
WHERE api_config->>'url' LIKE '%mock-bank%';

-- If there are other providers, we'd need more logic, but for now this is enough for the mock.
UPDATE public.bank_providers 
SET base_url = 'http://mock-bank:8081'
WHERE base_url IS NULL;

-- Make it NOT NULL for future
ALTER TABLE public.bank_providers ALTER COLUMN base_url SET NOT NULL;
