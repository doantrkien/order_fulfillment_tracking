import React, { useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { injectGreekStyles } from '../../styles/greekTheme';

export const Header: React.FC = () => {
  const { user, logout } = useAuth();
  useEffect(() => { injectGreekStyles(); }, []);

  if (!user) return null;

  return (
    <header className="greek-navbar">
      <div style={{ flex: 1 }} />

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <span style={{
            fontFamily: "'Inter', sans-serif",
            fontSize: '13px',
            color: 'rgba(232,213,163,0.70)',
            letterSpacing: '0.02em',
          }}>
            {user.email}
          </span>
          <span className="greek-badge greek-badge-gold">{user.role}</span>
        </div>

        <button className="greek-btn-secondary" onClick={logout}>
          Logout
        </button>
      </div>
    </header>
  );
};
