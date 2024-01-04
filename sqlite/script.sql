-- ChatGPT generated

CREATE TABLE IF NOT EXISTS "torrent" (
    "torrent_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "announce" TEXT NOT NULL,
    "size" INTEGER NOT NULL,
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
    "torrent_id" INTEGER NOT NULL,
    "ip" TEXT NOT NULL,
    "port" INTEGER NOT NULL,
    FOREIGN KEY ("torrent_id") REFERENCES "torrent" ("torrent_id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "piece" (
    "piece_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "torrent_id" INTEGER,
    "located_at_peer_id" INTEGER,
    "came_from_peer_id" INTEGER,
    "start" DATETIME,
    "end" DATETIME,
    "index" INTEGER,
    "length" INTEGER,
    FOREIGN KEY ("torrent_id") REFERENCES "torrent"("torrent_id"),
    FOREIGN KEY ("located_at_peer_id") REFERENCES "peer"("peer_id"),
    FOREIGN KEY ("came_from_peer_id") REFERENCES "peer"("peer_id")
);

CREATE TABLE IF NOT EXISTS "clients" (
    "client_id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "protocol_id" BLOB,
    "created" DATETIME
);

CREATE TABLE IF NOT EXISTS "connections" (
    "torrent_id"          INTEGER,
    "remote_peer_id"      INTEGER,
    "im_choked"           BOOLEAN,
    "remote_is_choked"    BOOLEAN,
    "im_interested"       BOOLEAN,
    "remote_is_interested" BOOLEAN,
    "download_rate"       REAL,
    "upload_rate"         REAL,
    "last_activity"       TIMESTAMP,
    
    PRIMARY KEY ("torrent_id", "remote_peer_id"),
    
    FOREIGN KEY ("torrent_id") REFERENCES "your_torrent_table_name"("torrent_id"),
    FOREIGN KEY ("remote_peer_id") REFERENCES "your_peer_table_name"("peer_id")
);