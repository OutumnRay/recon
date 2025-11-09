-- Convert all TIMESTAMP columns to TIMESTAMPTZ (timestamp with time zone)
-- This ensures all dates are stored in UTC

-- meetings table
ALTER TABLE meetings
  ALTER COLUMN scheduled_at TYPE TIMESTAMPTZ USING scheduled_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- meeting_participants table
ALTER TABLE meeting_participants
  ALTER COLUMN joined_at TYPE TIMESTAMPTZ USING joined_at AT TIME ZONE 'UTC',
  ALTER COLUMN left_at TYPE TIMESTAMPTZ USING left_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

-- meeting_subjects table
ALTER TABLE meeting_subjects
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- meeting_departments table
ALTER TABLE meeting_departments
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

-- users table
ALTER TABLE users
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- departments table
ALTER TABLE departments
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- groups table
ALTER TABLE groups
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- user_groups table
ALTER TABLE user_groups
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- livekit_rooms table
ALTER TABLE livekit_rooms
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN started_at TYPE TIMESTAMPTZ USING started_at AT TIME ZONE 'UTC',
  ALTER COLUMN finished_at TYPE TIMESTAMPTZ USING finished_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at_db TYPE TIMESTAMPTZ USING created_at_db AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- livekit_participants table
ALTER TABLE livekit_participants
  ALTER COLUMN joined_at TYPE TIMESTAMPTZ USING joined_at AT TIME ZONE 'UTC',
  ALTER COLUMN left_at TYPE TIMESTAMPTZ USING left_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at_db TYPE TIMESTAMPTZ USING created_at_db AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- livekit_tracks table
ALTER TABLE livekit_tracks
  ALTER COLUMN published_at TYPE TIMESTAMPTZ USING published_at AT TIME ZONE 'UTC',
  ALTER COLUMN unpublished_at TYPE TIMESTAMPTZ USING unpublished_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at_db TYPE TIMESTAMPTZ USING created_at_db AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- livekit_webhook_events table
ALTER TABLE livekit_webhook_events
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

-- files table (if exists)
ALTER TABLE files
  ALTER COLUMN uploaded_at TYPE TIMESTAMPTZ USING uploaded_at AT TIME ZONE 'UTC',
  ALTER COLUMN processed_at TYPE TIMESTAMPTZ USING processed_at AT TIME ZONE 'UTC',
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
