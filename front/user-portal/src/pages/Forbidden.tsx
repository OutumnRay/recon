import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuShieldAlert, LuLayoutDashboard, LuLogIn } from 'react-icons/lu';
import './Forbidden.css';

export default function Forbidden() {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const handleGoHome = () => {
    navigate('/dashboard/meetings');
  };

  const handleGoLogin = () => {
    navigate('/login');
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  // Check if user is logged in
  const isLoggedIn = !!(localStorage.getItem('token') || sessionStorage.getItem('token'));

  return (
    <div className="forbidden-page">
      <div className="forbidden-container">
        <div className="forbidden-header">
          <div className="forbidden-icon">
            <LuShieldAlert />
          </div>
          <h1>403</h1>
          <h2>{t('forbidden.title')}</h2>
          <p className="forbidden-description">
            {t('forbidden.description')}
          </p>
        </div>

        <div className="forbidden-actions">
          <button onClick={handleGoBack} className="btn-secondary">
            {t('forbidden.goBack')}
          </button>
          {isLoggedIn ? (
            <button onClick={handleGoHome} className="btn-primary">
              <LuLayoutDashboard className="btn-icon" />
              {t('forbidden.goHome')}
            </button>
          ) : (
            <button onClick={handleGoLogin} className="btn-primary">
              <LuLogIn className="btn-icon" />
              {t('forbidden.goLogin')}
            </button>
          )}
        </div>

        <div className="forbidden-footer">
          <p>{t('forbidden.footer')}</p>
        </div>
      </div>
    </div>
  );
}
