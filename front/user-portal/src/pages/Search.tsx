import React from 'react';
import { useTranslation } from 'react-i18next';

export const Search: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.search')}</h1>
      <p className="page-subtitle">Search through your meetings and transcripts</p>

      <div className="empty-state">
        <h2 className="empty-title">Search functionality coming soon</h2>
        <p className="empty-description">
          You'll be able to search through all your meetings, transcripts, and documents to find exactly what you need.
        </p>
      </div>
    </div>
  );
};

export default Search;
