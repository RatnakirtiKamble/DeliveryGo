CREATE TABLE IF NOT EXISTS batches (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS batch_orders (
  batch_id TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
  order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  PRIMARY KEY (batch_id, order_id)
);

CREATE INDEX IF NOT EXISTS idx_batches_status
ON batches (status);