CREATE TABLE IF NOT EXISTS path_templates (
    id SERIAL PRIMARY KEY,
    store_id TEXT NOT NULL,
    base_eta INT NOT NULL,
);

CREATE TABLE IF NOT EXISTS batch_paths (
    batch_id TEXT PRIMARY KEY REFERENCES batches(id) ON DELETE CASCADE,
    path_id TEXT NOT NULL REFERENCES path_templates(id) 
)

