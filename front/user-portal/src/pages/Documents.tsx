import React, { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { FileUpload } from '../components/FileUpload';
import { FilesList } from '../components/FilesList';

export const Documents: React.FC = () => {
  const { t } = useTranslation();

  // Set page title
  useEffect(() => {
    document.title = `Recontext - ${t('nav.documents')}`;
  }, [t]);

  return (
    <div className="page-container">
      <h1 className="page-title">{t('documents.title')}</h1>
      <p className="page-subtitle">{t('documents.subtitle')}</p>

      <FileUpload />
      <FilesList />
    </div>
  );
};

export default Documents;
