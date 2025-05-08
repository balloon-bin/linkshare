CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    kind TEXT CHECK(kind IN ('int', 'string', 'bool', 'json', 'glob')) NOT NULL
);

INSERT INTO settings (key, value, kind) VALUES ('schema-version', '1', 'int');

CREATE TABLE links (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT,
    is_private BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX idx_links_created_at ON links(created_at);
CREATE INDEX idx_links_url ON links(url);
