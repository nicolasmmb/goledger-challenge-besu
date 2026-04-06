CREATE TABLE IF NOT EXISTS contract_state_projection (
  contract_address TEXT PRIMARY KEY,
  current_value TEXT NOT NULL,
  last_block_number BIGINT NOT NULL,
  last_tx_hash TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contract_state_projection_block ON contract_state_projection(last_block_number);
