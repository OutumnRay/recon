import { useEffect, useRef } from 'react';
import Hls from 'hls.js';
import './HLSPlayer.css';

interface HLSPlayerProps {
  src: string;
  autoplay?: boolean;
  audioOnly?: boolean;
  className?: string;
}

export default function HLSPlayer({ src, autoplay = false, audioOnly = false, className }: HLSPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const audioRef = useRef<HTMLAudioElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  useEffect(() => {
    const mediaElement = audioOnly ? audioRef.current : videoRef.current;
    if (!mediaElement) return;

    const video = mediaElement;

    // Convert relative URL to absolute if needed
    const absoluteSrc = src.startsWith('http') || src.startsWith('blob:')
      ? src
      : `${window.location.origin}${src}`;

    // Check if HLS is supported
    if (Hls.isSupported()) {
      // Get JWT token from localStorage or sessionStorage
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');

      // Initialize HLS.js
      const hls = new Hls({
        enableWorker: true,
        lowLatencyMode: false,
        backBufferLength: 90,
        xhrSetup: (xhr: XMLHttpRequest) => {
          // Add Authorization header to all HLS requests
          if (token) {
            xhr.setRequestHeader('Authorization', `Bearer ${token}`);
          }
        },
      });

      hlsRef.current = hls;

      // Bind video element
      hls.attachMedia(video);

      // Wait for manifest to be parsed
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        console.log('HLS manifest parsed, ready to play');
        if (autoplay) {
          video.play().catch(err => console.error('Autoplay failed:', err));
        }
      });

      // Load source with absolute URL
      hls.loadSource(absoluteSrc);

      // Handle errors
      hls.on(Hls.Events.ERROR, (_event, data) => {
        console.error('HLS error:', data);
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              console.error('Fatal network error, trying to recover...');
              hls.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              console.error('Fatal media error, trying to recover...');
              hls.recoverMediaError();
              break;
            default:
              console.error('Fatal error, cannot recover');
              hls.destroy();
              break;
          }
        }
      });

      // Cleanup on unmount
      return () => {
        if (hlsRef.current) {
          hlsRef.current.destroy();
          hlsRef.current = null;
        }
      };
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS support (Safari, iOS)
      console.log('Using native HLS support');
      video.src = absoluteSrc;
      if (autoplay) {
        video.play().catch(err => console.error('Autoplay failed:', err));
      }

      return () => {
        video.src = '';
      };
    } else {
      console.error('HLS is not supported in this browser');
    }
  }, [src, autoplay, audioOnly]);

  return (
    <div className={`hls-player ${className || ''}`}>
      {audioOnly ? (
        <audio
          ref={audioRef}
          controls
          className="hls-audio"
        />
      ) : (
        <video
          ref={videoRef}
          controls
          playsInline
          className="hls-video"
        />
      )}
    </div>
  );
}
