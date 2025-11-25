package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/rabbitmq"
	"github.com/google/uuid"
)

// LiveKitWebhook godoc
// @Summary Конечная точка webhook LiveKit
// @Description Получает события webhook от сервера LiveKit (room_started, participant_joined, track_published, track_unpublished, participant_left, room_finished)
// @Tags LiveKit
// @Accept json
// @Produce json
// @Param webhook body models.WebhookRequest true "Webhook payload"
// @Success 200 {object} models.WebhookResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /webhook/meet [post]
func (mp *ManagingPortal) liveKitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		mp.logger.Errorf("Failed to read webhook body: %s", err.Error())
		mp.respondWithError(w, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()

	// Parse webhook event
	var webhookReq models.WebhookRequest
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		mp.logger.Errorf("Failed to parse webhook JSON: %s", err.Error())
		mp.respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	mp.logger.Infof("Received LiveKit webhook event: %s (ID: %s)", webhookReq.Event, webhookReq.ID)

	// Print webhook payload to console for debugging (controlled by LIVEKIT_WEBHOOK_DEBUG env var)
	if os.Getenv("LIVEKIT_WEBHOOK_DEBUG") == "true" {
		fmt.Printf("\n=== LiveKit Webhook ===\n")
		fmt.Printf("Event: %s\n", webhookReq.Event)
		fmt.Printf("ID: %s\n", webhookReq.ID)
		fmt.Printf("CreatedAt: %s\n", webhookReq.CreatedAt)
		fmt.Printf("Payload:\n%s\n", string(body))
		fmt.Printf("=======================\n\n")
	}

	// Log the raw event to database
	eventLog := &models.WebhookEventLog{
		ID:             uuid.New(),
		EventType:      webhookReq.Event,
		EventID:        webhookReq.ID,
		RoomSID:        "",
		ParticipantSID: "",
		TrackSID:       "",
		Payload:        body,
		CreatedAt:      time.Now(),
	}

	// Extract IDs from webhook
	if webhookReq.Room != nil {
		if sid, ok := webhookReq.Room["sid"].(string); ok {
			eventLog.RoomSID = sid
		}
	}
	if webhookReq.Participant != nil {
		if sid, ok := webhookReq.Participant["sid"].(string); ok {
			eventLog.ParticipantSID = sid
		}
	}
	if webhookReq.Track != nil {
		if sid, ok := webhookReq.Track["sid"].(string); ok {
			eventLog.TrackSID = sid
		}
	}

	// Save event log (temporarily disabled)
	// if err := mp.liveKitRepo.LogWebhookEvent(eventLog); err != nil {
	// 	mp.logger.Errorf("Failed to log webhook event: %s", err.Error())
	// 	// Continue processing even if logging fails
	// }

	// Return success response immediately (async processing)
	response := models.WebhookResponse{
		Status:  "success",
		Message: fmt.Sprintf("Event %s accepted for processing", webhookReq.Event),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// Process event asynchronously in background goroutine
	go func() {
		var err error
		switch webhookReq.Event {
		case "room_started":
			err = mp.handleRoomStarted(webhookReq)
		case "participant_joined":
			err = mp.handleParticipantJoined(webhookReq)
		case "track_published":
			err = mp.handleTrackPublished(webhookReq)
		case "track_unpublished":
			err = mp.handleTrackUnpublished(webhookReq)
		case "participant_left":
			err = mp.handleParticipantLeft(webhookReq)
		case "room_finished":
			err = mp.handleRoomFinished(webhookReq)
		case "egress_started":
			err = mp.handleEgressStarted(webhookReq)
		case "egress_updated":
			err = mp.handleEgressUpdated(webhookReq)
		case "egress_ended":
			err = mp.handleEgressEnded(webhookReq)
		default:
			mp.logger.Errorf("Unknown webhook event type: %s", webhookReq.Event)
			return
		}

		if err != nil {
			mp.logger.Errorf("Failed to process %s event: %s", webhookReq.Event, err.Error())
		} else {
			mp.logger.Infof("✅ Successfully processed %s event asynchronously", webhookReq.Event)
		}
	}()
}

// handleRoomStarted обрабатывает событие создания комнаты LiveKit
// Создаёт запись в базе данных, связывает с встречей, настраивает запись при необходимости
func (mp *ManagingPortal) handleRoomStarted(req models.WebhookRequest) error {
	mp.logger.Infof("🏠 Processing room_started event...")

	if req.Room == nil {
		mp.logger.Errorf("❌ Room data is missing in webhook request")
		return fmt.Errorf("room data is missing")
	}

	room := &models.Room{
		ID:          uuid.New(),
		Status:      "active",
		StartedAt:   time.Now(),
		CreatedAtDB: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Extract room data
	if sid, ok := req.Room["sid"].(string); ok {
		room.SID = sid
		mp.logger.Infof("  📌 Room SID: %s", sid)
	}
	if name, ok := req.Room["name"].(string); ok {
		room.Name = name
		mp.logger.Infof("  📌 Room Name: %s", name)
	}
	if emptyTimeout, ok := req.Room["emptyTimeout"].(float64); ok {
		room.EmptyTimeout = int(emptyTimeout)
	}
	if departureTimeout, ok := req.Room["departureTimeout"].(float64); ok {
		room.DepartureTimeout = int(departureTimeout)
	}
	if creationTime, ok := req.Room["creationTime"].(string); ok {
		room.CreationTime = creationTime
	}
	if creationTimeMs, ok := req.Room["creationTimeMs"].(string); ok {
		room.CreationTimeMs = creationTimeMs
	}
	if turnPassword, ok := req.Room["turnPassword"].(string); ok {
		room.TurnPassword = turnPassword
	}

	// Parse enabled codecs
	if enabledCodecs, ok := req.Room["enabledCodecs"].([]interface{}); ok {
		for _, codec := range enabledCodecs {
			if codecMap, ok := codec.(map[string]interface{}); ok {
				if mime, ok := codecMap["mime"].(string); ok {
					room.EnabledCodecs = append(room.EnabledCodecs, models.EnabledCodec{Mime: mime})
				}
			}
		}
	}

	// Link room to meeting if room.Name is a valid UUID (meeting ID) BEFORE saving to database
	var needsTranscription bool
	var meetingUUID *uuid.UUID

	if room.Name != "" {
		mp.logger.Infof("Checking if room name '%s' is a meeting ID...", room.Name)
		if meetingID, err := uuid.Parse(room.Name); err == nil {
			mp.logger.Infof("Room name is valid UUID, looking up meeting %s", meetingID)
			meetingUUID = &meetingID

			// Set MeetingID in room before saving
			room.MeetingID = meetingUUID
			mp.logger.Infof("✅ Set room.MeetingID to %s", meetingID)

			// Find meeting by ID and update its LiveKitRoomID
			meeting, err := mp.meetingRepo.GetMeetingByID(meetingID)
			if err == nil {
				mp.logger.Infof("Found meeting %s: NeedsRecord=%v, NeedsTranscription=%v",
					meetingID, meeting.NeedsRecord, meeting.NeedsTranscription)

				// Link meeting to the room ID (not SID - SID is a string like RM_xxx)
				// We'll update this after saving the room
				needsTranscription = meeting.NeedsTranscription

				// Auto-enable is_recording and is_transcribing based on needs_* settings
				// This allows user to start with enabled recording/transcription and then manually control it
				updateNeeded := false
				if meeting.NeedsRecord && !meeting.IsRecording {
					meeting.IsRecording = true
					updateNeeded = true
					mp.logger.Infof("Auto-enabling is_recording for meeting %s (needs_record=%v)",
						meetingID, meeting.NeedsRecord)
				}
				if needsTranscription && !meeting.IsTranscribing {
					meeting.IsTranscribing = true
					updateNeeded = true
					mp.logger.Infof("Auto-enabling is_transcribing for meeting %s (needs_transcription=%v)",
						meetingID, needsTranscription)
				}
				if updateNeeded {
					if err := mp.meetingRepo.UpdateMeeting(meeting); err != nil {
						mp.logger.Errorf("Failed to update meeting recording state: %v", err)
					}
				}
			} else {
				mp.logger.Errorf("Failed to find meeting %s for room linking: %v", meetingID, err)
			}
		} else {
			mp.logger.Infof("Room name '%s' is not a valid UUID: %v", room.Name, err)
		}
	}

	// Save room to database with MeetingID set
	mp.logger.Infof("💾 Saving room to database (MeetingID: %v)...", meetingUUID)
	if err := mp.liveKitRepo.CreateRoom(room); err != nil {
		mp.logger.Errorf("❌ Failed to save room to database: %v", err)
		return err
	}
	mp.logger.Infof("✅ Room saved successfully (DB ID: %s, SID: %s, MeetingID: %v)", room.ID, room.SID, meetingUUID)

	// Now link the meeting to the room ID
	if meetingUUID != nil {
		meeting, err := mp.meetingRepo.GetMeetingByID(*meetingUUID)
		if err == nil {
			meeting.LiveKitRoomID = &room.ID
			if err := mp.meetingRepo.UpdateMeeting(meeting); err != nil {
				mp.logger.Errorf("Failed to link meeting %s to room %s (SID: %s): %v", meetingUUID, room.ID, room.SID, err)
			} else {
				mp.logger.Infof("Successfully linked meeting %s to LiveKit room %s (SID: %s)", meetingUUID, room.ID, room.SID)
			}
		}
	}

	// Room-level composite recording removed in favor of per-track recordings
	mp.logger.Infof("ℹ️ Room composite egress skipped - tracks will be recorded individually when published")

	mp.logger.Infof("✅ room_started event processed successfully")
	return nil
}

// handleParticipantJoined обрабатывает событие присоединения участника к комнате
// Создаёт запись участника, обеспечивает существование комнаты, сохраняет метаданные
func (mp *ManagingPortal) handleParticipantJoined(req models.WebhookRequest) error {
	mp.logger.Infof("👤 Processing participant_joined event...")

	if req.Participant == nil {
		mp.logger.Errorf("❌ Participant data is missing in webhook request")
		return fmt.Errorf("participant data is missing")
	}

	participant := &models.Participant{
		ID:          uuid.New(),
		State:       "ACTIVE",
		CreatedAtDB: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Extract participant data
	if sid, ok := req.Participant["sid"].(string); ok {
		participant.SID = sid
		mp.logger.Infof("  📌 Participant SID: %s", sid)
	}
	if identity, ok := req.Participant["identity"].(string); ok {
		participant.Identity = identity
		mp.logger.Infof("  📌 Identity: %s", identity)
	}
	if name, ok := req.Participant["name"].(string); ok {
		participant.Name = name
		mp.logger.Infof("  📌 Name: %s", name)
	}
	if state, ok := req.Participant["state"].(string); ok {
		participant.State = state
		mp.logger.Infof("  📌 State: %s", state)
	}
	if joinedAt, ok := req.Participant["joinedAt"].(string); ok {
		participant.JoinedAt = joinedAt
	}
	if joinedAtMs, ok := req.Participant["joinedAtMs"].(string); ok {
		participant.JoinedAtMs = joinedAtMs
	}
	if version, ok := req.Participant["version"].(float64); ok {
		participant.Version = int(version)
	}
	if isPublisher, ok := req.Participant["isPublisher"].(bool); ok {
		participant.IsPublisher = isPublisher
	}

	// Marshal permission as JSON
	if permission, ok := req.Participant["permission"]; ok {
		permJSON, _ := json.Marshal(permission)
		participant.Permission = permJSON
	}

	// Extract room SID
	var roomName string
	if req.Room != nil {
		if roomSID, ok := req.Room["sid"].(string); ok {
			participant.RoomSID = roomSID
			mp.logger.Infof("  📌 Room SID: %s", roomSID)
		}
		if name, ok := req.Room["name"].(string); ok {
			roomName = name
		}
	}

	// Ensure room exists before saving participant (fix for race condition)
	if participant.RoomSID != "" {
		mp.logger.Infof("🔍 Checking if room %s exists...", participant.RoomSID)
		_, err := mp.liveKitRepo.GetRoomBySID(participant.RoomSID)
		if err != nil {
			mp.logger.Infof("⚠️ Room %s not found in database yet, creating placeholder room from participant_joined event...", participant.RoomSID)
			// Create placeholder room to satisfy foreign key constraint
			placeholderRoom := &models.Room{
				ID:          uuid.New(),
				SID:         participant.RoomSID,
				Name:        roomName,
				Status:      "active",
				StartedAt:   time.Now(),
				CreatedAtDB: time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Check if roomName is a valid meeting UUID and set MeetingID
			if roomName != "" {
				if meetingID, err := uuid.Parse(roomName); err == nil {
					placeholderRoom.MeetingID = &meetingID
					mp.logger.Infof("  📌 Set placeholder room MeetingID to %s", meetingID)
				}
			}

			// Extract room details from req.Room if available
			if req.Room != nil {
				if emptyTimeout, ok := req.Room["emptyTimeout"].(float64); ok {
					placeholderRoom.EmptyTimeout = int(emptyTimeout)
				}
				if departureTimeout, ok := req.Room["departureTimeout"].(float64); ok {
					placeholderRoom.DepartureTimeout = int(departureTimeout)
				}
				if creationTime, ok := req.Room["creationTime"].(string); ok {
					placeholderRoom.CreationTime = creationTime
				}
				if creationTimeMs, ok := req.Room["creationTimeMs"].(string); ok {
					placeholderRoom.CreationTimeMs = creationTimeMs
				}
				if turnPassword, ok := req.Room["turnPassword"].(string); ok {
					placeholderRoom.TurnPassword = turnPassword
				}
				// Parse enabled codecs
				if enabledCodecs, ok := req.Room["enabledCodecs"].([]interface{}); ok {
					for _, codec := range enabledCodecs {
						if codecMap, ok := codec.(map[string]interface{}); ok {
							if mime, ok := codecMap["mime"].(string); ok {
								placeholderRoom.EnabledCodecs = append(placeholderRoom.EnabledCodecs, models.EnabledCodec{Mime: mime})
							}
						}
					}
				}
			}
			if err := mp.liveKitRepo.CreateRoom(placeholderRoom); err != nil {
				mp.logger.Errorf("❌ Failed to create placeholder room: %v", err)
				return fmt.Errorf("room does not exist and failed to create placeholder: %w", err)
			}
			mp.logger.Infof("✅ Created placeholder room %s from participant_joined event", participant.RoomSID)
		} else {
			mp.logger.Infof("✅ Room %s exists", participant.RoomSID)
		}
	}

	mp.logger.Infof("💾 Saving participant to database...")
	if err := mp.liveKitRepo.CreateParticipant(participant); err != nil {
		mp.logger.Errorf("❌ Failed to save participant: %v", err)
		return err
	}
	mp.logger.Infof("✅ Participant saved successfully (DB ID: %s, SID: %s, Identity: %s)", participant.ID, participant.SID, participant.Identity)
	mp.logger.Infof("✅ participant_joined event processed successfully")
	return nil
}

// handleTrackPublished обрабатывает событие публикации медиа-трека (аудио/видео)
// Сохраняет метаданные трека, запускает запись egress если требуется, создаёт задачу транскрибации
func (mp *ManagingPortal) handleTrackPublished(req models.WebhookRequest) error {
	mp.logger.Infof("🎬 Processing track_published event...")

	if req.Track == nil {
		mp.logger.Errorf("❌ Track data is missing in webhook request")
		return fmt.Errorf("track data is missing")
	}

	track := &models.Track{
		ID:          uuid.New(),
		Status:      "published",
		PublishedAt: time.Now(),
		CreatedAtDB: time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Extract track data
	if sid, ok := req.Track["sid"].(string); ok {
		track.SID = sid
		mp.logger.Infof("  📌 Track SID: %s", sid)
	}
	if trackType, ok := req.Track["type"].(string); ok {
		track.Type = trackType
		mp.logger.Infof("  📌 Track Type: %s", trackType)
	}
	if source, ok := req.Track["source"].(string); ok {
		track.Source = source
		mp.logger.Infof("  📌 Track Source: %s", source)
	}
	if mimeType, ok := req.Track["mimeType"].(string); ok {
		track.MimeType = mimeType
		mp.logger.Infof("  📌 MIME Type: %s", mimeType)
	}
	if mid, ok := req.Track["mid"].(string); ok {
		track.Mid = mid
	}
	if width, ok := req.Track["width"].(float64); ok {
		track.Width = int(width)
	}
	if height, ok := req.Track["height"].(float64); ok {
		track.Height = int(height)
	}
	if simulcast, ok := req.Track["simulcast"].(bool); ok {
		track.Simulcast = simulcast
		if simulcast {
			mp.logger.Infof("  📌 Simulcast: enabled")
		}
	}
	if stream, ok := req.Track["stream"].(string); ok {
		track.Stream = stream
	}
	if backupCodecPolicy, ok := req.Track["backupCodecPolicy"].(string); ok {
		track.BackupCodecPolicy = backupCodecPolicy
	}

	// Marshal layers as JSON
	if layers, ok := req.Track["layers"]; ok {
		layersJSON, _ := json.Marshal(layers)
		track.Layers = layersJSON
		mp.logger.Infof("  📌 Layers: %d layer(s) configured", len(layers.([]interface{})))
	}

	// Marshal codecs as JSON
	if codecs, ok := req.Track["codecs"]; ok {
		codecsJSON, _ := json.Marshal(codecs)
		track.Codecs = codecsJSON
		mp.logger.Infof("  📌 Codecs: %d codec(s) configured", len(codecs.([]interface{})))
	}

	// Marshal version as JSON
	if version, ok := req.Track["version"]; ok {
		versionJSON, _ := json.Marshal(version)
		track.Version = versionJSON
	}

	// Extract audio features
	if audioFeatures, ok := req.Track["audioFeatures"].([]interface{}); ok {
		mp.logger.Infof("  📌 Audio Features found: %d feature(s)", len(audioFeatures))
		for _, feature := range audioFeatures {
			if featureStr, ok := feature.(string); ok {
				track.AudioFeatures = append(track.AudioFeatures, featureStr)
				mp.logger.Infof("    ✓ %s", featureStr)
			}
		}
	}

	// Extract participant SID
	var participantSID string
	if req.Participant != nil {
		if sid, ok := req.Participant["sid"].(string); ok {
			track.ParticipantSID = sid
			participantSID = sid
			mp.logger.Infof("  📌 Participant SID: %s", sid)
		}
	}

	// Extract room SID and room name
	var roomName string
	if req.Room != nil {
		if sid, ok := req.Room["sid"].(string); ok {
			track.RoomSID = sid
			mp.logger.Infof("  📌 Room SID: %s", sid)
		}
		if name, ok := req.Room["name"].(string); ok {
			roomName = name
			mp.logger.Infof("  📌 Room Name: %s", name)
		}
	}

	// Ensure room exists before saving track (fix for race condition)
	if track.RoomSID != "" {
		mp.logger.Infof("🔍 Checking if room %s exists...", track.RoomSID)
		_, err := mp.liveKitRepo.GetRoomBySID(track.RoomSID)
		if err != nil {
			mp.logger.Infof("⚠️ Room %s not found in database yet, creating placeholder room from track_published event...", track.RoomSID)
			// Create placeholder room to satisfy foreign key constraint
			placeholderRoom := &models.Room{
				ID:          uuid.New(),
				SID:         track.RoomSID,
				Name:        roomName,
				Status:      "active",
				StartedAt:   time.Now(),
				CreatedAtDB: time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Check if roomName is a valid meeting UUID and set MeetingID
			if roomName != "" {
				if meetingID, err := uuid.Parse(roomName); err == nil {
					placeholderRoom.MeetingID = &meetingID
					mp.logger.Infof("  📌 Set placeholder room MeetingID to %s", meetingID)
				}
			}

			// Extract room details from req.Room if available
			if req.Room != nil {
				if emptyTimeout, ok := req.Room["emptyTimeout"].(float64); ok {
					placeholderRoom.EmptyTimeout = int(emptyTimeout)
				}
				if departureTimeout, ok := req.Room["departureTimeout"].(float64); ok {
					placeholderRoom.DepartureTimeout = int(departureTimeout)
				}
				if creationTime, ok := req.Room["creationTime"].(string); ok {
					placeholderRoom.CreationTime = creationTime
				}
				if creationTimeMs, ok := req.Room["creationTimeMs"].(string); ok {
					placeholderRoom.CreationTimeMs = creationTimeMs
				}
				if turnPassword, ok := req.Room["turnPassword"].(string); ok {
					placeholderRoom.TurnPassword = turnPassword
				}
				// Parse enabled codecs
				if enabledCodecs, ok := req.Room["enabledCodecs"].([]interface{}); ok {
					for _, codec := range enabledCodecs {
						if codecMap, ok := codec.(map[string]interface{}); ok {
							if mime, ok := codecMap["mime"].(string); ok {
								placeholderRoom.EnabledCodecs = append(placeholderRoom.EnabledCodecs, models.EnabledCodec{Mime: mime})
							}
						}
					}
				}
			}
			if err := mp.liveKitRepo.CreateRoom(placeholderRoom); err != nil {
				mp.logger.Errorf("❌ Failed to create placeholder room: %v", err)
				return fmt.Errorf("room does not exist and failed to create placeholder: %w", err)
			}
			mp.logger.Infof("✅ Created placeholder room %s from track_published event", track.RoomSID)
		} else {
			mp.logger.Infof("✅ Room %s exists", track.RoomSID)
		}
	}

	// Save track to database
	mp.logger.Infof("💾 Saving track to database...")
	if err := mp.liveKitRepo.CreateTrack(track); err != nil {
		mp.logger.Errorf("❌ Failed to save track to database: %v", err)
		return err
	}
	mp.logger.Infof("✅ Track saved successfully (DB ID: %s)", track.ID)

	// Log track summary
	mp.logger.Infof("📊 Track Summary: SID=%s | Source=%s | Type=%s | Room=%s | Participant=%s | MimeType=%s",
		track.SID, track.Source, track.Type, roomName, participantSID, track.MimeType)

	// Получаем настройки записи напрямую из базы по room name (meeting ID)
	var needsTranscription bool
	var needsAudioRecord bool
	var needsVideoRecord bool
	if roomName != "" {
		if meetingID, err := uuid.Parse(roomName); err == nil {
			meeting, err := mp.meetingRepo.GetMeetingByID(meetingID)
			if err == nil {
				needsTranscription = meeting.NeedsTranscription
				needsAudioRecord = meeting.NeedsRecord    // needs_record includes both audio and video
				needsVideoRecord = meeting.NeedsRecord    // needs_record includes both audio and video
				mp.logger.Infof("📝 Meeting %s: NeedsTranscription=%v, NeedsRecord=%v",
					meetingID, needsTranscription, meeting.NeedsRecord)
			} else {
				mp.logger.Infof("📝 Room name '%s' is not a meeting ID or meeting not found: %v", roomName, err)
			}
		} else {
			mp.logger.Infof("📝 Room name '%s' is not a valid meeting UUID", roomName)
		}
	} else {
		mp.logger.Infof("📝 Room name is empty, cannot check transcription settings")
	}

	// Start per-track egress recording for all audio/video tracks when recording is required
	isAudioTrack := track.Source == "MICROPHONE" || (track.Source == "SCREEN_SHARE_AUDIO") ||
		(track.MimeType != "" && strings.HasPrefix(track.MimeType, "audio/"))
	isVideoTrack := track.Type == "video" || strings.EqualFold(track.Source, "camera") ||
		strings.EqualFold(track.Source, "screen_share") || (track.MimeType != "" && strings.HasPrefix(track.MimeType, "video/"))

	shouldRecordAudio := isAudioTrack && (needsAudioRecord || needsTranscription)
	shouldRecordVideo := isVideoTrack && needsVideoRecord

	if shouldRecordAudio || shouldRecordVideo {
		mp.logger.Infof("🎥 Track requires recording - preparing egress...")
		mp.logger.Infof("  ✓ Track Source: %s", track.Source)
		mp.logger.Infof("  ✓ MIME Type: %s", track.MimeType)
		mp.logger.Infof("  ✓ Room Name: %s (empty=%v)", roomName, roomName == "")
		mp.logger.Infof("  ✓ Track SID: %s (empty=%v)", track.SID, track.SID == "")
		mp.logger.Infof("  ✓ Audio required: %v | Video required: %v", shouldRecordAudio, shouldRecordVideo)

		if roomName != "" && track.SID != "" {
			audioTrackID := ""
			videoTrackID := ""
			if shouldRecordAudio {
				audioTrackID = track.SID
			}
			if shouldRecordVideo {
				videoTrackID = track.SID
			}

			// Parse meeting ID from room name
			var meetingUUID *uuid.UUID
			if parsedMeetingID, err := uuid.Parse(roomName); err == nil {
				meetingUUID = &parsedMeetingID
			}

			// Start egress asynchronously to avoid webhook timeout - egress can take long time to respond
			go func(roomName string, roomSID string, partSID string, meetingID *uuid.UUID, audioID string, videoID string, trackSID string) {
				egressID, err := mp.startTrackCompositeEgress(roomName, roomSID, audioID, videoID)
				if err != nil {
					mp.logger.Errorf("❌ Failed to start track composite egress: %v", err)
				} else if egressID != "" {
					mp.logger.Infof("✅ Track egress started successfully: %s (track: %s, audioID=%s, videoID=%s)", egressID, trackSID, audioID, videoID)

					// Save EgressID to database
					err = mp.db.DB.Model(&models.Track{}).Where("sid = ?", trackSID).Update("egress_id", egressID).Error
					if err != nil {
						mp.logger.Errorf("❌ Failed to save track egress ID to database: %v", err)
					} else {
						mp.logger.Infof("✅ Track egress ID saved to database: %s", egressID)
					}

					// Create EgressRecording entry
					trackSIDPtr := &trackSID
					partSIDPtr := &partSID
					audioOnly := videoID == "" && audioID != ""
					egressRec := &models.EgressRecording{
						EgressID:       egressID,
						Type:           "track_composite",
						Status:         "started",
						RoomSID:        roomSID,
						RoomName:       roomName,
						MeetingID:      meetingID,
						TrackSID:       trackSIDPtr,
						ParticipantSID: partSIDPtr,
						AudioOnly:      audioOnly,
						StartedAt:      time.Now(),
					}
					err = mp.db.DB.Create(egressRec).Error
					if err != nil {
						mp.logger.Errorf("❌ Failed to create egress recording entry: %v", err)
					} else {
						mp.logger.Infof("✅ Egress recording entry created: %s (Type: track_composite, Track: %s, Meeting: %v)", egressID, trackSID, meetingID)
					}
				} else {
					mp.logger.Infof("⚠️ Track egress not started (disabled in config or returned empty ID)")
				}
			}(roomName, track.RoomSID, participantSID, meetingUUID, audioTrackID, videoTrackID, track.SID)

			mp.logger.Infof("ℹ️ Track egress request sent asynchronously")
		} else {
			mp.logger.Infof("⚠️ Track egress not started - missing required data:")
			mp.logger.Infof("  • Room Name empty: %v", roomName == "")
			mp.logger.Infof("  • Track SID empty: %v", track.SID == "")
		}
	} else {
		mp.logger.Infof("ℹ️ Track egress skipped - requirements not met (AudioNeeded=%v, VideoNeeded=%v, Transcription=%v, AudioTrack=%v, VideoTrack=%v)",
			needsAudioRecord, needsVideoRecord, needsTranscription, isAudioTrack, isVideoTrack)
	}

	mp.logger.Infof("✅ track_published event processed successfully")
	return nil
}

// handleTrackUnpublished обрабатывает событие отключения медиа-трека
// Останавливает запись egress, обновляет статус, создаёт задачу транскрибации для аудио-треков
func (mp *ManagingPortal) handleTrackUnpublished(req models.WebhookRequest) error {
	mp.logger.Infof("🔴 Processing track_unpublished event...")

	if req.Track == nil {
		mp.logger.Errorf("❌ Track data is missing in webhook request")
		return fmt.Errorf("track data is missing")
	}

	var trackSID string
	if sid, ok := req.Track["sid"].(string); ok {
		trackSID = sid
		mp.logger.Infof("  📌 Track SID: %s", sid)
	} else {
		mp.logger.Errorf("❌ Track SID is missing")
		return fmt.Errorf("track SID is missing")
	}

	// Extract room info for logging
	var roomSID string
	if req.Room != nil {
		if sid, ok := req.Room["sid"].(string); ok {
			roomSID = sid
			mp.logger.Infof("  📌 Room SID: %s", sid)
		}
		if name, ok := req.Room["name"].(string); ok {
			mp.logger.Infof("  📌 Room Name: %s", name)
		}
	}

	// Extract participant info for logging
	var participantSID string
	if req.Participant != nil {
		if sid, ok := req.Participant["sid"].(string); ok {
			participantSID = sid
			mp.logger.Infof("  📌 Participant SID: %s", sid)
		}
	}

	// Get track to find egress ID
	mp.logger.Infof("🔍 Looking up track in database...")
	track, err := mp.liveKitRepo.GetTrackBySID(trackSID)
	if err != nil {
		mp.logger.Errorf("❌ Failed to get track for unpublish: %v", err)
		// Continue to mark as unpublished even if we can't stop egress
	} else {
		mp.logger.Infof("✅ Track found: Source=%s, Type=%s, MimeType=%s", track.Source, track.Type, track.MimeType)

		if track.EgressID != "" {
			// Stop track egress recording
			mp.logger.Infof("🛑 Stopping track egress: %s (track: %s)", track.EgressID, trackSID)
			if err := mp.stopEgress(track.EgressID); err != nil {
				mp.logger.Errorf("❌ Failed to stop track egress %s: %v", track.EgressID, err)
			} else {
				mp.logger.Infof("✅ Stopped track egress: %s", track.EgressID)
			}
		} else {
			mp.logger.Infof("ℹ️ No egress ID found for track - nothing to stop")
		}
	}

	mp.logger.Infof("💾 Marking track as unpublished in database...")
	if err := mp.liveKitRepo.UnpublishTrack(trackSID); err != nil {
		mp.logger.Errorf("❌ Failed to mark track as unpublished: %v", err)
		return err
	}

	// Create transcription task if track has egress and is audio
	if track != nil && track.EgressID != "" {
		mp.logger.Infof("🎙️ Checking if transcription is needed for track %s...", trackSID)

		// Check if this is an audio track
		isAudioTrack := false
		if strings.EqualFold(track.Source, "microphone") || track.Type == "audio" {
			isAudioTrack = true
		}

		if isAudioTrack {
			mp.logger.Infof("✅ Track is audio, will create transcription task...")

			// Get room to find meeting
			room, err := mp.liveKitRepo.GetRoomBySID(roomSID)
			if err != nil {
				mp.logger.Errorf("❌ Failed to get room for transcription task: %v", err)
			} else if room.MeetingID != nil {
				// Get meeting to check if transcription is enabled
				meeting, err := mp.meetingRepo.GetMeetingByID(*room.MeetingID)
				if err != nil {
					mp.logger.Errorf("❌ Failed to get meeting for transcription task: %v", err)
				} else if meeting.NeedsTranscription {
					mp.logger.Infof("✅ Transcription is enabled for meeting %s", meeting.ID)

					// Get participant to find user identity
					participant, err := mp.liveKitRepo.GetParticipantBySID(participantSID)
					if err != nil {
						mp.logger.Errorf("❌ Failed to get participant for transcription: %v", err)
					} else {
						// Parse participant identity as user UUID
						userUUID, err := uuid.Parse(participant.Identity)
						if err != nil {
							mp.logger.Errorf("❌ Invalid participant identity format: %v", err)
						} else {
							// Construct audio URL from meeting ID, room SID and track SID
							// Pattern: https://api.storage.recontext.online/recontext/{meetingId}_{roomSid}/tracks/{trackSid}.m3u8
							storageURL := "https://api.storage.recontext.online"
							bucket := "recontext"
							audioURL := fmt.Sprintf("%s/%s/%s_%s/tracks/%s.m3u8",
								storageURL, bucket, meeting.ID.String(), room.SID, track.SID)

							// Parse track ID as UUID
							trackUUID, err := uuid.Parse(track.ID.String())
							if err != nil {
								mp.logger.Errorf("❌ Invalid track ID format: %v", err)
							} else {
								// Log the task details
								mp.logger.Infof("📋 ========================================")
								mp.logger.Infof("📋 TRANSCRIPTION TASK DETAILS:")
								mp.logger.Infof("📋 ========================================")
								mp.logger.Infof("📋 Track ID:       %s", trackUUID.String())
								mp.logger.Infof("📋 Track SID:      %s", track.SID)
								mp.logger.Infof("📋 User ID:        %s", userUUID.String())
								mp.logger.Infof("📋 Audio URL:      %s", audioURL)
								mp.logger.Infof("📋 Room SID:       %s", room.SID)
								mp.logger.Infof("📋 Room Name:      %s", room.Name)
								mp.logger.Infof("📋 Meeting ID:     %s", meeting.ID.String())
								mp.logger.Infof("📋 Meeting Title:  %s", meeting.Title)
								mp.logger.Infof("📋 Egress ID:      %s", track.EgressID)
								mp.logger.Infof("📋 Source:         track_unpublished event")
								mp.logger.Infof("📋 ========================================")

								// Wait 10 seconds for egress to finish writing the file
								mp.logger.Infof("⏳ Waiting 10 seconds for egress to finish writing audio file...")

								// Send task in background goroutine with delay
								go func(trackID uuid.UUID, userID uuid.UUID, url string, sid string) {
									time.Sleep(10 * time.Second)

									mp.logger.Infof("⏰ 10 seconds elapsed, sending transcription task...")

									// Send message to RabbitMQ
									if mp.rabbitMQPublisher != nil {
										err := mp.rabbitMQPublisher.PublishTranscriptionTask(
											trackID,
											userID,
											url,
											"", // Auto-detect language
											"", // No auth token needed
										)
										if err != nil {
											mp.logger.Errorf("❌ Failed to send transcription task to RabbitMQ: %v", err)
										} else {
											mp.logger.Infof("✅ ========================================")
											mp.logger.Infof("✅ Transcription task SUCCESSFULLY sent to RabbitMQ!")
											mp.logger.Infof("✅ Track: %s (SID: %s)", trackID.String(), sid)
											mp.logger.Infof("✅ Queue: transcription_queue")
											mp.logger.Infof("✅ Event: track_unpublished (delayed 10s)")
											mp.logger.Infof("✅ ========================================")
										}
									} else {
										mp.logger.Infof("⚠️ RabbitMQ publisher not initialized, skipping transcription task")
									}
								}(trackUUID, userUUID, audioURL, track.SID)
							}
						}
					}
				} else {
					mp.logger.Infof("ℹ️ Transcription is disabled for meeting %s", meeting.ID)
				}
			} else {
				mp.logger.Infof("ℹ️ Room has no associated meeting, skipping transcription")
			}
		} else {
			mp.logger.Infof("ℹ️ Track is not audio (Source: %s, Type: %s), skipping transcription", track.Source, track.Type)
		}
	}

	mp.logger.Infof("✅ track_unpublished event processed successfully (Track: %s, Room: %s, Participant: %s)",
		trackSID, roomSID, participantSID)
	return nil
}

// handleParticipantLeft обрабатывает событие выхода участника из комнаты
// Обновляет статус участника, сохраняет причину отключения и время выхода
func (mp *ManagingPortal) handleParticipantLeft(req models.WebhookRequest) error {
	mp.logger.Infof("👋 Processing participant_left event...")

	if req.Participant == nil {
		mp.logger.Errorf("❌ Participant data is missing in webhook request")
		return fmt.Errorf("participant data is missing")
	}

	var participantSID string
	if sid, ok := req.Participant["sid"].(string); ok {
		participantSID = sid
		mp.logger.Infof("  📌 Participant SID: %s", sid)
	} else {
		mp.logger.Errorf("❌ Participant SID is missing")
		return fmt.Errorf("participant SID is missing")
	}

	// Extract identity for logging
	if identity, ok := req.Participant["identity"].(string); ok {
		mp.logger.Infof("  📌 Identity: %s", identity)
	}

	// Extract room info for logging
	if req.Room != nil {
		if roomSID, ok := req.Room["sid"].(string); ok {
			mp.logger.Infof("  📌 Room SID: %s", roomSID)
		}
	}

	var disconnectReason string
	if reason, ok := req.Participant["disconnectReason"].(string); ok {
		disconnectReason = reason
		mp.logger.Infof("  📌 Disconnect Reason: %s", reason)
	}

	mp.logger.Infof("💾 Updating participant status in database...")
	if err := mp.liveKitRepo.UpdateParticipantLeft(participantSID, disconnectReason); err != nil {
		mp.logger.Errorf("❌ Failed to update participant: %v", err)
		return err
	}

	mp.logger.Infof("✅ participant_left event processed successfully (Participant: %s, Reason: %s)",
		participantSID, disconnectReason)
	return nil
}

// handleRoomFinished обрабатывает событие завершения комнаты
// Останавливает все записи egress, обновляет статус встречи, помечает комнату как завершённую
func (mp *ManagingPortal) handleRoomFinished(req models.WebhookRequest) error {
	mp.logger.Infof("🏁 Processing room_finished event...")

	if req.Room == nil {
		mp.logger.Errorf("❌ Room data is missing in webhook request")
		return fmt.Errorf("room data is missing")
	}

	var roomSID string
	if sid, ok := req.Room["sid"].(string); ok {
		roomSID = sid
		mp.logger.Infof("  📌 Room SID: %s", sid)
	} else {
		mp.logger.Errorf("❌ Room SID is missing")
		return fmt.Errorf("room SID is missing")
	}

	// Get room to find egress ID
	room, err := mp.liveKitRepo.GetRoomBySID(roomSID)
	if err != nil {
		mp.logger.Errorf("Failed to get room for finish: %v", err)
	} else if room.EgressID != "" {
		// Stop room egress recording
		mp.logger.Infof("🛑 Stopping room egress: %s", room.EgressID)
		if err := mp.stopEgress(room.EgressID); err != nil {
			mp.logger.Errorf("❌ Failed to stop room egress %s: %v", room.EgressID, err)
		} else {
			mp.logger.Infof("✅ Stopped room egress: %s", room.EgressID)
		}
	}

	// Also stop all track egress for this room
	mp.logger.Infof("🔍 Looking for track egress to stop...")
	tracks, err := mp.liveKitRepo.ListTracksByRoom(roomSID)
	if err != nil {
		mp.logger.Errorf("Failed to get tracks for room finish: %v", err)
	} else {
		mp.logger.Infof("Found %d tracks in room", len(tracks))
		for _, track := range tracks {
			if track.EgressID != "" {
				mp.logger.Infof("🛑 Stopping track egress: %s (track: %s)", track.EgressID, track.SID)
				if err := mp.stopEgress(track.EgressID); err != nil {
					mp.logger.Errorf("❌ Failed to stop track egress %s: %v", track.EgressID, err)
				} else {
					mp.logger.Infof("✅ Stopped track egress: %s", track.EgressID)
				}
			}
		}
	}

	// Room settings are no longer cached in memory, they are read from database
	mp.logger.Infof("🗑️ Room finished: %s", roomSID)

	mp.logger.Infof("💾 Marking room as finished in database...")
	if err := mp.liveKitRepo.FinishRoom(roomSID); err != nil {
		mp.logger.Errorf("❌ Failed to finish room: %v", err)
		return err
	}

	// Get the room to find the meeting ID
	mp.logger.Infof("📋 Looking up room to find associated meeting...")
	var dbRoom models.Room
	var meeting *models.Meeting

	if err := mp.db.DB.Where("sid = ?", roomSID).First(&dbRoom).Error; err != nil {
		mp.logger.Errorf("❌ Failed to find room by SID: %v", err)
	} else if dbRoom.MeetingID != nil {
		mp.logger.Infof("📅 Room is associated with meeting: %s", *dbRoom.MeetingID)

		// Update meeting status and recording flags
		meeting, err = mp.meetingRepo.GetMeetingByID(*dbRoom.MeetingID)
		if err != nil {
			mp.logger.Errorf("❌ Failed to get meeting %s: %v", *dbRoom.MeetingID, err)
		} else {
			mp.logger.Infof("🔄 Updating meeting recording flags...")

			// Only set status to completed if it's not a permanent meeting
			if !meeting.IsPermanent {
				mp.logger.Infof("  Setting meeting status to completed (not permanent)")
				meeting.Status = "completed"
			} else {
				mp.logger.Infof("  Keeping meeting status unchanged (permanent meeting)")
			}
			meeting.IsRecording = false
			meeting.IsTranscribing = false

			if err := mp.meetingRepo.UpdateMeeting(meeting); err != nil {
				mp.logger.Errorf("❌ Failed to update meeting status: %v", err)
			} else {
				if meeting.IsPermanent {
					mp.logger.Infof("✅ Meeting %s recording stopped (is_recording=false, is_transcribing=false, permanent meeting)", *dbRoom.MeetingID)
				} else {
					mp.logger.Infof("✅ Meeting %s marked as completed (is_recording=false, is_transcribing=false)", *dbRoom.MeetingID)
				}
			}
		}
	} else {
		mp.logger.Infof("ℹ️ Room is not associated with any meeting")
	}

	// Note: Composite video generation is handled by VideoPostProcessor
	// after all tracks are transcribed. See transcription_consumer.go
	mp.logger.Infof("ℹ️ Composite video will be generated by VideoPostProcessor after transcription completes")

	mp.logger.Infof("✅ room_finished event processed successfully")
	return nil
}

// handleEgressStarted обрабатывает событие начала записи egress
func (mp *ManagingPortal) handleEgressStarted(req models.WebhookRequest) error {
	mp.logger.Infof("🎬 Processing egress_started event...")

	egressInfo, ok := req.EgressInfo.(map[string]interface{})
	if !ok {
		mp.logger.Errorf("❌ Invalid egress_info in webhook")
		return fmt.Errorf("invalid egress_info")
	}

	egressID, _ := egressInfo["egress_id"].(string)
	roomName, _ := egressInfo["room_name"].(string)
	status, _ := egressInfo["status"].(string)

	mp.logger.Infof("📌 Egress ID: %s", egressID)
	mp.logger.Infof("📌 Room Name: %s", roomName)
	mp.logger.Infof("📌 Status: %s", status)

	// Обновляем статус записи в таблице EgressRecording
	if egressID != "" {
		err := mp.db.DB.Model(&models.EgressRecording{}).
			Where("egress_id = ?", egressID).
			Update("status", "active").Error
		if err != nil {
			mp.logger.Errorf("❌ Failed to update egress recording status: %v", err)
		} else {
			mp.logger.Infof("✅ Egress recording status updated to 'active': %s", egressID)
		}
	}

	mp.logger.Infof("✅ egress_started event processed successfully")
	return nil
}

// handleEgressUpdated обрабатывает событие обновления статуса egress
func (mp *ManagingPortal) handleEgressUpdated(req models.WebhookRequest) error {
	mp.logger.Infof("📊 Processing egress_updated event...")

	egressInfo, ok := req.EgressInfo.(map[string]interface{})
	if !ok {
		mp.logger.Errorf("❌ Invalid egress_info in webhook")
		return fmt.Errorf("invalid egress_info")
	}

	egressID, _ := egressInfo["egress_id"].(string)
	status, _ := egressInfo["status"].(string)
	errorStr, _ := egressInfo["error"].(string)

	mp.logger.Infof("📌 Egress ID: %s", egressID)
	mp.logger.Infof("📌 Status: %s", status)
	if errorStr != "" {
		mp.logger.Infof("📌 Error: %s", errorStr)
	}

	// Обновляем статус и сообщение об ошибке в таблице EgressRecording
	if egressID != "" && status != "" {
		updates := map[string]interface{}{
			"status": status,
		}
		if errorStr != "" {
			updates["error_message"] = errorStr
		}

		err := mp.db.DB.Model(&models.EgressRecording{}).
			Where("egress_id = ?", egressID).
			Updates(updates).Error
		if err != nil {
			mp.logger.Errorf("❌ Failed to update egress recording: %v", err)
		} else {
			mp.logger.Infof("✅ Egress recording updated: %s (status: %s)", egressID, status)
		}
	}

	mp.logger.Infof("✅ egress_updated event processed successfully")
	return nil
}

// handleEgressEnded обрабатывает событие завершения записи egress
func (mp *ManagingPortal) handleEgressEnded(req models.WebhookRequest) error {
	mp.logger.Infof("🏁 Processing egress_ended event...")

	egressInfo, ok := req.EgressInfo.(map[string]interface{})
	if !ok {
		mp.logger.Errorf("❌ Invalid egress_info in webhook")
		return fmt.Errorf("invalid egress_info")
	}

	// Try both camelCase (new format) and snake_case (old format) for egress_id
	egressID, _ := egressInfo["egressId"].(string)
	if egressID == "" {
		egressID, _ = egressInfo["egress_id"].(string)
	}

	// Try both camelCase (new format) and snake_case (old format) for room_name
	roomName, _ := egressInfo["roomName"].(string)
	if roomName == "" {
		roomName, _ = egressInfo["room_name"].(string)
	}

	status, _ := egressInfo["status"].(string)
	errorStr, _ := egressInfo["error"].(string)

	// Extract file paths from segments (new format), file_results, or file (old format)
	var filePath string

	// Try segments.playlistName (new format for track egress)
	if segments, ok := egressInfo["segments"].(map[string]interface{}); ok {
		if playlistName, ok := segments["playlistName"].(string); ok {
			filePath = playlistName
		}
	}

	// Fallback to file_results or file (old format)
	if filePath == "" {
		if fileResults, ok := egressInfo["file_results"].([]interface{}); ok && len(fileResults) > 0 {
			if fileResult, ok := fileResults[0].(map[string]interface{}); ok {
				if fp, ok := fileResult["filename"].(string); ok {
					filePath = fp
				}
			}
		} else if file, ok := egressInfo["file"].(map[string]interface{}); ok {
			if fp, ok := file["filename"].(string); ok {
				filePath = fp
			}
		}
	}

	mp.logger.Infof("📌 Egress ID: %s", egressID)
	mp.logger.Infof("📌 Room Name: %s", roomName)
	mp.logger.Infof("📌 Status: %s", status)
	if filePath != "" {
		mp.logger.Infof("📌 File Path: %s", filePath)
	}
	if errorStr != "" {
		mp.logger.Infof("📌 Error: %s", errorStr)
	}

	// Debug: Check conditions for transcription
	mp.logger.Infof("🔍 Transcription check: egressID=%v, status=%v (not failed=%v), filePath=%v, rabbitmq=%v",
		egressID != "", status, status != "failed", filePath != "", mp.rabbitMQPublisher != nil)

	// Обновляем статус завершения записи, путь к файлу и время окончания в таблице EgressRecording
	if egressID != "" {
		now := time.Now()
		updates := map[string]interface{}{
			"status":   "ended",
			"ended_at": now,
		}
		if filePath != "" {
			updates["file_path"] = filePath
		}
		if errorStr != "" {
			updates["error_message"] = errorStr
			updates["status"] = "failed"
		}

		err := mp.db.DB.Model(&models.EgressRecording{}).
			Where("egress_id = ?", egressID).
			Updates(updates).Error
		if err != nil {
			mp.logger.Errorf("❌ Failed to update egress recording: %v", err)
		} else {
			mp.logger.Infof("✅ Egress recording ended: %s (status: %s, file: %s)", egressID, updates["status"], filePath)
		}
	}

	// Send transcription task to RabbitMQ if this is a track recording that completed successfully
	if egressID != "" && status != "failed" && filePath != "" {
		// Check if RabbitMQ publisher is available, try to reconnect if not
		if mp.rabbitMQPublisher == nil {
			mp.logger.Infof("⚠️ RabbitMQ publisher not initialized, attempting to reconnect...")
			rabbitMQHost := os.Getenv("RABBITMQ_HOST")
			if rabbitMQHost == "" {
				rabbitMQHost = "localhost"
			}
			rabbitMQPort := 5672
			if portStr := os.Getenv("RABBITMQ_PORT"); portStr != "" {
				if port, err := strconv.Atoi(portStr); err == nil {
					rabbitMQPort = port
				}
			}
			rabbitMQUser := os.Getenv("RABBITMQ_USER")
			if rabbitMQUser == "" {
				rabbitMQUser = "guest"
			}
			rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
			if rabbitMQPassword == "" {
				rabbitMQPassword = "guest"
			}
			rabbitMQQueue := os.Getenv("RABBITMQ_QUEUE")
			if rabbitMQQueue == "" {
				rabbitMQQueue = "transcription_queue"
			}

			publisher, err := rabbitmq.NewPublisher(
				rabbitMQHost,
				rabbitMQPort,
				rabbitMQUser,
				rabbitMQPassword,
				rabbitMQQueue,
			)
			if err != nil {
				mp.logger.Errorf("❌ Failed to reconnect to RabbitMQ: %v", err)
			} else {
				mp.logger.Infof("✅ Successfully reconnected to RabbitMQ!")
				mp.rabbitMQPublisher = publisher
			}
		}

		if mp.rabbitMQPublisher != nil {
			mp.logger.Infof("🔍 Checking if egress %s is associated with a track...", egressID)
		// Check if this egress is associated with a track
		var track models.Track
		err := mp.db.DB.Where("egress_id = ?", egressID).First(&track).Error
		if err == nil {
			mp.logger.Infof("✅ Found track: SID=%s, Type=%s, Source=%s, MimeType=%s",
				track.SID, track.Type, track.Source, track.MimeType)
			isAudioTrack := strings.EqualFold(track.Type, "audio") ||
				strings.EqualFold(track.Source, "microphone") ||
				strings.EqualFold(track.Source, "screen_share_audio") ||
				(track.MimeType != "" && strings.HasPrefix(track.MimeType, "audio/"))
			mp.logger.Infof("🔍 Is audio track: %v", isAudioTrack)
			if !isAudioTrack {
				mp.logger.Infof("ℹ️ Egress %s belongs to non-audio track %s, skipping transcription task", egressID, track.SID)
			} else {
				// Found a track with this egress ID - send transcription task
				mp.logger.Infof("📝 Audio track confirmed! Checking meeting transcription settings...")
				mp.logger.Infof("📝 Sending transcription task for track %s (egress: %s)", track.SID, egressID)

				// Build audio URL from file path
				// File path format from LiveKit: "recontext/<egress_id>/..." or just "<egress_id>/..."
				// For track recordings, LiveKit creates m3u8 playlists
				storageURL := os.Getenv("MINIO_ENDPOINT")
				if storageURL == "" {
					storageURL = "minio:9000"
				}
				bucket := os.Getenv("MINIO_BUCKET")
				if bucket == "" {
					bucket = "recontext"
				}

				// Construct URL to m3u8 playlist
				// LiveKit stores track recordings in: <bucket>/<egress_id>/playlist.m3u8
				var audioURL string
				if strings.HasSuffix(filePath, ".m3u8") {
					// File path already points to playlist
					audioURL = fmt.Sprintf("http://%s/%s/%s", storageURL, bucket, filePath)
				} else if strings.Contains(filePath, "/") {
					// File path is a directory or file path, append playlist.m3u8
					audioURL = fmt.Sprintf("http://%s/%s/%s/playlist.m3u8", storageURL, bucket, strings.TrimSuffix(filePath, "/"))
				} else {
					// File path is just egress ID
					audioURL = fmt.Sprintf("http://%s/%s/%s/playlist.m3u8", storageURL, bucket, filePath)
				}

				mp.logger.Infof("📌 Constructed audio URL: %s", audioURL)

				// Get room to find user_id
				var room models.Room
				err := mp.db.DB.Where("sid = ?", track.RoomSID).First(&room).Error
				if err != nil {
					mp.logger.Errorf("❌ Failed to get room for track %s: %v", track.SID, err)
				} else {
					mp.logger.Infof("✅ Found room: SID=%s, Name=%s", room.SID, room.Name)
					// Get meeting to find user_id and check if transcription is enabled
					var meeting models.Meeting
					err := mp.db.DB.Where("id = ?", room.Name).First(&meeting).Error
					if err != nil {
						mp.logger.Errorf("❌ Failed to get meeting for room %s: %v", room.Name, err)
					} else if !meeting.NeedsTranscription {
						mp.logger.Infof("ℹ️ Transcription is disabled for meeting %s (needs_transcription=%v), skipping transcription task",
							meeting.ID, meeting.NeedsTranscription)
					} else {
						mp.logger.Infof("✅ Transcription is ENABLED for meeting %s!", meeting.ID)
						// Parse track ID as UUID
						trackUUID, err := uuid.Parse(track.ID.String())
						if err != nil {
							mp.logger.Errorf("❌ Invalid track ID format: %v", err)
						} else {
							// Log the complete task details before sending
							mp.logger.Infof("📋 ========================================")
							mp.logger.Infof("📋 TRANSCRIPTION TASK DETAILS:")
							mp.logger.Infof("📋 ========================================")
							mp.logger.Infof("📋 Track ID:       %s", trackUUID.String())
							mp.logger.Infof("📋 Track SID:      %s", track.SID)
							mp.logger.Infof("📋 User ID:        %s", meeting.CreatedBy.String())
							mp.logger.Infof("📋 Audio URL:      %s", audioURL)
							mp.logger.Infof("📋 Room SID:       %s", room.SID)
							mp.logger.Infof("📋 Room Name:      %s", room.Name)
							mp.logger.Infof("📋 Meeting ID:     %s", meeting.ID.String())
							mp.logger.Infof("📋 Meeting Title:  %s", meeting.Title)
							mp.logger.Infof("📋 Egress ID:      %s", egressID)
							mp.logger.Infof("📋 File Path:      %s", filePath)
							mp.logger.Infof("📋 Language:       auto-detect")
							mp.logger.Infof("📋 ========================================")

							// Send message to RabbitMQ
							err := mp.rabbitMQPublisher.PublishTranscriptionTask(
								trackUUID,
								meeting.CreatedBy,
								audioURL,
								"", // Auto-detect language
								"", // No auth token needed for MinIO access from transcription service
							)
							if err != nil {
								mp.logger.Errorf("❌ Failed to send transcription task to RabbitMQ: %v", err)
								mp.logger.Errorf("❌ Task that failed: track_id=%s, user_id=%s, audio_url=%s",
									trackUUID.String(), meeting.CreatedBy.String(), audioURL)
							} else {
								mp.logger.Infof("✅ ========================================")
								mp.logger.Infof("✅ Transcription task SUCCESSFULLY sent to RabbitMQ!")
								mp.logger.Infof("✅ Track: %s (SID: %s)", trackUUID.String(), track.SID)
								mp.logger.Infof("✅ Queue: transcription_queue")
								mp.logger.Infof("✅ Message format: {track_id, user_id, audio_url}")
								mp.logger.Infof("✅ ========================================")
							}
						}
					}
				}
			}
		} else {
			// Not a track egress - this is expected
			// Composite video is created by VideoPostProcessor after all tracks are transcribed
			mp.logger.Infof("ℹ️ Egress %s is not associated with a track - this is expected for screen share or other non-audio tracks", egressID)
		}
		} else {
			mp.logger.Infof("⚠️ Skipping transcription - RabbitMQ publisher still not available after reconnection attempt")
		}
	} else {
		mp.logger.Infof("⚠️ Skipping transcription check - conditions not met: egressID=%v, status=%v, filePath=%v",
			egressID != "", status, filePath != "")
	}

	mp.logger.Infof("✅ egress_ended event processed successfully")
	return nil
}

// GetLiveKitRooms godoc
// @Summary Получить комнаты LiveKit
// @Description Получить список комнат LiveKit с дополнительным фильтром по статусу
// @Tags LiveKit
// @Produce json
// @Param status query string false "Filter by status (active, finished)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/livekit/rooms [get]
func (mp *ManagingPortal) listLiveKitRoomsHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	rooms, err := mp.liveKitRepo.ListRooms(status, limit, offset)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list rooms", err.Error())
		return
	}

	// Return standardized response format
	response := map[string]interface{}{
		"items":     rooms,
		"total":     len(rooms),
		"offset":    offset,
		"page_size": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLiveKitRoom godoc
// @Summary Получить детали комнаты LiveKit
// @Description Получить детальную информацию о конкретной комнате
// @Tags LiveKit
// @Produce json
// @Param sid path string true "Room SID"
// @Success 200 {object} models.Room
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/livekit/rooms/{sid} [get]
func (mp *ManagingPortal) getLiveKitRoomHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	if sid == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	room, err := mp.liveKitRepo.GetRoomBySID(sid)
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Room not found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

// GetLiveKitParticipants godoc
// @Summary Получить участников комнаты
// @Description Получить всех участников для конкретной комнаты
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/livekit/participants [get]
func (mp *ManagingPortal) listLiveKitParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	roomSID := r.URL.Query().Get("room_sid")
	if roomSID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	participants, err := mp.liveKitRepo.ListParticipantsByRoom(roomSID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list participants", err.Error())
		return
	}

	// Return standardized response format
	response := map[string]interface{}{
		"items":     participants,
		"total":     len(participants),
		"offset":    0,
		"page_size": len(participants),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLiveKitTracks godoc
// @Summary Получить треки комнаты
// @Description Получить все треки для конкретной комнаты
// @Tags LiveKit
// @Produce json
// @Param room_sid query string true "Room SID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/livekit/tracks [get]
func (mp *ManagingPortal) listLiveKitTracksHandler(w http.ResponseWriter, r *http.Request) {
	roomSID := r.URL.Query().Get("room_sid")
	if roomSID == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Room SID is required", "")
		return
	}

	tracks, err := mp.liveKitRepo.ListTracksByRoom(roomSID)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list tracks", err.Error())
		return
	}

	// Return standardized response format
	response := map[string]interface{}{
		"items":     tracks,
		"total":     len(tracks),
		"offset":    0,
		"page_size": len(tracks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetWebhookEvents godoc
// @Summary Получить логи событий webhook
// @Description Получить список событий webhook с дополнительными фильтрами
// @Tags LiveKit
// @Produce json
// @Param event_type query string false "Filter by event type"
// @Param room_sid query string false "Filter by room SID"
// @Param limit query int false "Limit results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/livekit/webhook-events [get]
func (mp *ManagingPortal) listWebhookEventsHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("event_type")
	roomSID := r.URL.Query().Get("room_sid")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	events, err := mp.liveKitRepo.GetWebhookEvents(eventType, roomSID, limit, offset)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list webhook events", err.Error())
		return
	}

	// Return standardized response format
	response := map[string]interface{}{
		"items":     events,
		"total":     len(events),
		"offset":    offset,
		"page_size": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
