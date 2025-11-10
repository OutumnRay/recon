-- Add force_end_at_duration column to meetings table
ALTER TABLE meetings
ADD COLUMN force_end_at_duration BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN meetings.force_end_at_duration IS 'Force end meeting after scheduled duration';
