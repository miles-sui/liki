-- 027_scene_almanac.sql
-- Rename scene 'dates' to 'almanac' to match the 5-domain model.

UPDATE reports SET scene = 'almanac' WHERE scene = 'dates';
