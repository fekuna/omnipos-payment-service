CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,
    order_id UUID NOT NULL UNIQUE,
    amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    payment_method INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 0,
    reference_number VARCHAR(100),
    provider VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payments_merchant_id ON payments(merchant_id);
CREATE INDEX idx_payments_order_id ON payments(order_id);
