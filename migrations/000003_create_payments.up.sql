CREATE TABLE payments (
    id             uuid PRIMARY KEY,
    household_id   uuid NOT NULL REFERENCES households (id) ON DELETE RESTRICT,
    waste_id       uuid NOT NULL REFERENCES waste_pickups (id) ON DELETE RESTRICT,
    amount         numeric(12,2) NOT NULL CHECK (amount > 0),
    payment_date   timestamptz,
    status         text NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'paid', 'failed')),
    proof_file_url text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_household_status ON payments (household_id, status);
CREATE INDEX idx_payments_waste ON payments (waste_id);
CREATE INDEX idx_payments_created ON payments (created_at);
