ALTER TABLE platform.learner_profiles
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS gender;

ALTER TABLE platform.users
    DROP COLUMN IF EXISTS theme,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS full_name;