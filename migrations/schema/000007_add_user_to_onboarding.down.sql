DROP INDEX IF EXISTS idx_users_onboarding_step;

ALTER TABLE users
  DROP COLUMN current_onboarding_step_id,
  DROP COLUMN onboarding_completed_at;

DROP INDEX IF EXISTS idx_salon_onboarding_user_step;
DROP INDEX IF EXISTS idx_salon_onboarding_user_id;

ALTER TABLE salon_onboarding_selection 
  DROP COLUMN user_id;
