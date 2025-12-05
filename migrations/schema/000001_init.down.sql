-- Drop triggers
DROP TRIGGER IF EXISTS update_salons_updated_at ON salons;
DROP TRIGGER IF EXISTS update_sso_updated_at ON sso;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_salon_onboarding_user_step;
DROP INDEX IF EXISTS idx_salon_onboarding_user_id;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_appointment_id;
DROP INDEX IF EXISTS idx_appointments_status;
DROP INDEX IF EXISTS idx_appointments_date;
DROP INDEX IF EXISTS idx_appointments_user_id;
DROP INDEX IF EXISTS idx_appointments_salon_id;
DROP INDEX IF EXISTS idx_clients_phone;
DROP INDEX IF EXISTS idx_clients_salon_id;
DROP INDEX IF EXISTS idx_services_user_id;
DROP INDEX IF EXISTS idx_services_salon_id;
DROP INDEX IF EXISTS idx_salons_owner_id;
DROP INDEX IF EXISTS idx_salons_slug;
DROP INDEX IF EXISTS idx_sso_user_id;
DROP INDEX IF EXISTS idx_users_onboarding_step;
DROP INDEX IF EXISTS idx_users_salon_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_onboarding_options_step;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS salon_onboarding_selection;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS contact_form;
DROP TABLE IF EXISTS salons;
DROP TABLE IF EXISTS sso;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS onboarding_options;
DROP TABLE IF EXISTS onboarding_steps;

-- Drop custom types
DROP TYPE IF EXISTS contact_form_subject;
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS appointment_status;
DROP TYPE IF EXISTS user_role;