-- Profile/settings columns required by the frontend (Settings page, avatar
-- selection) and by the signup flow (full name). Additive only.

ALTER TABLE platform.users
    ADD COLUMN full_name  TEXT,
    ADD COLUMN timezone   TEXT NOT NULL DEFAULT 'UTC',
    ADD COLUMN theme      TEXT NOT NULL DEFAULT 'default';

ALTER TABLE platform.learner_profiles
    ADD COLUMN gender      TEXT,
    ADD COLUMN avatar_url  TEXT,
    ADD COLUMN role        TEXT;