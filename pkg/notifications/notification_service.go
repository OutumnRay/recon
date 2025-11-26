package notifications

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType represents different types of real-time events
type EventType string

const (
	// Meeting events
	EventMeetingStatusChanged   EventType = "meeting.status_changed"
	EventMeetingUpdated         EventType = "meeting.updated"
	EventMeetingParticipantJoin EventType = "meeting.participant_join"
	EventMeetingParticipantLeave EventType = "meeting.participant_leave"

	// Recording events
	EventRecordingStarted       EventType = "recording.started"
	EventRecordingCompleted     EventType = "recording.completed"
	EventRecordingFailed        EventType = "recording.failed"
	EventRecordingProcessing    EventType = "recording.processing"

	// Transcription events
	EventTranscriptionStarted   EventType = "transcription.started"
	EventTranscriptionCompleted EventType = "transcription.completed"
	EventTranscriptionFailed    EventType = "transcription.failed"
	EventTranscriptionProgress  EventType = "transcription.progress"

	// Summary events
	EventSummaryStarted         EventType = "summary.started"
	EventSummaryCompleted       EventType = "summary.completed"
	EventSummaryFailed          EventType = "summary.failed"
	EventSummaryProgress        EventType = "summary.progress"

	// Composite video events
	EventCompositeVideoStarted  EventType = "composite_video.started"
	EventCompositeVideoCompleted EventType = "composite_video.completed"
	EventCompositeVideoFailed   EventType = "composite_video.failed"
)

// Notification represents a real-time notification
type Notification struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	EntityType  string                 `json:"entity_type"`  // meeting, recording, transcription, summary
	EntityID    string                 `json:"entity_id"`
	MeetingID   *uuid.UUID             `json:"meeting_id,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	ChangedFields map[string]interface{} `json:"changed_fields,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Message     string                 `json:"message,omitempty"`
}

// Subscriber represents a client subscribed to notifications
type Subscriber struct {
	ID       string
	UserID   uuid.UUID
	Channel  chan *Notification
	Filters  SubscriptionFilters
}

// SubscriptionFilters defines what notifications a subscriber wants
type SubscriptionFilters struct {
	MeetingIDs  []uuid.UUID  // Subscribe to specific meetings
	EventTypes  []EventType  // Subscribe to specific event types
	EntityTypes []string     // Subscribe to specific entity types
}

// NotificationService manages real-time notifications
type NotificationService struct {
	subscribers map[string]*Subscriber
	mu          sync.RWMutex
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		subscribers: make(map[string]*Subscriber),
	}
}

// Subscribe registers a new subscriber
func (ns *NotificationService) Subscribe(userID uuid.UUID, filters SubscriptionFilters) *Subscriber {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	subscriber := &Subscriber{
		ID:      uuid.New().String(),
		UserID:  userID,
		Channel: make(chan *Notification, 256),
		Filters: filters,
	}

	ns.subscribers[subscriber.ID] = subscriber
	return subscriber
}

// Unsubscribe removes a subscriber
func (ns *NotificationService) Unsubscribe(subscriberID string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if subscriber, ok := ns.subscribers[subscriberID]; ok {
		close(subscriber.Channel)
		delete(ns.subscribers, subscriberID)
	}
}

// Notify sends a notification to relevant subscribers
func (ns *NotificationService) Notify(notification *Notification) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	for _, subscriber := range ns.subscribers {
		if ns.shouldReceiveNotification(subscriber, notification) {
			// Non-blocking send
			select {
			case subscriber.Channel <- notification:
				// Successfully sent
			default:
				// Channel is full, skip this subscriber
				// TODO: Consider logging or handling slow subscribers
			}
		}
	}
}

