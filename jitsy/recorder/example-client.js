/**
 * Example client integration for Jitsi Kurento Recorder
 *
 * This script demonstrates how to integrate the recorder API
 * with your Jitsi Meet client application.
 */

class JitsiRecorderClient {
    constructor(apiUrl = 'http://localhost:9888') {
        this.apiUrl = apiUrl;
        this.activeRecordings = new Map();
    }

    /**
     * Start recording for a user
     * @param {string} userId - User identifier
     * @param {string} roomName - Room name
     * @returns {Promise<Object>} Recording info
     */
    async startRecording(userId, roomName) {
        try {
            const response = await fetch(
                `${this.apiUrl}/record/start?user=${userId}&room=${roomName}`,
                { method: 'POST' }
            );

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.detail || 'Failed to start recording');
            }

            const data = await response.json();
            this.activeRecordings.set(userId, {
                ...data,
                roomName
            });

            console.log(`✅ Recording started for ${userId}`, data);
            return data;

        } catch (error) {
            console.error(`❌ Failed to start recording for ${userId}:`, error);
            throw error;
        }
    }

    /**
     * Stop recording for a user
     * @param {string} userId - User identifier
     * @param {string} roomName - Room name
     * @returns {Promise<Object>} Recording info
     */
    async stopRecording(userId, roomName) {
        try {
            const response = await fetch(
                `${this.apiUrl}/record/stop?user=${userId}&room=${roomName}`,
                { method: 'POST' }
            );

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.detail || 'Failed to stop recording');
            }

            const data = await response.json();
            this.activeRecordings.delete(userId);

            console.log(`⏹️ Recording stopped for ${userId}`, data);
            return data;

        } catch (error) {
            console.error(`❌ Failed to stop recording for ${userId}:`, error);
            throw error;
        }
    }

    /**
     * Get recording status for a user
     * @param {string} userId - User identifier
     * @returns {Promise<Object>} Recording status
     */
    async getStatus(userId) {
        try {
            const response = await fetch(
                `${this.apiUrl}/record/status?user=${userId}`
            );

            if (response.status === 404) {
                return { status: 'not_found' };
            }

            if (!response.ok) {
                throw new Error('Failed to get recording status');
            }

            return await response.json();

        } catch (error) {
            console.error(`❌ Failed to get status for ${userId}:`, error);
            throw error;
        }
    }

    /**
     * Get list of all active recordings
     * @returns {Promise<Object>} List of recordings
     */
    async listRecordings() {
        try {
            const response = await fetch(`${this.apiUrl}/record/list`);

            if (!response.ok) {
                throw new Error('Failed to list recordings');
            }

            return await response.json();

        } catch (error) {
            console.error('❌ Failed to list recordings:', error);
            throw error;
        }
    }

    /**
     * Stop all active recordings
     * @returns {Promise<Object>} Stop result
     */
    async stopAll() {
        try {
            const response = await fetch(
                `${this.apiUrl}/record/stop-all`,
                { method: 'DELETE' }
            );

            if (!response.ok) {
                throw new Error('Failed to stop all recordings');
            }

            const data = await response.json();
            this.activeRecordings.clear();

            console.log('🛑 All recordings stopped', data);
            return data;

        } catch (error) {
            console.error('❌ Failed to stop all recordings:', error);
            throw error;
        }
    }

    /**
     * Check if recording is active for user
     * @param {string} userId - User identifier
     * @returns {boolean} True if recording
     */
    isRecording(userId) {
        return this.activeRecordings.has(userId);
    }
}


// ============================================================================
// Integration with Jitsi Meet API
// ============================================================================

/**
 * Initialize recorder integration with Jitsi Meet
 * @param {Object} jitsiAPI - Jitsi External API instance
 * @param {string} recorderApiUrl - Recorder API URL
 */
