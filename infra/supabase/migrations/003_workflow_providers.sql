-- Migration: 003_workflow_providers.sql
-- Description: Create workflow_providers table to decouple workflow steps from specific provider configurations.

CREATE TABLE IF NOT EXISTS public.workflow_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_name VARCHAR(10) NOT NULL, -- e.g., 'PT', 'CO'
    provider_id INTEGER NOT NULL REFERENCES public.bank_providers(id),
    event_step VARCHAR(100) NOT NULL, -- e.g., 'FetchBankData'
    endpoint_path VARCHAR(255) NOT NULL, -- e.g., '/validate'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trigger to update updated_at
CREATE TRIGGER update_workflow_providers_updated_at
    BEFORE UPDATE ON workflow_providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Index for lookups
CREATE INDEX IF NOT EXISTS idx_workflow_providers_lookup
    ON workflow_providers(workflow_name, event_step, is_active);

-- Seed initial mappings
INSERT INTO public.workflow_providers (workflow_name, provider_id, event_step, endpoint_path)
VALUES
    ('PT', (SELECT id FROM bank_providers WHERE provider_name = 'Santander Totta'), 'FETCH_BANK_DATA', '/validate'),
    ('PT', (SELECT id FROM bank_providers WHERE provider_name = 'Millennium BCP'), 'FETCH_BANK_DATA', '/validate'),
    ('PT', (SELECT id FROM bank_providers WHERE provider_name = 'Santander Totta'), 'VALIDATE_USER_IDENTITY', '/fetch-user-data'),
    ('CO', (SELECT id FROM bank_providers WHERE provider_name = 'Bancolombia'), 'FETCH_BANK_DATA', '/validate')
ON CONFLICT DO NOTHING;

CREATE POLICY "Admins can manage workflow providers" ON public.workflow_providers
    FOR ALL TO authenticated USING (public.is_admin());