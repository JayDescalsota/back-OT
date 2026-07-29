CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT 'User',
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_validated BOOLEAN NOT NULL DEFAULT false,
    validated_at TIMESTAMPTZ,
    validation_token TEXT,
    password_reset_at TIMESTAMPTZ,
    password_reset_token TEXT,
    password_reset_expires_at TIMESTAMPTZ,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_action TEXT,
    updated_action TEXT
);

CREATE INDEX idx_users_email ON users (email);

CREATE TABLE user_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token TEXT,
    refresh_token TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_refresh_token ON user_sessions (refresh_token);
CREATE INDEX idx_user_sessions_user_id ON user_sessions (user_id);

CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT,
    middle_name TEXT,
    suffix TEXT,
    title TEXT,
    phone TEXT,
    mobile TEXT,
    date_of_birth DATE,
    gender TEXT,
    street_address TEXT,
    barangay TEXT,
    city TEXT,
    province TEXT,
    zip TEXT,
    country TEXT NOT NULL DEFAULT 'PH',
    timezone TEXT NOT NULL DEFAULT 'Asia/Manila',
    preferred_language TEXT NOT NULL DEFAULT 'en',
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,
    emergency_contact_relation TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_action TEXT,
    updated_action TEXT
);

CREATE TABLE user_practitioner_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    license_number TEXT,
    license_state TEXT,
    npi_number TEXT UNIQUE,
    specialty TEXT,
    sub_specialty TEXT,
    qualifications TEXT[],
    credentials TEXT[],
    education TEXT,
    years_of_experience INT,
    bio TEXT,
    is_accepting_patients BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_action TEXT,
    updated_action TEXT
);
