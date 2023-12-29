-- ChatGPT generated

CREATE TABLE IF NOT EXISTS "torrent" (
    "torrent_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "hash_info" BLOB NOT NULL,
    "created_time" DATETIME NOT NULL,
    "paused" BOOLEAN NOT NULL,
    "location" TEXT,
    "progress" INTEGER,
    "raw_meta_info" BLOB
);

CREATE TABLE IF NOT EXISTS "tracker_announce" (
    "tracker_announce_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "torrent_id" INTEGER NOT NULL,
    "announce_time" DATETIME NOT NULL,
    "announciation" TEXT NOT NULL,
    "scheduled_time" DATETIME,
    "err" TEXT,
    "done" BOOLEAN,
    "raw_response" BLOB,
    FOREIGN KEY ("torrent_id") REFERENCES "torrent" ("torrent_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "peer" (
    "peer_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "protocol_peer_id" BLOB NOT NULL,
    "ip" TEXT NOT NULL,
    "port" INTEGER NOT NULL,
    "torrent_id" INTEGER NOT NULL,
    "reachable" BOOLEAN NOT NULL,
    FOREIGN KEY ("torrent_id") REFERENCES "torrent" ("torrent_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "piece" (
    "piece_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "torrent_id" INTEGER NOT NULL,
    "peer_id" INTEGER NOT NULL,
    "is_downloaded" BOOLEAN NOT NULL,
    "start" DATETIME NOT NULL,
    "end" DATETIME,
    "index" INTEGER NOT NULL,
    "length" INTEGER NOT NULL,
    "confirmed" BOOLEAN,
    FOREIGN KEY ("torrent_id") REFERENCES "torrent" ("torrent_id") ON DELETE CASCADE,
    FOREIGN KEY ("peer_id") REFERENCES "peer" ("peer_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "clients" (
    "client_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "protocol_id" BLOB,
    "created" DATETIME
);
