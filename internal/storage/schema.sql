-- messages 
CREATE TABLES messages {
    id INTEGER PRIMARY KEY AUTOINCREMENT
    room_id TEXT,
    sender_id TEXT,
    sender TEXT, 
    content TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP 
};

-- users 
CREAT TABLES users {
    id TEXT PRIMARY KEY,
    username TEXT,
    created_at DATETIME,
    last_active DATETIME
};