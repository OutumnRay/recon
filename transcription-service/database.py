"""Database operations for transcription service."""
import psycopg2
from psycopg2.extras import RealDictCursor
from datetime import datetime
from typing import List, Dict, Optional
import uuid
import time
from config import Config


class DatabaseManager:
    """Manage database connections and operations."""

    def __init__(self):
        self.connection_string = Config.get_db_connection_string()
        self.conn = None

    def connect(self, max_retries=5, retry_delay=2):
        """Establish database connection with retry logic."""
        last_error = None
        for attempt in range(max_retries):
            try:
                print(f"Attempting to connect to database (attempt {attempt + 1}/{max_retries})...")
                self.conn = psycopg2.connect(self.connection_string)
                print("✅ Database connection established successfully")
                return self.conn
            except psycopg2.OperationalError as e:
                last_error = e
                if attempt < max_retries - 1:
                    print(f"⚠️  Database connection failed: {e}")
                    print(f"   Retrying in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                else:
                    print(f"❌ Failed to connect to database after {max_retries} attempts")
                    raise
        raise last_error

    def close(self):
        """Close database connection."""
        if self.conn:
            self.conn.close()
            self.conn = None

    def create_transcription_tables(self):
        """Create tables for transcription data if they don't exist."""
        with self.connect() as conn:
            with conn.cursor() as cur:
                # Create transcription_phrases table
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS transcription_phrases (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        track_id UUID NOT NULL,
                        user_id UUID NOT NULL,
                        phrase_index INTEGER NOT NULL,
                        start_time NUMERIC(10, 3) NOT NULL,
                        end_time NUMERIC(10, 3) NOT NULL,
                        text TEXT NOT NULL,
                        confidence NUMERIC(5, 4),
                        language VARCHAR(10),
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                        updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
                    );
                """)

                # Create indexes
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS idx_transcription_phrases_track_id
                    ON transcription_phrases(track_id);
                """)
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS idx_transcription_phrases_user_id
                    ON transcription_phrases(user_id);
                """)
                cur.execute("""
                    CREATE INDEX IF NOT EXISTS idx_transcription_phrases_start_time
                    ON transcription_phrases(start_time);
                """)

                # Create transcription_status table
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS transcription_status (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        track_id UUID NOT NULL UNIQUE,
                        status VARCHAR(50) NOT NULL,
                        started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                        completed_at TIMESTAMP WITH TIME ZONE,
                        error_message TEXT,
                        phrase_count INTEGER DEFAULT 0,
                        total_duration NUMERIC(10, 3)
                    );
                """)

                cur.execute("""
                    CREATE INDEX IF NOT EXISTS idx_transcription_status_track_id
                    ON transcription_status(track_id);
                """)

                conn.commit()

    def save_transcription_phrases(
        self,
        track_id: str,
        user_id: str,
        phrases: List[Dict]
    ) -> int:
        """
        Save transcription phrases to database.

        Args:
            track_id: UUID of the track
            user_id: UUID of the user
            phrases: List of phrase dictionaries with start, end, text, confidence

        Returns:
            Number of phrases saved
        """
        with self.connect() as conn:
            with conn.cursor() as cur:
                # Delete existing phrases for this track (in case of re-transcription)
                cur.execute(
                    "DELETE FROM transcription_phrases WHERE track_id = %s",
                    (track_id,)
                )

                # Insert new phrases
                for idx, phrase in enumerate(phrases):
                    cur.execute("""
                        INSERT INTO transcription_phrases
                        (track_id, user_id, phrase_index, start_time, end_time, text, confidence, language)
                        VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                    """, (
                        track_id,
                        user_id,
                        idx,
                        phrase.get('start', 0),
                        phrase.get('end', 0),
                        phrase.get('text', ''),
                        phrase.get('confidence'),
                        phrase.get('language')
                    ))

                conn.commit()
                return len(phrases)

    def update_transcription_status(
        self,
        track_id: str,
        status: str,
        error_message: Optional[str] = None,
        phrase_count: Optional[int] = None,
        total_duration: Optional[float] = None
    ):
        """
        Update transcription status for a track.

        Args:
            track_id: UUID of the track
            status: Status string (pending, processing, completed, failed)
            error_message: Error message if failed
            phrase_count: Number of phrases transcribed
            total_duration: Total duration of transcription
        """
        with self.connect() as conn:
            with conn.cursor() as cur:
                # Check if status record exists
                cur.execute(
                    "SELECT id FROM transcription_status WHERE track_id = %s",
                    (track_id,)
                )
                exists = cur.fetchone()

                if exists:
                    # Update existing record
                    if status == 'completed':
                        cur.execute("""
                            UPDATE transcription_status
                            SET status = %s, completed_at = NOW(), phrase_count = %s, total_duration = %s
                            WHERE track_id = %s
                        """, (status, phrase_count, total_duration, track_id))
                    elif status == 'failed':
                        cur.execute("""
                            UPDATE transcription_status
                            SET status = %s, error_message = %s, completed_at = NOW()
                            WHERE track_id = %s
                        """, (status, error_message, track_id))
                    else:
                        cur.execute("""
                            UPDATE transcription_status
                            SET status = %s
                            WHERE track_id = %s
                        """, (status, track_id))
                else:
                    # Insert new record
                    cur.execute("""
                        INSERT INTO transcription_status
                        (track_id, status, error_message, phrase_count, total_duration)
                        VALUES (%s, %s, %s, %s, %s)
                    """, (track_id, status, error_message, phrase_count, total_duration))

                conn.commit()

    def mark_track_ready(self, track_id: str):
        """
        Mark track as ready in the track_recordings table.

        Args:
            track_id: UUID of the track
        """
        with self.connect() as conn:
            with conn.cursor() as cur:
                # Update track_recordings table to mark as ready
                cur.execute("""
                    UPDATE track_recordings
                    SET transcription_status = 'completed', updated_at = NOW()
                    WHERE id = %s
                """, (track_id,))

                conn.commit()

    def get_transcription_phrases(self, track_id: str) -> List[Dict]:
        """
        Retrieve transcription phrases for a track.

        Args:
            track_id: UUID of the track

        Returns:
            List of phrase dictionaries
        """
        with self.connect() as conn:
            with conn.cursor(cursor_factory=RealDictCursor) as cur:
                cur.execute("""
                    SELECT phrase_index, start_time, end_time, text, confidence, language
                    FROM transcription_phrases
                    WHERE track_id = %s
                    ORDER BY phrase_index
                """, (track_id,))

                return cur.fetchall()
