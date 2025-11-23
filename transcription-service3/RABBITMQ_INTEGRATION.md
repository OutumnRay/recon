# RabbitMQ Integration - Queue-Based Results

## Overview

Changed from webhook-based notifications to **RabbitMQ queue-based results**. The transcription service now sends completed transcription data back to a RabbitMQ result queue, which the Go backend consumes.

## Architecture

```
┌──────────────┐                  ┌─────────────────────────┐
│  Go Backend  │                  │ Transcription Service   │
│              │  ──task──>       │  (Python Worker)        │
│              │  transcription   │                         │
│              │    _queue        │  1. Download audio      │
│              │                  │  2. Transcribe          │
│              │                  │  3. Save JSON to MinIO  │
│              │  <──result──     │  4. Send to result queue│
│              │  transcription   │                         │
│              │   _results       └─────────────────────────┘
│              │
│  Consumer    │
│  Goroutine   │
└──────────────┘
```

## Changes Made

### Python Service (transcription-service3)

#### 1. Removed Webhook Functionality
- ✅ Removed `send_webhook()` method
- ✅ Removed `requests` import
- ✅ Removed `WEBHOOK_URL` from config

#### 2. Added RabbitMQ Result Queue
- ✅ Added `RABBITMQ_RESULT_QUEUE` configuration
- ✅ Created `send_result_to_queue()` method
- ✅ Declared result queue on connection

#### 3. Result Message Format

**Queue**: `transcription_results` (configurable)

**Message**:
```json
{
  "event": "transcription_completed",
  "track_id": "uuid",
  "user_id": "uuid",
  "audio_url": "https://...",
  "json_url": "https://api.storage.recontext.online/bucket/path/transcription.json",
  "transcription": {
    "phrases": [
      {
        "start": 0.0,
        "end": 2.5,
        "text": "Hello world",
        "confidence": -0.234,
        "language": "en"
      }
    ],
    "phrase_count": 10,
    "total_duration": 45.5
  },
  "timestamp": "2024-11-23T19:30:00Z",
  "status": "completed"
}
```

### Go Backend

#### 1. Created Consumer (`transcription_consumer.go`)
- ✅ Consumes messages from `transcription_results` queue
- ✅ Parses JSON results
- ✅ Logs transcription completion
- ✅ Placeholder for database updates

#### 2. Integrated in `main.go`
- ✅ Starts consumer in background goroutine
- ✅ Runs alongside web server

#### 3. Key Functions

**`StartTranscriptionConsumer()`**
- Connects to RabbitMQ
- Declares result queue
- Starts consuming messages
- Runs forever in background

**`processTranscriptionResult()`**
- Parses message JSON
- Logs result details
- Calls database update function

**`updateTrackTranscriptionStatus()`**
- Placeholder for database updates
- TODO: Implement actual DB update logic

## Configuration

### Python (.env)
```bash
RABBITMQ_HOST=5.129.227.21
RABBITMQ_PORT=5672
RABBITMQ_USER=recontext
RABBITMQ_PASSWORD=je9rO4k6CQ3M
RABBITMQ_QUEUE=transcription_queue          # Input queue
RABBITMQ_RESULT_QUEUE=transcription_results # Output queue
```

### Go (environment variables)
```bash
RABBITMQ_URL=amqp://recontext:je9rO4k6CQ3M@5.129.227.21:5672/
RABBITMQ_RESULT_QUEUE=transcription_results
```

## Data Flow

1. **Backend sends transcription task**
   ```
   Queue: transcription_queue
   Message: {track_id, user_id, audio_url, language}
   ```

2. **Python service processes**
   - Downloads audio from MinIO
   - Transcribes with Whisper
   - Saves JSON to MinIO

3. **Python service sends result**
   ```
   Queue: transcription_results
   Message: {event, track_id, json_url, transcription...}
   ```

4. **Go backend receives result**
   - Consumer goroutine receives message
   - Processes result
   - Updates database
   - (Optional) Notifies user

## Database Schema Updates

You should add these fields to your `livekit_tracks` table:

