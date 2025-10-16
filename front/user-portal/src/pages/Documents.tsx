import React from 'react';
import { useTranslation } from 'react-i18next';

export const Documents: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.documents')}</h1>
      <p className="page-subtitle">Access your transcripts and meeting documents</p>

      <div className="empty-state">
        <h2 className="empty-title">No documents available</h2>
        <p className="empty-description">
          Your transcripts, summaries, and other meeting documents will be stored here for easy access.
        </p>
      </div>
    </div>
  );
};

export default Documents;
