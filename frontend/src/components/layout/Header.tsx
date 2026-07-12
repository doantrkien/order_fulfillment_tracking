import React, { useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { injectGreekStyles } from '../../styles/greekTheme';
import { ThemeToggle } from '../ui/ThemeToggle';

export const Header: React.FC = () => {
  const { user, logout } = useAuth();
  useEffect(() => { injectGreekStyles(); }, []);

  if (!user) return null;

  return (
    <header className="greek-navbar">
      <div style={{ flex: 1 }} />

      <div className="flex items-center gap-4">
        <ThemeToggle />
        <div className="flex items-center gap-2">
          <span className="greek-text" style={{ fontSize: '13px', letterSpacing: '0.02em' }}>
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
