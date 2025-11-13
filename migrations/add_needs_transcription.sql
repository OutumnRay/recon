-- Add needs_transcription column to meetings table
-- This migration adds support for per-meeting transcription settings

-- Add the column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'meetings'
        AND column_name = 'needs_transcription'
    ) THEN
        ALTER TABLE meetings
        ADD COLUMN needs_transcription BOOLEAN NOT NULL DEFAULT false;

        RAISE NOTICE 'Added needs_transcription column to meetings table';
    ELSE
        RAISE NOTICE 'Column needs_transcription already exists in meetings table';
    END IF;
END $$;

-- Add comment to document the column
COMMENT ON COLUMN meetings.needs_transcription IS 'Indicates whether to record individual audio tracks for transcription';
