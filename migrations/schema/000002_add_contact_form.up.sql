CREATE TYPE contact_form_subject AS ENUM ('support', 'sales', 'partnership');
-- Contact Form table
CREATE TABLE contact_form (
    id UUID PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    salon_name VARCHAR(100),
    subject contact_form_subject NOT NULL,
    message VARCHAR(400) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMPTZ
);