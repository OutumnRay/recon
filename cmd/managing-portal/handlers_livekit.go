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

	// Process event based on type
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
	default:
		mp.logger.Errorf("Unknown webhook event type: %s", webhookReq.Event)
		mp.respondWithError(w, http.StatusBadRequest, "Unknown event type", "")
		return
	}

	if err != nil {
		mp.logger.Errorf("Failed to process %s event: %s", webhookReq.Event, err.Error())
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to process event", err.Error())
		return
	}

	// Return success response
	response := models.WebhookResponse{
		Status:  "success",
		Message: fmt.Sprintf("Event %s processed successfully", webhookReq.Event),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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

	// Save room to database
	mp.logger.Infof("💾 Saving room to database...")
	if err := mp.liveKitRepo.CreateRoom(room); err != nil {
		mp.logger.Errorf("❌ Failed to save room to database: %v", err)
		return err
	}
	mp.logger.Infof("✅ Room saved successfully (DB ID: %s, SID: %s)", room.ID, room.SID)

	// Link room to meeting if room.Name is a valid UUID (meeting ID)
	var needsVideoRecord bool
	var needsAudioRecord bool
	var needsTranscription bool

	if room.Name != "" {
		mp.logger.Infof("Checking if room name '%s' is a meeting ID...", room.Name)
		if meetingID, err := uuid.Parse(room.Name); err == nil {
			mp.logger.Infof("Room name is valid UUID, looking up meeting %s", meetingID)
			// Find meeting by ID and update its LiveKitRoomID
			meeting, err := mp.meetingRepo.GetMeetingByID(meetingID)
			if err == nil {
				mp.logger.Infof("Found meeting %s: NeedsVideoRecord=%v, NeedsAudioRecord=%v, NeedsTranscription=%v",
					meetingID, meeting.NeedsVideoRecord, meeting.NeedsAudioRecord, meeting.NeedsTranscription)

				// Link meeting to the room ID (not SID - SID is a string like RM_xxx)
				meeting.LiveKitRoomID = &room.ID
				if err := mp.meetingRepo.UpdateMeeting(meeting); err != nil {
					mp.logger.Errorf("Failed to link meeting %s to room %s (SID: %s): %v", meetingID, room.ID, room.SID, err)
				} else {
					mp.logger.Infof("Successfully linked meeting %s to LiveKit room %s (SID: %s)", meetingID, room.ID, room.SID)
				}

				// Get recording settings from meeting
				needsVideoRecord = meeting.NeedsVideoRecord
				needsAudioRecord = meeting.NeedsAudioRecord
				needsTranscription = meeting.NeedsTranscription

				// Если включено видео, автоматически включаем аудио
				if needsVideoRecord {
					needsAudioRecord = true
					mp.logger.Infof("Video recording enabled, automatically enabling audio recording")
				}

				// Сохраняем настройки транскрибации для этой комнаты
				mp.setRoomSettings(room.SID, &RoomSettings{
					NeedsTranscription: needsTranscription,
				})
				mp.logger.Infof("Room settings saved: NeedsTranscription=%v", needsTranscription)
			} else {
				mp.logger.Errorf("Failed to find meeting %s for room linking: %v", meetingID, err)
			}
		} else {
			mp.logger.Infof("Room name '%s' is not a valid UUID: %v", room.Name, err)
		}
	}

	// Start room composite egress recording if enabled in meeting settings
	mp.logger.Infof("📹 Checking room egress requirements...")
	mp.logger.Infof("  • Video Recording: %v", needsVideoRecord)
	mp.logger.Infof("  • Audio Recording: %v", needsAudioRecord)
	mp.logger.Infof("  • Room Name: %s", room.Name)

	if room.Name != "" && (needsVideoRecord || needsAudioRecord) {
		// If video is not needed, record audio only
		audioOnly := !needsVideoRecord && needsAudioRecord

		mp.logger.Infof("🚀 Starting room composite egress for room '%s' (audioOnly=%v)...", room.Name, audioOnly)
		egressID, err := mp.startRoomCompositeEgress(room.Name, audioOnly)
		if err != nil {
			mp.logger.Errorf("❌ Failed to start room composite egress: %v", err)
		} else if egressID != "" {
			mp.logger.Infof("✅ Room egress started successfully: %s (audioOnly=%v)", egressID, audioOnly)
			// Update room with egress ID
			room.EgressID = egressID
			if err := mp.liveKitRepo.CreateRoom(room); err != nil {
				mp.logger.Errorf("❌ Failed to update room with egress ID: %v", err)
			} else {
				mp.logger.Infof("✅ Room updated with egress ID")
			}
		} else {
			mp.logger.Infof("⚠️ Room egress not started (disabled in config or returned empty ID)")
		}
	} else {
		mp.logger.Infof("ℹ️ Room egress skipped:")
		mp.logger.Infof("  • Room Name: '%s' (empty=%v)", room.Name, room.Name == "")
		mp.logger.Infof("  • Video needed: %v", needsVideoRecord)
		mp.logger.Infof("  • Audio needed: %v", needsAudioRecord)
		mp.logger.Infof("  • Any recording needed: %v", needsVideoRecord || needsAudioRecord)
	}

	mp.logger.Infof("✅ room_started event processed successfully")
	return nil
}

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
	if req.Room != nil {
		if roomSID, ok := req.Room["sid"].(string); ok {
			participant.RoomSID = roomSID
			mp.logger.Infof("  📌 Room SID: %s", roomSID)
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

	// Проверяем настройки транскрибации для комнаты
	roomSettings := mp.getRoomSettings(track.RoomSID)
	needsTranscription := roomSettings != nil && roomSettings.NeedsTranscription

	mp.logger.Infof("📝 Room transcription settings: NeedsTranscription=%v (RoomSID=%s)", needsTranscription, track.RoomSID)

	// Start track egress recording for audio tracks if transcription is enabled
	isAudioTrack := track.Source == "MICROPHONE" || (track.Source == "SCREEN_SHARE_AUDIO") ||
		(track.MimeType != "" && strings.HasPrefix(track.MimeType, "audio/"))

	if isAudioTrack && needsTranscription {
		mp.logger.Infof("🎤 Audio track detected with transcription enabled - checking egress eligibility...")
		mp.logger.Infof("  ✓ Track Source: %s", track.Source)
		mp.logger.Infof("  ✓ MIME Type: %s", track.MimeType)
		mp.logger.Infof("  ✓ Room Name: %s (empty=%v)", roomName, roomName == "")
		mp.logger.Infof("  ✓ Track SID: %s (empty=%v)", track.SID, track.SID == "")
		mp.logger.Infof("  ✓ Transcription needed: %v", needsTranscription)

		if roomName != "" && track.SID != "" {
			mp.logger.Infof("🚀 Starting track composite egress for audio track %s in room '%s'...", track.SID, roomName)
			egressID, err := mp.startTrackCompositeEgress(roomName, track.SID)
			if err != nil {
				mp.logger.Errorf("❌ Failed to start track composite egress: %v", err)
			} else if egressID != "" {
				mp.logger.Infof("✅ Track egress started successfully: %s (track: %s)", egressID, track.SID)
				// Update track with egress ID
				track.EgressID = egressID
				if err := mp.liveKitRepo.CreateTrack(track); err != nil {
					mp.logger.Errorf("❌ Failed to update track with egress ID: %v", err)
				} else {
					mp.logger.Infof("✅ Track updated with egress ID")
				}
			} else {
				mp.logger.Infof("⚠️ Track egress not started (disabled in config or returned empty ID)")
			}
		} else {
			mp.logger.Infof("⚠️ Track egress not started - missing required data:")
			mp.logger.Infof("  • Room Name empty: %v", roomName == "")
			mp.logger.Infof("  • Track SID empty: %v", track.SID == "")
		}
	} else {
		if !isAudioTrack {
			mp.logger.Infof("ℹ️ Track egress skipped - not an audio track (Source=%s, MimeType=%s)", track.Source, track.MimeType)
		} else {
			mp.logger.Infof("ℹ️ Track egress skipped - transcription not enabled for this room (NeedsTranscription=%v)", needsTranscription)
		}
	}

	mp.logger.Infof("✅ track_published event processed successfully")
	return nil
}

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

	mp.logger.Infof("✅ track_unpublished event processed successfully (Track: %s, Room: %s, Participant: %s)",
		trackSID, roomSID, participantSID)
	return nil
}

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

	// Удаляем настройки комнаты из памяти
	mp.deleteRoomSettings(roomSID)
	mp.logger.Infof("🗑️ Room settings deleted for room: %s", roomSID)

	mp.logger.Infof("💾 Marking room as finished in database...")
	if err := mp.liveKitRepo.FinishRoom(roomSID); err != nil {
		mp.logger.Errorf("❌ Failed to finish room: %v", err)
		return err
	}

	mp.logger.Infof("✅ room_finished event processed successfully")
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
