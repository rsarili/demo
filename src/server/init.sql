CREATE TABLE IF NOT EXISTS devices (
    device_id VARCHAR(50) UNIQUE NOT NULL PRIMARY KEY,
    device_type VARCHAR(100) NOT NULL
);

INSERT INTO devices (device_id, device_type) VALUES
('sample-device', 'tv');