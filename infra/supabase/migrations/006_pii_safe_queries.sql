-- Phase 1: Foundation - PII Safe Queries (AES-256-GCM + Blind Index)

-- 1. Ensure identity_document is TEXT to accommodate base64 ciphertext
ALTER TABLE public.profiles ALTER COLUMN identity_document TYPE TEXT;

-- 2. Add identity_document_bidx for exact-match searches
ALTER TABLE public.profiles ADD COLUMN identity_document_bidx TEXT;

-- 3. Create index for fast searches
CREATE INDEX idx_profiles_identity_document_bidx ON public.profiles (identity_document_bidx);

-- 4. Update unique constraint to use blind index
-- First, drop the old one (naming convention from Postgres for UNIQUE (a, b) is table_col1_col2_key)
ALTER TABLE public.profiles DROP CONSTRAINT IF EXISTS profiles_identity_document_country_id_key;
-- Then add the new one
ALTER TABLE public.profiles ADD CONSTRAINT profiles_identity_document_bidx_country_id_key UNIQUE (identity_document_bidx, country_id);

-- 5. Update handle_new_user trigger function to populate identity_document_bidx
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, full_name, identity_document, identity_document_bidx, country_id, role)
    VALUES (
        new.id,
        COALESCE(new.raw_user_meta_data->>'full_name', 'User'),
        new.raw_user_meta_data->>'identity_document',
        new.raw_user_meta_data->>'identity_document_bidx',
        (new.raw_user_meta_data->>'country_id')::INTEGER,
        COALESCE(new.raw_user_meta_data->>'role', 
            CASE
                WHEN new.email LIKE '%@globaltask%' THEN 'ADMIN'
                ELSE 'USER'
            END
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
