-- 000008_finance_and_payments.up.sql

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL,
    hours NUMERIC(10, 2) NOT NULL,
    format VARCHAR(50) NOT NULL CHECK (format IN ('individual', 'pair', 'group')),
    payment_method VARCHAR(50) NOT NULL DEFAULT 'transfer' CHECK (payment_method IN ('transfer', 'cash', 'card', 'other')),
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_teacher_id ON payments (teacher_id);
CREATE INDEX IF NOT EXISTS idx_payments_client_id ON payments (client_id);
CREATE INDEX IF NOT EXISTS idx_payments_paid_at ON payments (paid_at);

CREATE TABLE IF NOT EXISTS partner_payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    period_month VARCHAR(7) NOT NULL, -- Формат YYYY-MM
    gross_amount NUMERIC(10, 2) NOT NULL,
    commission_amount NUMERIC(10, 2) NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_partner_payouts_teacher_id ON partner_payouts (teacher_id);
CREATE INDEX IF NOT EXISTS idx_partner_payouts_tag_id ON partner_payouts (tag_id);
CREATE INDEX IF NOT EXISTS idx_partner_payouts_period_month ON partner_payouts (period_month);
