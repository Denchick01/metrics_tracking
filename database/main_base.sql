DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id INT PRIMARY KEY,
    name varchar(255),
    email varchar(255) NOT NULL
);

DROP TABLE IF EXISTS devices CASCADE;
CREATE TABLE devices (
    id INT PRIMARY KEY,
    name varchar(255) NOT NULL,
    user_id INT NOT NULL,
    CONSTRAINT devices_user_id_fk FOREIGN KEY(user_id) REFERENCES users (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS device_metrics;
CREATE TABLE device_metrics (
    id INT PRIMARY KEY,
    device_id INT NOT NULL,
    metric_1 INT,
    metric_2 INT,
    metric_3 INT, 
    metric_4 INT,
    metric_5 INT,
    local_time TIMESTAMP, 
    server_time TIMESTAMP DEFAULT NOW(), 
    CONSTRAINT device_metrics_device_id_fk FOREIGN KEY (device_id) REFERENCES devices (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS device_alerts CASCADE;
CREATE TABLE device_alerts (
    id INT PRIMARY KEY,
    device_id INT,
    message TEXT
);