function initJitsiRecorder(jitsiAPI, recorderApiUrl = 'http://localhost:9888') {
    const recorder = new JitsiRecorderClient(recorderApiUrl);
    const roomName = jitsiAPI.getRoomName();

    // Get local participant ID
    let localParticipantId = null;

    jitsiAPI.addEventListener('videoConferenceJoined', (event) => {
        localParticipantId = event.id;
        console.log('📹 Joined conference:', event);
    });

    // Optional: Auto-record on join
    jitsiAPI.addEventListener('participantJoined', async (event) => {
        console.log('👤 Participant joined:', event);

        // Uncomment to auto-start recording for all participants
        // try {
        //     await recorder.startRecording(event.id, roomName);
        // } catch (error) {
        //     console.error('Failed to start auto-recording:', error);
        // }
    });

    // Optional: Auto-stop on leave
    jitsiAPI.addEventListener('participantLeft', async (event) => {
        console.log('👋 Participant left:', event);

        // Stop recording if it was active
        if (recorder.isRecording(event.id)) {
            try {
                await recorder.stopRecording(event.id, roomName);
            } catch (error) {
                console.error('Failed to stop recording on leave:', error);
            }
        }
    });

    // Add custom toolbar button for recording
    jitsiAPI.addEventListener('toolbarButtonClicked', async (event) => {
        if (event.key === 'record') {
            const isRecording = recorder.isRecording(localParticipantId);

            if (isRecording) {
                // Stop recording
                try {
                    await recorder.stopRecording(localParticipantId, roomName);
                    jitsiAPI.executeCommand('toggleRecording', { enabled: false });
                    alert('Recording stopped ⏹️');
                } catch (error) {
                    alert('Failed to stop recording: ' + error.message);
                }
            } else {
                // Start recording
                try {
                    await recorder.startRecording(localParticipantId, roomName);
                    jitsiAPI.executeCommand('toggleRecording', { enabled: true });
                    alert('Recording started 🔴');
                } catch (error) {
                    alert('Failed to start recording: ' + error.message);
                }
            }
        }
    });

    // Cleanup on leave
    jitsiAPI.addEventListener('videoConferenceLeft', async () => {
        console.log('🚪 Left conference');

        // Stop recording for local user
        if (localParticipantId && recorder.isRecording(localParticipantId)) {
            try {
                await recorder.stopRecording(localParticipantId, roomName);
            } catch (error) {
                console.error('Failed to stop recording on conference leave:', error);
            }
        }
    });

    return recorder;
}


// ============================================================================
// Example usage with Jitsi External API
// ============================================================================

/*
// HTML:
<div id="jitsi-container"></div>

// JavaScript:
const domain = 'meet.recontext.online';
const options = {
    roomName: 'MyTestRoom',
    width: '100%',
    height: 600,
    parentNode: document.querySelector('#jitsi-container'),
    configOverwrite: {
        enableRecording: true
    },
    interfaceConfigOverwrite: {
        TOOLBAR_BUTTONS: [
            'camera',
            'microphone',
            'desktop',
            'chat',
            'participants-pane',
            'record' // Custom record button
        ]
    }
};

const api = new JitsiMeetExternalAPI(domain, options);

// Initialize recorder integration
const recorder = initJitsiRecorder(api, 'http://localhost:9888');

// Manual recording control
document.getElementById('start-recording-btn').addEventListener('click', async () => {
    const userId = 'my-user-id'; // Get from your auth system
    const roomName = api.getRoomName();

    try {
        await recorder.startRecording(userId, roomName);
        console.log('Recording started!');
    } catch (error) {
        console.error('Recording failed:', error);
    }
});

document.getElementById('stop-recording-btn').addEventListener('click', async () => {
    const userId = 'my-user-id';
    const roomName = api.getRoomName();

    try {
        await recorder.stopRecording(userId, roomName);
        console.log('Recording stopped!');
    } catch (error) {
        console.error('Stop failed:', error);
    }
});

// Get status
document.getElementById('status-btn').addEventListener('click', async () => {
    const userId = 'my-user-id';

    try {
        const status = await recorder.getStatus(userId);
        console.log('Recording status:', status);
    } catch (error) {
        console.error('Status check failed:', error);
    }
});
*/


// ============================================================================
// Standalone usage (without Jitsi API)
// ============================================================================

/*
const recorder = new JitsiRecorderClient('http://localhost:9888');

// Start recording
await recorder.startRecording('user123', 'room456');

// Wait some time...
await new Promise(resolve => setTimeout(resolve, 60000));

// Stop recording
await recorder.stopRecording('user123', 'room456');

// List all recordings
const list = await recorder.listRecordings();
console.log('Active recordings:', list);
*/


// Export for Node.js/CommonJS
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { JitsiRecorderClient, initJitsiRecorder };
}

// Export for ES6 modules
if (typeof exports !== 'undefined') {
    exports.JitsiRecorderClient = JitsiRecorderClient;
    exports.initJitsiRecorder = initJitsiRecorder;
}
