CREATE TABLE line_send_logs (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL REFERENCES clinics(id),
  owner_id BIGINT NOT NULL REFERENCES owners(id),
  sent_by_user_id BIGINT NOT NULL REFERENCES staffs(id),
  message_type VARCHAR(20) NOT NULL,
  content_summary TEXT NOT NULL,
  line_message_id VARCHAR(100),
  status VARCHAR(20) NOT NULL,
  error_message TEXT,
  sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_send_logs_clinic_owner ON line_send_logs (clinic_id, owner_id, sent_at DESC);
