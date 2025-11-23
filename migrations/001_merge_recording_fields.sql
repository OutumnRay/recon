-- Migration: Merge needs_video_record and needs_audio_record into needs_record
-- Date: 2025-01-23
-- Description: This migration consolidates the separate video and audio recording flags
--              into a single 'needs_record' field that covers both audio and video recording.

-- Step 1: Add the new needs_record column
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS needs_record BOOLEAN NOT NULL DEFAULT false;

-- Step 2: Migrate existing data
-- If either video or audio recording was enabled, set needs_record to true
UPDATE meetings
SET needs_record = (needs_video_record = true OR needs_audio_record = true)
WHERE needs_video_record IS NOT NULL OR needs_audio_record IS NOT NULL;

-- Step 3: Drop the old columns (commented out for safety - uncomment after verifying data migration)
-- ALTER TABLE meetings DROP COLUMN IF EXISTS needs_video_record;
-- ALTER TABLE meetings DROP COLUMN IF EXISTS needs_audio_record;

-- Note: After running this migration and verifying that everything works correctly,
-- you should uncomment the DROP COLUMN statements above and run them to complete the migration.
