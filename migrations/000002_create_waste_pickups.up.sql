CREATE TABLE waste_pickups (
    id           uuid PRIMARY KEY,
    household_id uuid NOT NULL REFERENCES households (id) ON DELETE RESTRICT,
    type         text NOT NULL CHECK (type IN ('organic', 'plastic', 'paper', 'electronic')),
    status       text NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'scheduled', 'completed', 'canceled')),
    pickup_date  timestamptz,
    safety_check boolean,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_pickups_household ON waste_pickups (household_id);
CREATE INDEX idx_pickups_autocancel ON waste_pickups (type, status, created_at);
