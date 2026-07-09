CREATE TABLE households (
    id         uuid PRIMARY KEY,
    owner_name text NOT NULL,
    address    text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
