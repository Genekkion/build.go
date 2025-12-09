CREATE TABLE IF NOT EXISTS hashes
(
    file_path TEXT PRIMARY KEY,
    hash      BLOB
);