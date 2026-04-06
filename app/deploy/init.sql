CREATE TABLE IF NOT EXISTS contract_writes (
  id BIGSERIAL PRIMARY KEY,
  request_id UUID NOT NULL UNIQUE,
  tx_hash TEXT NOT NULL UNIQUE,
  value TEXT NOT NULL,
  from_address TEXT NOT NULL,
  contract_address TEXT NOT NULL,
  status TEXT NOT NULL,
  block_number BIGINT NULL,
  error_message TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contract_writes_status ON contract_writes(status);
CREATE INDEX IF NOT EXISTS idx_contract_writes_tx_hash ON contract_writes(tx_hash);
CREATE INDEX IF NOT EXISTS idx_contract_writes_created_at ON contract_writes(created_at);
CREATE INDEX IF NOT EXISTS idx_contract_writes_status_created_id ON contract_writes(status, created_at, id);

CREATE TABLE IF NOT EXISTS contract_state_projection (
  contract_address TEXT PRIMARY KEY,
  current_value TEXT NOT NULL,
  last_block_number BIGINT NOT NULL,
  last_tx_hash TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contract_state_projection_block ON contract_state_projection(last_block_number);