```sql
ALTER TABLE livekit_tracks ADD COLUMN IF NOT EXISTS transcription_json_url TEXT;
ALTER TABLE livekit_tracks ADD COLUMN IF NOT EXISTS transcription_status VARCHAR(50);
ALTER TABLE livekit_tracks ADD COLUMN IF NOT EXISTS transcription_phrase_count INTEGER;
ALTER TABLE livekit_tracks ADD COLUMN IF NOT EXISTS transcription_duration FLOAT;
```

## Implementation TODO

In `transcription_consumer.go`, uncomment and adapt this code:

```go
func updateTrackTranscriptionStatus(trackID string, jsonURL string, phraseCount int, duration float64) {
    db := getDB() // Your database connection
    _, err := db.Exec(`
        UPDATE livekit_tracks
        SET
            transcription_status = 'completed',
            transcription_json_url = $1,
            transcription_phrase_count = $2,
            transcription_duration = $3,
            updated_at = NOW()
        WHERE id = $4
    `, jsonURL, phraseCount, duration, trackID)

    if err != nil {
        log.Printf("❌ Failed to update track %s: %v", trackID, err)
    } else {
        log.Printf("✅ Track %s updated successfully", trackID)
    }
}
```

## Benefits

### vs Webhooks

| Feature | Webhooks | RabbitMQ Queue |
|---------|----------|----------------|
| Reliability | ❌ HTTP errors, timeouts | ✅ Guaranteed delivery |
| Retries | ❌ Manual implementation | ✅ Built-in |
| Order | ❌ Not guaranteed | ✅ FIFO |
| Backpressure | ❌ No control | ✅ QoS control |
| Monitoring | ❌ Harder to monitor | ✅ Queue metrics |
| Coupling | ❌ Tight coupling | ✅ Loose coupling |

### Advantages

1. **Reliability**: Messages are persisted and guaranteed to be delivered
2. **Decoupling**: Services don't need to know each other's URLs
3. **Scalability**: Multiple consumers can process results in parallel
4. **Monitoring**: RabbitMQ management UI shows queue stats
5. **Error Handling**: Failed messages can be retried or sent to dead-letter queue
6. **Consistency**: Same messaging system for both directions

## Testing

### 1. Start Services

**Python:**
```bash
cd transcription-service3
python main.py
```

**Go:**
```bash
cd cmd/managing-portal
go run .
```

### 2. Send Test Task

Use your existing task sending code or manually publish to `transcription_queue`

### 3. Monitor Queues

Visit RabbitMQ Management UI:
```
http://5.129.227.21:15672
User: recontext
Password: je9rO4k6CQ3M
```

### 4. Check Logs

**Python output:**
```
📤 Sending result to RabbitMQ queue: transcription_results
✅ Result sent to queue successfully
```

**Go output:**
```
📥 Received transcription result:
  Track ID: uuid
  Status: completed
  JSON URL: https://...
  Phrases: 10
  Duration: 45.50 seconds
```

## Monitoring

### Queue Metrics

Monitor these metrics in RabbitMQ:
- **Message rate**: Messages published/delivered per second
- **Queue length**: Number of pending messages
- **Consumer count**: Number of active consumers
- **Ack rate**: Messages acknowledged per second

### Alerts

Set up alerts for:
- Queue length > threshold (e.g., 100 messages)
- No consumers connected
- High message age
- Consumer errors

## Error Handling

### Python Service
- If result queue send fails, logs warning but doesn't fail transcription
- Message is persistent (delivery_mode=2)

### Go Backend
- Manual acknowledgment ensures processing
- Failed messages can be requeued or sent to dead-letter queue
- Consumer reconnects automatically if connection drops

## Next Steps

1. ✅ Test integration end-to-end
2. ⬜ Implement database updates in `updateTrackTranscriptionStatus()`
3. ⬜ Add user notifications when transcription completes
4. ⬜ Set up monitoring and alerts
5. ⬜ Configure dead-letter queue for failed messages
6. ⬜ Add metrics collection

## Files Modified

### Python
- `worker.py` - Replaced webhook with queue sending
- `config.py` - Added RABBITMQ_RESULT_QUEUE
- `.env.example` - Added result queue configuration

### Go
- `transcription_consumer.go` - New file
- `main.go` - Start consumer on startup

## Summary

Successfully migrated from **HTTP webhooks** to **RabbitMQ queues** for transcription results. This provides better reliability, decoupling, and monitoring capabilities while keeping the same data format and workflow.
