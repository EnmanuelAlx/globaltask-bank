-- ==========================================
-- GlobalTask Bank - RLS Policies
-- ==========================================

-- Helper function to check if the current user is an admin
-- SECURITY DEFINER allows this to run with higher privileges to read profiles
CREATE OR REPLACE FUNCTION public.is_admin()
RETURNS BOOLEAN AS $$
BEGIN
  RETURN (
    EXISTS (
      SELECT 1 FROM public.profiles 
      WHERE id = auth.uid() AND role = 'ADMIN'
    )
  );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 1. LOAN APPLICATIONS
ALTER TABLE public.loan_applications ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can manage all applications" ON public.loan_applications 
    FOR ALL TO authenticated USING (public.is_admin());

CREATE POLICY "Users can see own applications" ON public.loan_applications 
    FOR SELECT TO authenticated USING (user_id = auth.uid());

-- 2. PROFILES
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can manage all profiles" ON public.profiles 
    FOR ALL TO authenticated USING (public.is_admin());

CREATE POLICY "Users can see own profile" ON public.profiles 
    FOR SELECT TO authenticated USING (id = auth.uid());

-- 3. COUNTRIES (Listable for authenticated users)
ALTER TABLE public.countries ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can manage countries" ON public.countries 
    FOR ALL TO authenticated USING (public.is_admin());

CREATE POLICY "Anyone authenticated can list countries" ON public.countries 
    FOR SELECT TO authenticated USING (true);

-- 4. BANK PROVIDERS
ALTER TABLE public.bank_providers ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can manage bank providers" ON public.bank_providers 
    FOR ALL TO authenticated USING (public.is_admin());

-- 5. EVENT OUTBOX (Strictly Admin only)
ALTER TABLE public.event_outbox ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Admins can manage event outbox" ON public.event_outbox 
    FOR ALL TO authenticated USING (public.is_admin());
