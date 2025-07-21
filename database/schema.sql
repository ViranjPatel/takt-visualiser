-- Zone Breakdown Structure
CREATE TABLE zones (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    parent_id INTEGER REFERENCES zones(id),
    name VARCHAR(255) NOT NULL,
    level INTEGER NOT NULL,
    path VARCHAR(1000) NOT NULL, -- Materialized path for fast queries
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tasks (Activities)
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    zone_id INTEGER REFERENCES zones(id),
    name VARCHAR(500) NOT NULL,
    start_date DATE NOT NULL,
    duration INTEGER NOT NULL, -- in days
    trade_id INTEGER,
    status VARCHAR(50) DEFAULT 'planned',
    sequence_number INTEGER,
    color VARCHAR(7), -- Hex color
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Trades (for color coding)
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(7) NOT NULL,
    project_id INTEGER NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_tasks_project_zone_date ON tasks(project_id, zone_id, start_date);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_zones_path ON zones(path);
CREATE INDEX idx_zones_project ON zones(project_id);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_zones_updated_at BEFORE UPDATE ON zones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
