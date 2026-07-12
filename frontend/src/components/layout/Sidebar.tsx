import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useEffect } from 'react';
import { injectGreekStyles } from '../../styles/greekTheme';

const SIDEBAR_BRAND_STYLE: React.CSSProperties = {
  padding: '28px 20px 20px',
  borderBottom: '1px solid rgba(201,168,76,0.15)',
  background: 'linear-gradient(180deg, rgba(201,168,76,0.06) 0%, transparent 100%)',
};

const BRAND_TEXT_STYLE: React.CSSProperties = {
  fontFamily: "'Cinzel', serif",
  fontSize: '13px',
  fontWeight: 700,
  letterSpacing: '0.22em',
  textTransform: 'uppercase',
  background: 'linear-gradient(135deg, #C9A84C 0%, #F0C040 50%, #C9A84C 100%)',
  backgroundSize: '200% auto',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  backgroundClip: 'text',
  animation: 'greek-shimmer 5s linear infinite',
};

export const Sidebar: React.FC = () => {
  const { user } = useAuth();
  useEffect(() => { injectGreekStyles(); }, []);

  if (!user) return null;

  return (
    <aside className="greek-sidebar" style={{ width: '240px', flexShrink: 0 }}>
      <div style={SIDEBAR_BRAND_STYLE}>
        <span style={BRAND_TEXT_STYLE}>TrackPro</span>
      </div>

      <nav style={{ padding: '12px 0', display: 'flex', flexDirection: 'column', gap: '2px', flex: 1, overflowY: 'auto' }}>
        {user.role === 'admin' ? (
          <>
            <NavLink to="/admin/dashboard" className={({ isActive }) => `greek-sidebar-item${isActive ? ' active' : ''}`}>
              Dashboard
            </NavLink>
            <NavLink to="/admin/orders" className={({ isActive }) => `greek-sidebar-item${isActive ? ' active' : ''}`}>
              Orders
            </NavLink>
            <NavLink to="/admin/reports" className={({ isActive }) => `greek-sidebar-item${isActive ? ' active' : ''}`}>
              Reports
            </NavLink>
          </>
        ) : (
          <NavLink to="/driver/orders" className={({ isActive }) => `greek-sidebar-item${isActive ? ' active' : ''}`}>
            My Deliveries
          </NavLink>
        )}
      </nav>
    </aside>
  );
};
