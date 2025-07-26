
-- Warehouse Management System PostgreSQL Setup Script

-- 1. Database Creation
-- Connect to your PostgreSQL instance and run this command first if the database doesn't exist.
-- You might need to run this command separately from the rest of the script.
-- CREATE DATABASE warehouse_management;

-- After creating the database, connect to it to run the rest of the script.
-- \c warehouse_management

-- 2. Table Creation

-- Table for Materials
-- Stores information about each physical material in the warehouse.
CREATE TABLE IF NOT EXISTS materials (
    id VARCHAR(255) PRIMARY KEY,
    barcode VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- available, in_use, reserved, maintenance
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_materials_barcode ON materials(barcode);
CREATE INDEX IF NOT EXISTS idx_materials_status ON materials(status);

-- Table for Shelves and Slots
-- Stores the layout and status of each slot on every smart shelf.
CREATE TABLE IF NOT EXISTS slots (
    id VARCHAR(255) PRIMARY KEY,
    shelf_id VARCHAR(255) NOT NULL,
    "row" INT NOT NULL,
    "column" INT NOT NULL,
    status VARCHAR(50) NOT NULL, -- empty, occupied, reserved, maintenance
    material_id VARCHAR(255) REFERENCES materials(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1, -- For optimistic locking
    UNIQUE(shelf_id, "row", "column")
);
CREATE INDEX IF NOT EXISTS idx_slots_shelf_id ON slots(shelf_id);
CREATE INDEX IF NOT EXISTS idx_slots_status ON slots(status);
CREATE INDEX IF NOT EXISTS idx_slots_material_id ON slots(material_id);

-- Table for Operations
-- Records every operation (placement, removal, move) performed by operators or the system.
CREATE TABLE IF NOT EXISTS operations (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- placement, removal, move, reservation
    material_id VARCHAR(255) NOT NULL REFERENCES materials(id),
    slot_id VARCHAR(255) NOT NULL REFERENCES slots(id),
    operator_id VARCHAR(255) NOT NULL,
    shelf_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL -- pending, completed, failed, cancelled
);
CREATE INDEX IF NOT EXISTS idx_operations_material_id ON operations(material_id);
CREATE INDEX IF NOT EXISTS idx_operations_slot_id ON operations(slot_id);
CREATE INDEX IF NOT EXISTS idx_operations_operator_id ON operations(operator_id);
CREATE INDEX IF NOT EXISTS idx_operations_timestamp ON operations(timestamp DESC);

-- Table for Alerts
-- Stores system-generated alerts for issues like low stock, slot errors, etc.
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    shelf_id VARCHAR(255),
    slot_id VARCHAR(255),
    message TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL, -- low, medium, high, critical
    status VARCHAR(50) NOT NULL, -- active, acknowledged, resolved
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    metadata JSONB
);
CREATE INDEX IF NOT EXISTS idx_alerts_status_severity ON alerts(status, severity);
CREATE INDEX IF NOT EXISTS idx_alerts_shelf_id ON alerts(shelf_id);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- Table for Failed Events (Dead-Letter Queue)
-- Stores events that failed to be published to the message queue after several retries.
CREATE TABLE IF NOT EXISTS failed_events (
    id VARCHAR(255) PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    error TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT
);
CREATE INDEX IF NOT EXISTS idx_failed_events_resolved_created_at ON failed_events(resolved, created_at ASC);

-- Function to automatically update updated_at timestamps
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to update the updated_at column on every update
CREATE TRIGGER set_materials_timestamp
BEFORE UPDATE ON materials
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_slots_timestamp
BEFORE UPDATE ON slots
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_alerts_timestamp
BEFORE UPDATE ON alerts
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();


-- End of script
