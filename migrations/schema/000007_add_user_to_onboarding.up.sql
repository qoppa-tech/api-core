ALTER TABLE salon_onboarding_selection 
  ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_salon_onboarding_user_id 
  ON salon_onboarding_selection(user_id);

CREATE INDEX idx_salon_onboarding_user_step 
  ON salon_onboarding_selection(user_id, step_id);

ALTER TABLE users
  ADD COLUMN current_onboarding_step_id INT REFERENCES onboarding_steps(id) ON DELETE SET NULL,
  ADD COLUMN onboarding_completed_at TIMESTAMPTZ;

CREATE INDEX idx_users_onboarding_step 
  ON users(current_onboarding_step_id);
