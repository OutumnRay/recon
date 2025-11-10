-- Revert TIMESTAMPTZ back to TIMESTAMP (not recommended, included for completeness)

-- meetings table
ALTER TABLE meetings
  ALTER COLUMN scheduled_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- meeting_participants table
ALTER TABLE meeting_participants
  ALTER COLUMN joined_at TYPE TIMESTAMP,
  ALTER COLUMN left_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP;

-- meeting_subjects table
ALTER TABLE meeting_subjects
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- meeting_departments table
ALTER TABLE meeting_departments
  ALTER COLUMN created_at TYPE TIMESTAMP;

-- users table
ALTER TABLE users
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- departments table
ALTER TABLE departments
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- groups table
ALTER TABLE groups
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- user_groups table
ALTER TABLE user_groups
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- livekit_rooms table
ALTER TABLE livekit_rooms
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN started_at TYPE TIMESTAMP,
  ALTER COLUMN finished_at TYPE TIMESTAMP,
  ALTER COLUMN created_at_db TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- livekit_participants table
ALTER TABLE livekit_participants
  ALTER COLUMN joined_at TYPE TIMESTAMP,
  ALTER COLUMN left_at TYPE TIMESTAMP,
  ALTER COLUMN created_at_db TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- livekit_tracks table
ALTER TABLE livekit_tracks
  ALTER COLUMN published_at TYPE TIMESTAMP,
  ALTER COLUMN unpublished_at TYPE TIMESTAMP,
  ALTER COLUMN created_at_db TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;

-- livekit_webhook_events table
ALTER TABLE livekit_webhook_events
  ALTER COLUMN created_at TYPE TIMESTAMP;

-- files table
ALTER TABLE files
  ALTER COLUMN uploaded_at TYPE TIMESTAMP,
  ALTER COLUMN processed_at TYPE TIMESTAMP,
  ALTER COLUMN created_at TYPE TIMESTAMP,
  ALTER COLUMN updated_at TYPE TIMESTAMP;
