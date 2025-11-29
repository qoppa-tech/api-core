-- Drop triggers
DROP TRIGGER IF EXISTS update_professionals_updated_at ON professionals;
DROP TRIGGER IF EXISTS update_salons_updated_at ON salons;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
DROP TRIGGER IF EXISTS update_sso_updated_at ON sso;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_appointment_id;
DROP INDEX IF EXISTS idx_appointments_status;
DROP INDEX IF EXISTS idx_appointments_date;
DROP INDEX IF EXISTS idx_appointments_professional_id;
DROP INDEX IF EXISTS idx_appointments_salon_id;
DROP INDEX IF EXISTS idx_clients_phone;
DROP INDEX IF EXISTS idx_clients_salon_id;
DROP INDEX IF EXISTS idx_services_professional_id;
DROP INDEX IF EXISTS idx_services_salon_id;
DROP INDEX IF EXISTS idx_professionals_salon_id;
DROP INDEX IF EXISTS idx_professionals_user_id;
DROP INDEX IF EXISTS idx_salons_owner_id;
DROP INDEX IF EXISTS idx_salons_slug;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sso_user_id;
DROP INDEX IF EXISTS idx_users_salon_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS professionals;
DROP TABLE IF EXISTS salons;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS sso;
DROP TABLE IF EXISTS users;

-- Drop custom types
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS appointment_status;
DROP TYPE IF EXISTS user_role;