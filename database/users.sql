CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    topics_opened INTEGER DEFAULT 0,
    messages_sent INTEGER DEFAULT 0,
    creation_date DATETIME DEFAULT CURRENT_TIMESTAMP
);