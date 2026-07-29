ALTER TABLE booking_appointments DROP COLUMN IF EXISTS reschedule_reason;
ALTER TABLE booking_appointments DROP COLUMN IF EXISTS rescheduled_to_id;
ALTER TABLE booking_appointments DROP COLUMN IF EXISTS rescheduled_from_id;
ALTER TABLE booking_appointments DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE booking_appointments DROP COLUMN IF EXISTS cancellation_reason;
