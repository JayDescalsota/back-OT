ALTER TABLE booking_appointments ADD COLUMN cancellation_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE booking_appointments ADD COLUMN cancelled_at TIMESTAMPTZ;
ALTER TABLE booking_appointments ADD COLUMN rescheduled_from_id UUID REFERENCES booking_appointments(id);
ALTER TABLE booking_appointments ADD COLUMN rescheduled_to_id UUID REFERENCES booking_appointments(id);
ALTER TABLE booking_appointments ADD COLUMN reschedule_reason TEXT NOT NULL DEFAULT '';
