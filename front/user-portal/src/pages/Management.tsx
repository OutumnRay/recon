import React from 'react';
import { useTranslation } from 'react-i18next';

export const Management: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('management.title')}</h1>
      <p className="page-subtitle">{t('management.subtitle')}</p>

      <div className="empty-state">
        <h2 className="empty-title">{t('management.panel')}</h2>
        <p className="empty-description">
          {t('management.description')}
        </p>
      </div>
    </div>
  );
};

export default Management;
