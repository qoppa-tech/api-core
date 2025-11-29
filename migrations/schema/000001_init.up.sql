CREATE TYPE user_role as ENUM('admin', 'owner', 'employee', 'customer');
CREATE TYPE appointment_status AS ENUM ('pending', 'confirmed', 'cancelled', 'done', 'no-show');
CREATE TYPE notification_type AS ENUM ('whatsapp', 'email', 'sms');
CREATE TYPE notification_status AS ENUM ('queued', 'sent', 'failed');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    role user_role NOT NULL DEFAULT 'customer',
    salon_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SSO table (for OAuth providers)
CREATE TABLE sso (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(51) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Session table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Salons table
CREATE TABLE salons (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    address TEXT NOT NULL,
    whatsapp VARCHAR(20),
    timezone VARCHAR(10) NOT NULL DEFAULT 'UTC',
    business_hours JSONB NOT NULL DEFAULT '{
        "monday": ["6:00-12:00", "14:00-20:00"],
        "tuesday": ["6:00-12:00", "14:00-20:00"],
        "wednesday": ["6:00-12:00", "14:00-20:00"],
        "thursday": ["6:00-12:00", "14:00-20:00"],
        "friday": ["6:00-12:00", "14:00-20:00"],
        "saturday": ["6:00-12:00", "14:00-20:00"],
        "sunday": ["6:00-12:00", "14:00-20:00"]
    }'::jsonb,
    logo_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add foreign key constraint to users.salon_id after salons table is created
ALTER TABLE users ADD CONSTRAINT fk_users_salon FOREIGN KEY (salon_id) REFERENCES salons(id) ON DELETE SET NULL;

-- Professionals table
CREATE TABLE professionals (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    photo_url VARCHAR(500),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, salon_id)
);

-- Services table
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    professional_id UUID REFERENCES professionals(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    duration INTEGER NOT NULL CHECK (duration > 0),
    price INTEGER NOT NULL CHECK (price >= 0),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Clients table
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    birthday DATE,
    total_appointments INTEGER NOT NULL DEFAULT 0,
    last_appointment TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(salon_id, phone)
);

-- Appointments table
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    professional_id UUID NOT NULL REFERENCES professionals(id) ON DELETE RESTRICT,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    client_name VARCHAR(100) NOT NULL,
    client_phone VARCHAR(20) NOT NULL,
    client_email VARCHAR(255),
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status appointment_status NOT NULL DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

-- Notifications table (future feature)
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuidv4(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    status notification_status NOT NULL DEFAULT 'queued',
    scheduled_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_salon_id ON users(salon_id);
CREATE INDEX idx_sso_user_id ON sso(user_id);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_salons_slug ON salons(slug);
CREATE INDEX idx_salons_owner_id ON salons(owner_id);
CREATE INDEX idx_professionals_user_id ON professionals(user_id);
CREATE INDEX idx_professionals_salon_id ON professionals(salon_id);
CREATE INDEX idx_services_salon_id ON services(salon_id);
CREATE INDEX idx_services_professional_id ON services(professional_id);
CREATE INDEX idx_clients_salon_id ON clients(salon_id);
CREATE INDEX idx_clients_phone ON clients(salon_id, phone);
CREATE INDEX idx_appointments_salon_id ON appointments(salon_id);
CREATE INDEX idx_appointments_professional_id ON appointments(professional_id);
CREATE INDEX idx_appointments_date ON appointments(date);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_notifications_appointment_id ON notifications(appointment_id);
CREATE INDEX idx_notifications_status ON notifications(status);

-- WARN: THIS MUST BE REMEMBERED DURING DOWN MIGRATION
-- Create trigger function for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sso_updated_at BEFORE UPDATE ON sso
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_salons_updated_at BEFORE UPDATE ON salons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_professionals_updated_at BEFORE UPDATE ON professionals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
