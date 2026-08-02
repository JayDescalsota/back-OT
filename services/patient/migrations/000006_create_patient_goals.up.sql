CREATE TABLE IF NOT EXISTS patient_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    goal TEXT NOT NULL,
    target TEXT,
    progress INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'Not Started',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    created_action TEXT,
    updated_by UUID,
    updated_action TEXT
);

CREATE INDEX idx_patient_goals_patient ON patient_goals (patient_id);
CREATE INDEX idx_patient_goals_active ON patient_goals (is_active);
