CREATE TABLE IF NOT EXISTS batch_riders (
    batch_id TEXT PRIMARY KEY REFERENCES batches(id) ON DELETE CASCADE,
    rider_id TEXT NOT NULL REFERENCES riders(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ
)