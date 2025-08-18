CREATE TABLE IF NOT EXISTS transaction_drafts (
    id TEXT PRIMARY KEY,
    telegram_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    currency TEXT,
    description TEXT,
    category_id TEXT,
    occurred_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


