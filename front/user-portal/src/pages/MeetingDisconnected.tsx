import { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import './MeetingDisconnected.css';

type LocationState = {
  isAnonymous?: boolean;
};

export default function MeetingDisconnected() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const state = location.state as LocationState | null;

  const isAnonymous = useMemo(() => Boolean(state?.isAnonymous), [state]);

  const handleGoHome = () => {
    if (isAnonymous) {
      window.location.href = 'https://recontext.online';
    } else {
      navigate('/dashboard/meetings', { replace: true });
    }
  };

  return (
    <div className="meeting-disconnected-page">
      <div className="meeting-disconnected-card surface-card elevated">
        <div className="meeting-disconnected-icon">⚠️</div>
        <h1>{t('meetingRoom.disconnectNotice.title')}</h1>
        <p>{t('meetingRoom.disconnectNotice.description')}</p>
        <button className="btn btn-primary" onClick={handleGoHome}>
          {t('meetingRoom.disconnectNotice.button')}
        </button>
      </div>
    </div>
  );
}
