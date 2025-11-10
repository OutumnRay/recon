import React, { useEffect } from 'react';
import { useTranslation } from 'react-i18next';

export const Search: React.FC = () => {
  const { t } = useTranslation();

  // Set page title
  useEffect(() => {
    document.title = `Recontext - ${t('nav.search')}`;
  }, [t]);

  return (
    <div className="page-container">
      <h1 className="page-title">{t('search.title')}</h1>
      <p className="page-subtitle">{t('search.subtitle')}</p>

      <div className="empty-state">
        <h2 className="empty-title">{t('search.comingSoon')}</h2>
        <p className="empty-description">
          {t('search.description')}
        </p>
      </div>
    </div>
  );
};

export default Search;
