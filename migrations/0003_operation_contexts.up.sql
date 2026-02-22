CREATE TABLE IF NOT EXISTS operation_contexts (
    op_id TEXT PRIMARY KEY,
    telegram_id INTEGER NOT NULL,
    tenant_id TEXT NOT NULL,
    transaction_id TEXT,
    description_original TEXT NOT NULL,
    category_id_selected TEXT,
    category_name_selected TEXT,
    selection_source TEXT NOT NULL,
    tx_type TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    currency TEXT NOT NULL,
    occurred_at TIMESTAMP,
    category_list_message_id INTEGER,
    confirmation_message_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
