CREATE TABLE IF NOT EXISTS jobs (
  name TEXT,
  run_at TIMESTAMP,
  PRIMARY KEY (name, run_at)
);
