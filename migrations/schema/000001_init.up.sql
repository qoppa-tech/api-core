-- Base schema with all migrations merged; professionals removed and professional references now point to users

CREATE TYPE user_role AS ENUM('admin', 'owner', 'employee', 'customer');
CREATE TYPE appointment_status AS ENUM ('pending', 'confirmed', 'cancelled', 'done', 'no-show');
CREATE TYPE notification_type AS ENUM ('whatsapp', 'email', 'sms'); -- SMS dont exist
CREATE TYPE notification_status AS ENUM ('queued', 'sent', 'failed');
CREATE TYPE contact_form_subject AS ENUM ('support', 'sales', 'partnership');

-- Onboarding reference data
CREATE TABLE onboarding_steps (
    id INT PRIMARY KEY
);

CREATE TABLE onboarding_options (
    id SERIAL PRIMARY KEY,
    step_id INT NOT NULL REFERENCES onboarding_steps(id) ON DELETE CASCADE,
    option_index INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_onboarding_options_step ON onboarding_options(step_id);

INSERT INTO onboarding_steps (id) VALUES
    (1),
    (2),
    (3),
    (4),
    (5);

INSERT INTO onboarding_options (step_id, option_index) VALUES
    (2, 1), (2, 2), (2, 3), (2, 4), (2, 5), (2, 6), (2, 7), (2, 8), (2, 9),
    (3, 1), (3, 2), (3, 3), (3, 4),
    (4, 1), (4, 2), (4, 3), (4, 4), (4, 5), (4, 6);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    role user_role NOT NULL DEFAULT 'customer',
    salon_id UUID,
    current_onboarding_step_id INT REFERENCES onboarding_steps(id) ON DELETE SET NULL,
    onboarding_completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- SSO table (for OAuth providers)
CREATE TABLE sso (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(51) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Salons table
CREATE TABLE salons (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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

-- Services table (professional replaced with user)
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    duration INTEGER NOT NULL CHECK (duration > 0),
    price INTEGER NOT NULL CHECK (price >= 0),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Clients table
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
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

-- Appointments table (professional replaced with user)
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    salon_id UUID NOT NULL REFERENCES salons(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
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
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    status notification_status NOT NULL DEFAULT 'queued',
    scheduled_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Contact Form table
CREATE TABLE contact_form (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    salon_name VARCHAR(100),
    subject contact_form_subject NOT NULL,
    message VARCHAR(400) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMPTZ
);

CREATE TABLE salon_onboarding_selection (
    salon_id UUID NOT NULL,
    step_id INT NOT NULL REFERENCES onboarding_steps(id),
    option_index INT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    chosen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (salon_id, step_id, option_index)
);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_salon_id ON users(salon_id);
CREATE INDEX idx_users_onboarding_step ON users(current_onboarding_step_id);
CREATE INDEX idx_sso_user_id ON sso(user_id);
CREATE INDEX idx_salons_slug ON salons(slug);
CREATE INDEX idx_salons_owner_id ON salons(owner_id);
CREATE INDEX idx_services_salon_id ON services(salon_id);
CREATE INDEX idx_services_user_id ON services(user_id);
CREATE INDEX idx_clients_salon_id ON clients(salon_id);
CREATE INDEX idx_clients_phone ON clients(salon_id, phone);
CREATE INDEX idx_appointments_salon_id ON appointments(salon_id);
CREATE INDEX idx_appointments_user_id ON appointments(user_id);
CREATE INDEX idx_appointments_date ON appointments(date);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_notifications_appointment_id ON notifications(appointment_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_salon_onboarding_user_id ON salon_onboarding_selection(user_id);
CREATE INDEX idx_salon_onboarding_user_step ON salon_onboarding_selection(user_id, step_id);

-- Create trigger function for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sso_updated_at BEFORE UPDATE ON sso
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_salons_updated_at BEFORE UPDATE ON salons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
