CREATE TABLE IF NOT EXISTS hashes
(
    step_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    hash      BLOB NOT NULL,
    PRIMARY KEY (step_name, file_path)
);
