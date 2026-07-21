DELETE FROM schedule_exception WHERE created_action = 'seed';
DELETE FROM appointment WHERE created_action = 'seed';
DELETE FROM appointment_slot WHERE created_action = 'seed';
DELETE FROM schedule_template WHERE created_action = 'seed';
DELETE FROM practitioner_availability WHERE created_action = 'seed';
DELETE FROM practitioner_branch WHERE created_action = 'seed';
DELETE FROM branch_hour WHERE created_action = 'seed';
