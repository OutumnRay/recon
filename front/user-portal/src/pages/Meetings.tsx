import React from 'react';
import { useTranslation } from 'react-i18next';

export const Meetings: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.meetings')}</h1>
      <p className="page-subtitle">View and manage your recorded meetings</p>

      <div className="empty-state">
        <h2 className="empty-title">No meetings yet</h2>
        <p className="empty-description">
          Your recorded meetings will appear here. Start a meeting to see it in this list.
        </p>
      </div>
    </div>
  );
};

export default Meetings;
