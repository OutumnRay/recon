import { useEffect, useRef } from 'react';
import Hls from 'hls.js';
import './HLSPlayer.css';

interface HLSPlayerProps {
  src: string;
  autoplay?: boolean;
}

export default function HLSPlayer({ src, autoplay = false }: HLSPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  useEffect(() => {
    if (!videoRef.current) return;

    const video = videoRef.current;

    // Check if HLS is supported
    if (Hls.isSupported()) {
      // Get JWT token from localStorage
      const token = localStorage.getItem('token');

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

      // Load source
      hls.loadSource(src);

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
      video.src = src;
      if (autoplay) {
        video.play().catch(err => console.error('Autoplay failed:', err));
      }

      return () => {
        video.src = '';
      };
    } else {
      console.error('HLS is not supported in this browser');
    }
  }, [src, autoplay]);

  return (
    <div className="hls-player">
      <video
        ref={videoRef}
        controls
        playsInline
        className="hls-video"
      />
    </div>
  );
}
