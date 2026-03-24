CREATE TABLE IF NOT EXISTS endpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uri TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    content TEXT NULL,

    CHECK (method IN ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD')),
    CHECK (json_valid(content)),
    UNIQUE(uri, method)
);

CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint_id INTEGER NOT NULL,
    uri TEXT NOT NULL,
    hit_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    request_headers TEXT,
    request_body TEXT,

    FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE,
    CHECK (request_headers IS NULL OR json_valid(request_headers))
    CHECK (request_body IS NULL OR json_valid(request_body))
);

CREATE TABLE IF NOT EXISTS assertions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    asserted_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE
    CHECK (status IN ('ok', 'failed')),
    UNIQUE(request_id)
);
