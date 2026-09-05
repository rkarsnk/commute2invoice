-- DB スキーマ定義（SQLite）
-- このスキーマは公式仕様書に基づいています。

-- 区間ごとの運賃証跡テーブル
CREATE TABLE IF NOT EXISTS fare_evidence (
  id              INTEGER PRIMARY KEY,
  label           TEXT NOT NULL,
  company_text    TEXT NOT NULL,
  from_text       TEXT NOT NULL,
  to_text         TEXT NOT NULL,
  fare_oneway     INTEGER NOT NULL,
  fare_roundtrip  INTEGER,
  image           BLOB NOT NULL,
  image_mime      TEXT NOT NULL DEFAULT 'image/png',
  valid_from      TEXT NOT NULL,
  valid_until     TEXT,
  created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 利用実績
CREATE TABLE IF NOT EXISTS trips (
  id                INTEGER PRIMARY KEY,
  date              TEXT NOT NULL,
  fare_evidence_id  INTEGER NOT NULL REFERENCES fare_evidence(id),
  round_trip        INTEGER NOT NULL DEFAULT 1,
  billing_target    TEXT NOT NULL CHECK (billing_target IN ('自社','客先')),
  created_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_trips_date ON trips(date);
CREATE INDEX IF NOT EXISTS idx_trips_billing_date ON trips(billing_target, date);

-- 月末に確定した請求書PDFのスナップショット
CREATE TABLE IF NOT EXISTS invoices (
  id              INTEGER PRIMARY KEY,
  year_month      TEXT NOT NULL,
  billing_target  TEXT NOT NULL CHECK (billing_target IN ('自社','客先')),
  pdf             BLOB NOT NULL,
  generated_at    TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(year_month, billing_target)
);
