import React from 'react';
import { useTranslation } from 'react-i18next';

export const Management: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.management')}</h1>
      <p className="page-subtitle">Manage your account settings and preferences</p>

      <div className="empty-state">
        <h2 className="empty-title">Management panel</h2>
        <p className="empty-description">
          Configure your account settings, manage integrations, and customize your workspace.
        </p>
      </div>
    </div>
  );
};

export default Management;
