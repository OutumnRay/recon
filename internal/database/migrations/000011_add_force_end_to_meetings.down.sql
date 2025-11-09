-- Remove force_end_at_duration column from meetings table
ALTER TABLE meetings
DROP COLUMN force_end_at_duration;
