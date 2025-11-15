"""Main entry point for transcription service."""
from worker import TranscriptionConsumer


def main():
    """Start the transcription service."""
    consumer = TranscriptionConsumer()
    consumer.start()


if __name__ == '__main__':
    main()
