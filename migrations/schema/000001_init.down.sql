-- Drop triggers
DROP TRIGGER IF EXISTS update_professionals_updated_at ON professionals;
DROP TRIGGER IF EXISTS update_salons_updated_at ON salons;
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
DROP TRIGGER IF EXISTS update_sso_updated_at ON sso;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

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