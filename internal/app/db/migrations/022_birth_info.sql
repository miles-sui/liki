-- 022_birth_info.sql — Store birth info for BaZi chart on user profile.
ALTER TABLE users ADD COLUMN birth_info_json TEXT NOT NULL DEFAULT '';
