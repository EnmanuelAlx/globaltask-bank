-- Migration: 005_add_priority_to_workflow_providers.sql
-- Description: Add priority column to workflow_providers and seed initial priorities.

ALTER TABLE public.workflow_providers
ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;

-- Seed initial priorities for Portugal
UPDATE public.workflow_providers
SET priority = 1
WHERE workflow_name = 'PT'
  AND event_step = 'FETCH_BANK_DATA'
  AND provider_id = (SELECT id FROM bank_providers WHERE provider_name = 'Santander Totta');

UPDATE public.workflow_providers
SET priority = 2
WHERE workflow_name = 'PT'
  AND event_step = 'FETCH_BANK_DATA'
  AND provider_id = (SELECT id FROM bank_providers WHERE provider_name = 'Millennium BCP');

-- Seed initial priority for identity validation
UPDATE public.workflow_providers
SET priority = 1
WHERE workflow_name = 'PT'
  AND event_step = 'VALIDATE_USER_IDENTITY'
  AND provider_id = (SELECT id FROM bank_providers WHERE provider_name = 'Santander Totta');

-- Seed initial priority for Colombia
UPDATE public.workflow_providers
SET priority = 1
WHERE workflow_name = 'CO'
  AND event_step = 'FETCH_BANK_DATA'
  AND provider_id = (SELECT id FROM bank_providers WHERE provider_name = 'Bancolombia');
