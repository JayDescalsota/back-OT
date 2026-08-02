CREATE TABLE IF NOT EXISTS booking_appointment_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    appointment_id UUID NOT NULL REFERENCES booking_appointments(id) ON DELETE CASCADE,
    goal_id UUID NOT NULL,
    progress INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'Not Started',
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE INDEX idx_booking_appointment_goals_appointment ON booking_appointment_goals (appointment_id);
CREATE INDEX idx_booking_appointment_goals_goal ON booking_appointment_goals (goal_id);
CREATE INDEX idx_booking_appointment_goals_active ON booking_appointment_goals (is_active);

CREATE TABLE IF NOT EXISTS booking_soap_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    appointment_id UUID NOT NULL UNIQUE REFERENCES booking_appointments(id) ON DELETE CASCADE,
    subjective TEXT,
    objective TEXT,
    assessment TEXT,
    plan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE INDEX idx_booking_soap_notes_appointment ON booking_soap_notes (appointment_id);
