-- Backfill user_profiles from existing users (split name by first space)
INSERT INTO user_profiles (user_id, first_name, last_name)
SELECT id,
       CASE WHEN POSITION(' ' IN COALESCE(name, email)) > 0
            THEN LEFT(COALESCE(name, email), POSITION(' ' IN COALESCE(name, email)) - 1)
            ELSE COALESCE(name, email)
       END,
       CASE WHEN POSITION(' ' IN COALESCE(name, email)) > 0
            THEN SUBSTRING(COALESCE(name, email) FROM POSITION(' ' IN COALESCE(name, email)) + 1)
            ELSE NULL
       END
FROM users
ON CONFLICT (user_id) DO NOTHING;
