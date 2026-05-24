CREATE TABLE IF NOT EXISTS risk_decisions (
  decision_id TEXT PRIMARY KEY,
  user_hash TEXT,
  input_type TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  score NUMERIC(5, 4) NOT NULL,
  confidence NUMERIC(5, 4) NOT NULL,
  scam_type TEXT NOT NULL,
  top_signals JSONB NOT NULL DEFAULT '[]'::jsonb,
  recommended_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
  needs_human_review BOOLEAN NOT NULL DEFAULT FALSE,
  payee_hash TEXT,
  report_id TEXT,
  model_versions JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payee_risk_profiles (
  payee_hash TEXT PRIMARY KEY,
  risk_score NUMERIC(5, 4) NOT NULL DEFAULT 0,
  complaint_count INTEGER NOT NULL DEFAULT 0,
  aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  connected_risk_count INTEGER NOT NULL DEFAULT 0,
  review_status TEXT NOT NULL DEFAULT 'UNREVIEWED'
);

CREATE TABLE IF NOT EXISTS feedback_events (
  feedback_id TEXT PRIMARY KEY,
  decision_id TEXT,
  user_hash TEXT,
  verdict TEXT NOT NULL,
  payee_hash TEXT,
  comment TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS recovery_reports (
  report_id TEXT PRIMARY KEY,
  user_hash TEXT,
  decision_id TEXT,
  status TEXT NOT NULL,
  structured_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  official_steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS evidence_objects (
  evidence_id TEXT PRIMARY KEY,
  report_id TEXT REFERENCES recovery_reports(report_id),
  object_key TEXT NOT NULL,
  media_type TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  retention_until TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS event_log (
  event_id TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  schema_version TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  causation_id TEXT,
  producer TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_risk_decisions_created_at ON risk_decisions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_decisions_payee_hash ON risk_decisions(payee_hash);
CREATE INDEX IF NOT EXISTS idx_feedback_events_decision_id ON feedback_events(decision_id);
CREATE INDEX IF NOT EXISTS idx_event_log_correlation_id ON event_log(correlation_id);

