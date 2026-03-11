-- ==========================================
-- GlobalTask Bank - Business Schema
-- ==========================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Supabase schemas
CREATE SCHEMA IF NOT EXISTS auth;

-- Profiles table
CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID PRIMARY KEY,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'USER' CHECK (role IN ('ADMIN', 'USER')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Countries table
CREATE TABLE IF NOT EXISTS public.countries (
    id SERIAL PRIMARY KEY,
    iso_code VARCHAR(3) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    currency VARCHAR(10) NOT NULL
);

-- Bank providers table
CREATE TABLE IF NOT EXISTS public.bank_providers (
    id SERIAL PRIMARY KEY,
    country_id INTEGER NOT NULL REFERENCES countries(id),
    provider_name VARCHAR(255) NOT NULL,
    api_config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Loan applications table
CREATE TABLE IF NOT EXISTS public.loan_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES public.profiles(id),
    country_id INTEGER NOT NULL REFERENCES countries(id),
    borrower_name VARCHAR(255),
    identity_document VARCHAR(50) NOT NULL,
    requested_amount NUMERIC(15, 2) NOT NULL,
    monthly_income NUMERIC(15, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'PENDING_VALIDATION', 'AWAITING_BANK_DATA', 'ANALYZING_RISK', 'APPROVED', 'REJECTED')),
    bank_information JSONB DEFAULT '{}',
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_loan_applications_user_id ON loan_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_loan_applications_status ON loan_applications(status);

-- Event outbox
CREATE TABLE IF NOT EXISTS public.event_outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PROCESSING', 'DONE', 'FAILED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    locked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_event_outbox_status ON event_outbox(status) WHERE status = 'PENDING';

-- Seed data
INSERT INTO countries (id, iso_code, name, currency) VALUES
    (1, 'PT', 'Portugal', 'EUR'),
    (2, 'CO', 'Colombia', 'COP')
ON CONFLICT (iso_code) DO NOTHING;

-- Bank Providers for Portugal
INSERT INTO bank_providers (country_id, provider_name, api_config) VALUES
    (1, 'Santander Totta', '{"url": "http://mock-bank:8081/validate", "timeout": 10}'),
    (1, 'Millennium BCP', '{"url": "http://mock-bank:8081/validate", "timeout": 5}')
ON CONFLICT DO NOTHING;

-- Bank Providers for Colombia
INSERT INTO bank_providers (country_id, provider_name, api_config) VALUES
    (2, 'Bancolombia', '{"url": "http://mock-bank:8081/validate", "timeout": 8, "provides_income": true}')
ON CONFLICT DO NOTHING;

-- Utils
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_loan_applications_updated_at
    BEFORE UPDATE ON loan_applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Automatically create a profile when a new user signs up
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, full_name, role)
    VALUES (
        new.id,
        COALESCE(new.raw_user_meta_data->>'full_name', 'User'),
        CASE
            WHEN new.email LIKE '%@globaltask%' THEN 'ADMIN'
            ELSE 'USER'
        END
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Real-time notifications for loan applications
CREATE OR REPLACE FUNCTION public.notify_loan_application_update()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('loan_application_updates', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_loan_application_update
    AFTER INSERT OR UPDATE ON public.loan_applications
    FOR EACH ROW EXECUTE FUNCTION public.notify_loan_application_update();
