CREATE INDEX IF NOT EXISTS idx_transaction_events_user_id_created_at 
ON transaction_events(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transaction_events_user_id_type_created_at 
ON transaction_events(user_id, transaction_type, created_at DESC);

DROP INDEX IF EXISTS idx_transaction_events_user_id;

