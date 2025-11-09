import express from 'express';
import { AccessToken } from 'livekit-server-sdk';
import cors from 'cors';
import dotenv from 'dotenv';

dotenv.config();

const app = express();
const port = 3000;

// Enable CORS for frontend
app.use(cors());
app.use(express.json());

// API credentials
const LIVEKIT_API_KEY = process.env.LIVEKIT_API_KEY || 'APIBj3yrXtyPRNq';
const LIVEKIT_API_SECRET = process.env.LIVEKIT_API_SECRET || '2Q66dFk7HWpxTuTneMT4fQlsxeIlmkn47ApjnJiSukiA';

/**
 * Create LiveKit access token
 * @param {string} roomName - Name of the room to join
 * @param {string} participantName - Name/identity of the participant
 * @returns {Promise<string>} JWT token
 */
const createToken = async (roomName, participantName) => {
  const at = new AccessToken(LIVEKIT_API_KEY, LIVEKIT_API_SECRET, {
    identity: participantName,
    name: participantName,
    // Token to expire after 6 hours
    ttl: '6h',
  });

  at.addGrant({
    roomJoin: true,
    room: roomName,
    canPublish: true,
    canSubscribe: true,
    canPublishData: true,
  });

  return await at.toJwt();
};

// GET endpoint for simple token generation
app.get('/getToken', async (req, res) => {
  try {
    const roomName = req.query.room || 'my-room';
    const participantName = req.query.name || 'anonymous-' + Math.random().toString(36).substring(7);

    const token = await createToken(roomName, participantName);

    res.json({
      token,
      url: process.env.LIVEKIT_URL || 'wss://video.recontext.online',
      roomName,
      participantName,
    });
  } catch (error) {
    console.error('Error generating token:', error);
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

// POST endpoint for token generation
app.post('/getToken', async (req, res) => {
  try {
    const { room, name } = req.body;

    if (!room || !name) {
      return res.status(400).json({ error: 'Room name and participant name are required' });
    }

    const token = await createToken(room, name);

    res.json({
      token,
      url: process.env.LIVEKIT_URL || 'wss://video.recontext.online',
      roomName: room,
      participantName: name,
    });
  } catch (error) {
    console.error('Error generating token:', error);
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

app.listen(port, () => {
  console.log(`LiveKit token server listening on port ${port}`);
  console.log(`API Key: ${LIVEKIT_API_KEY}`);
  console.log(`Get token: http://localhost:${port}/getToken?room=my-room&name=username`);
});
