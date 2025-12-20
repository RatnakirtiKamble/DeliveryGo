CREATE TABLE IF NOT EXISTS path_templates (
    id TEXT PRIMARY KEY,
    store_id TEXT NOT NULL,
    h3_cell TEXT NOT NULL,
    base_eta INT NOT NULL,
    polyline JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uniq_store_h3_path
ON path_templates(store_id, h3_cell);

CREATE TABLE IF NOT EXISTS batch_paths (
    batch_id TEXT PRIMARY KEY REFERENCES batches(id) ON DELETE CASCADE,
    path_id TEXT NOT NULL REFERENCES path_templates(id),
    assigned_at TIMESTAMP NOT NULL DEFAULT now()
);
