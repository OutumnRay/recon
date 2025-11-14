import { useEffect, useRef, useState } from 'react';
import { LuX, LuSettings } from 'react-icons/lu';
import { useTranslation } from 'react-i18next';
import './MediaSettingsModal.css';

interface MediaSettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onApplySettings: (videoDeviceId: string, audioDeviceId: string) => void;
  currentVideoDeviceId?: string;
  currentAudioDeviceId?: string;
}

export default function MediaSettingsModal({
  isOpen,
  onClose,
  onApplySettings,
  currentVideoDeviceId,
  currentAudioDeviceId,
}: MediaSettingsModalProps) {
  const { t } = useTranslation();
  const [videoDevices, setVideoDevices] = useState<MediaDeviceInfo[]>([]);
  const [audioDevices, setAudioDevices] = useState<MediaDeviceInfo[]>([]);
  const [selectedVideoDevice, setSelectedVideoDevice] = useState<string>(currentVideoDeviceId || '');
  const [selectedAudioDevice, setSelectedAudioDevice] = useState<string>(currentAudioDeviceId || '');
  const [audioLevel, setAudioLevel] = useState<number>(0);

  const videoPreviewRef = useRef<HTMLVideoElement>(null);
  const previewStreamRef = useRef<MediaStream | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const animationFrameRef = useRef<number | null>(null);

  // Load available devices
  useEffect(() => {
    const loadDevices = async () => {
      try {
        // Request permissions first
        await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
          .then(stream => stream.getTracks().forEach(track => track.stop()));

        const devices = await navigator.mediaDevices.enumerateDevices();
        const videoInputs = devices.filter(device => device.kind === 'videoinput');
        const audioInputs = devices.filter(device => device.kind === 'audioinput');

        setVideoDevices(videoInputs);
        setAudioDevices(audioInputs);

        // Set default devices if not already set
        if (!selectedVideoDevice && videoInputs.length > 0) {
          setSelectedVideoDevice(videoInputs[0].deviceId);
        }
        if (!selectedAudioDevice && audioInputs.length > 0) {
          setSelectedAudioDevice(audioInputs[0].deviceId);
        }
      } catch (err) {
        console.error('Failed to enumerate devices:', err);
      }
    };

    if (isOpen) {
      loadDevices();
    }
  }, [isOpen, selectedVideoDevice, selectedAudioDevice]);

  // Update preview when selected video device changes
  useEffect(() => {
    if (!isOpen || !selectedVideoDevice) return;

    const startPreview = async () => {
      try {
        // Stop previous stream
        if (previewStreamRef.current) {
          previewStreamRef.current.getTracks().forEach(track => track.stop());
        }

        // Start new stream with selected device
        const stream = await navigator.mediaDevices.getUserMedia({
          video: { deviceId: { exact: selectedVideoDevice } },
          audio: false,
        });

        previewStreamRef.current = stream;

        if (videoPreviewRef.current) {
          videoPreviewRef.current.srcObject = stream;
        }
      } catch (err) {
        console.error('Failed to start video preview:', err);
      }
    };

    startPreview();

    return () => {
      if (previewStreamRef.current) {
        previewStreamRef.current.getTracks().forEach(track => track.stop());
      }
    };
  }, [isOpen, selectedVideoDevice]);

  // Update audio level meter when selected audio device changes
  useEffect(() => {
    if (!isOpen || !selectedAudioDevice) return;

    const startAudioMonitoring = async () => {
      try {
        // Stop previous audio monitoring
        if (animationFrameRef.current) {
          cancelAnimationFrame(animationFrameRef.current);
        }
        if (audioContextRef.current) {
          audioContextRef.current.close();
        }

        // Start new audio stream with selected device
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: { deviceId: { exact: selectedAudioDevice } },
          video: false,
        });

        const audioContext = new AudioContext();
        const analyser = audioContext.createAnalyser();
        const microphone = audioContext.createMediaStreamSource(stream);

        analyser.fftSize = 256;
        analyser.smoothingTimeConstant = 0.8;
        microphone.connect(analyser);

        audioContextRef.current = audioContext;
        analyserRef.current = analyser;

        const dataArray = new Uint8Array(analyser.frequencyBinCount);

        const updateLevel = () => {
          if (!analyserRef.current) return;

          analyserRef.current.getByteFrequencyData(dataArray);
          const average = dataArray.reduce((a, b) => a + b) / dataArray.length;
          const normalizedLevel = Math.min(100, (average / 255) * 150); // Scale to 0-100
          setAudioLevel(normalizedLevel);

          animationFrameRef.current = requestAnimationFrame(updateLevel);
        };

        updateLevel();

        // Clean up the audio stream (we only need it for monitoring, not playback)
        stream.getTracks().forEach(track => {
          // Keep track alive for monitoring but don't play it
          track.enabled = true;
        });

      } catch (err) {
        console.error('Failed to start audio monitoring:', err);
      }
    };

    startAudioMonitoring();

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
      if (audioContextRef.current) {
        audioContextRef.current.close();
      }
    };
  }, [isOpen, selectedAudioDevice]);

  // Cleanup on close
  useEffect(() => {
    if (!isOpen) {
      if (previewStreamRef.current) {
        previewStreamRef.current.getTracks().forEach(track => track.stop());
        previewStreamRef.current = null;
      }
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
      if (audioContextRef.current) {
        audioContextRef.current.close();
        audioContextRef.current = null;
      }
      setAudioLevel(0);
    }
  }, [isOpen]);

  const handleApply = () => {
    onApplySettings(selectedVideoDevice, selectedAudioDevice);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="media-settings-backdrop" onClick={onClose}>
      <div className="media-settings-modal" onClick={(e) => e.stopPropagation()}>
        <div className="media-settings-header">
          <div className="media-settings-title">
            <LuSettings />
            <h2>{t('mediaSettings.title')}</h2>
          </div>
          <button className="media-settings-close" onClick={onClose}>
            <LuX />
          </button>
        </div>

        <div className="media-settings-content">
          <div className="media-settings-section">
            <label htmlFor="video-select">{t('mediaSettings.camera')}</label>
            <select
              id="video-select"
              value={selectedVideoDevice}
              onChange={(e) => setSelectedVideoDevice(e.target.value)}
            >
              {videoDevices.map((device) => (
                <option key={device.deviceId} value={device.deviceId}>
                  {device.label || t('mediaSettings.cameraFallback', { id: device.deviceId.slice(0, 8) })}
                </option>
              ))}
            </select>

            <div className="video-preview-container">
              <video
                ref={videoPreviewRef}
                autoPlay
                playsInline
                muted
                className="video-preview"
              />
            </div>
          </div>

          <div className="media-settings-section">
            <label htmlFor="audio-select">{t('mediaSettings.microphone')}</label>
            <select
              id="audio-select"
              value={selectedAudioDevice}
              onChange={(e) => setSelectedAudioDevice(e.target.value)}
            >
              {audioDevices.map((device) => (
                <option key={device.deviceId} value={device.deviceId}>
                  {device.label || t('mediaSettings.microphoneFallback', { id: device.deviceId.slice(0, 8) })}
                </option>
              ))}
            </select>

            <div className="audio-level-container">
              <div className="audio-level-meter">
                <div className="audio-level-fill" style={{ height: `${audioLevel}%` }} />
              </div>
              <span className="audio-level-label">{t('mediaSettings.audioLevel')}</span>
            </div>
          </div>
        </div>

        <div className="media-settings-footer">
          <button className="btn btn-ghost" onClick={onClose}>
            {t('mediaSettings.cancel')}
          </button>
          <button className="btn btn-primary" onClick={handleApply}>
            {t('mediaSettings.apply')}
          </button>
        </div>
      </div>
    </div>
  );
}
