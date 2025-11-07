import React from 'react';
import { useTranslation } from 'react-i18next';
import { FileUpload } from '../components/FileUpload';
import { FilesList } from '../components/FilesList';

export const Documents: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.documents')}</h1>
      <p className="page-subtitle">Upload and transcribe audio/video files</p>

      <FileUpload />
      <FilesList />
    </div>
  );
};

export default Documents;