// shouldReceiveNotification determines if a subscriber should receive a notification
func (ns *NotificationService) shouldReceiveNotification(subscriber *Subscriber, notification *Notification) bool {
	// Check if notification is for a specific meeting the subscriber is interested in
	if len(subscriber.Filters.MeetingIDs) > 0 && notification.MeetingID != nil {
		found := false
		for _, meetingID := range subscriber.Filters.MeetingIDs {
			if meetingID == *notification.MeetingID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if notification is for the subscriber's user
	if notification.UserID != nil && *notification.UserID != subscriber.UserID {
		return false
	}

	// Check event type filter
	if len(subscriber.Filters.EventTypes) > 0 {
		found := false
		for _, eventType := range subscriber.Filters.EventTypes {
			if eventType == notification.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check entity type filter
	if len(subscriber.Filters.EntityTypes) > 0 {
		found := false
		for _, entityType := range subscriber.Filters.EntityTypes {
			if entityType == notification.EntityType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetSubscriberCount returns the number of active subscribers
func (ns *NotificationService) GetSubscriberCount() int {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return len(ns.subscribers)
}

// Helper functions for creating notifications

// NewMeetingStatusNotification creates a notification for meeting status change
func NewMeetingStatusNotification(meetingID uuid.UUID, oldStatus, newStatus string, userID *uuid.UUID) *Notification {
	return &Notification{
		ID:          uuid.New().String(),
		Type:        EventMeetingStatusChanged,
		EntityType:  "meeting",
		EntityID:    meetingID.String(),
		MeetingID:   &meetingID,
		UserID:      userID,
		ChangedFields: map[string]interface{}{
			"status": map[string]string{
				"old": oldStatus,
				"new": newStatus,
			},
		},
		Timestamp: time.Now(),
		Message:   "Meeting status changed to " + newStatus,
	}
}

// NewRecordingStatusNotification creates a notification for recording status change
func NewRecordingStatusNotification(recordingID, roomSID string, meetingID *uuid.UUID, status string, eventType EventType) *Notification {
	return &Notification{
		ID:         uuid.New().String(),
		Type:       eventType,
		EntityType: "recording",
		EntityID:   recordingID,
		MeetingID:  meetingID,
		ChangedFields: map[string]interface{}{
			"status":   status,
			"room_sid": roomSID,
		},
		Timestamp: time.Now(),
		Message:   "Recording status: " + status,
	}
}

// NewTranscriptionStatusNotification creates a notification for transcription status change
func NewTranscriptionStatusNotification(trackID string, meetingID *uuid.UUID, status string, eventType EventType, errorMsg string) *Notification {
	notification := &Notification{
		ID:         uuid.New().String(),
		Type:       eventType,
		EntityType: "transcription",
		EntityID:   trackID,
		MeetingID:  meetingID,
		ChangedFields: map[string]interface{}{
			"status":   status,
			"track_id": trackID,
		},
		Timestamp: time.Now(),
		Message:   "Transcription status: " + status,
	}

	if errorMsg != "" {
		notification.ChangedFields["error"] = errorMsg
	}

	return notification
}

// NewSummaryStatusNotification creates a notification for summary generation status change
// roomSID - идентификатор комнаты для которой генерируется резюме
func NewSummaryStatusNotification(meetingID uuid.UUID, roomSID string, status string, eventType EventType, errorMsg string) *Notification {
	notification := &Notification{
		ID:         uuid.New().String(),
		Type:       eventType,
		EntityType: "summary",
		EntityID:   roomSID, // Используем roomSID как EntityID для идентификации конкретной комнаты
		MeetingID:  &meetingID,
		ChangedFields: map[string]interface{}{
			"status":   status,
			"room_sid": roomSID,
		},
		Timestamp: time.Now(),
		Message:   "Summary generation status: " + status,
	}

	if errorMsg != "" {
		notification.ChangedFields["error"] = errorMsg
	}

	return notification
}

// NewCompositeVideoStatusNotification creates a notification for composite video status change
func NewCompositeVideoStatusNotification(roomSID string, meetingID *uuid.UUID, status string, eventType EventType) *Notification {
	return &Notification{
		ID:         uuid.New().String(),
		Type:       eventType,
		EntityType: "composite_video",
		EntityID:   roomSID,
		MeetingID:  meetingID,
		ChangedFields: map[string]interface{}{
			"status":   status,
			"room_sid": roomSID,
		},
		Timestamp: time.Now(),
		Message:   "Composite video status: " + status,
	}
}

// ToJSON converts notification to JSON
func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}
