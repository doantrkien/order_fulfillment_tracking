import React, { useState, useEffect } from 'react';
import { THEME_KEY, applyTheme } from '../../styles/greekTheme';

const IconMoon = () => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path d="M15 10.5A6 6 0 0 1 7.5 3a6 6 0 1 0 7.5 7.5z" fill="#C9A84C" opacity="0.85"/>
  </svg>
);
const IconSun = () => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <circle cx="9" cy="9" r="3.5" fill="#A07814" opacity="0.85"/>
    <g stroke="#A07814" strokeWidth="1.3" strokeLinecap="round" opacity="0.7">
      <line x1="9"    y1="1.5"  x2="9"    y2="3"   />
      <line x1="9"    y1="15"   x2="9"    y2="16.5" />
      <line x1="1.5"  y1="9"    x2="3"    y2="9"    />
      <line x1="15"   y1="9"    x2="16.5" y2="9"    />
      <line x1="3.6"  y1="3.6"  x2="4.7"  y2="4.7"  />
      <line x1="13.3" y1="13.3" x2="14.4" y2="14.4" />
      <line x1="14.4" y1="3.6"  x2="13.3" y2="4.7"  />
      <line x1="4.7"  y1="13.3" x2="3.6"  y2="14.4" />
    </g>
  </svg>
);

export const ThemeToggle: React.FC<{ style?: React.CSSProperties }> = ({ style }) => {
  const [isLight, setIsLight] = useState(
    () => localStorage.getItem(THEME_KEY) === 'light'
  );

  const toggle = () => {
    const next = !isLight;
    applyTheme(next ? 'light' : 'dark');
    setIsLight(next);
  };

  // Sync if another tab changes the theme
  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key === THEME_KEY) setIsLight(e.newValue === 'light');
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, []);

  return (
    <button
      onClick={toggle}
      title={isLight ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
      aria-label="Toggle theme"
      style={{
        width: 38, height: 38, borderRadius: '50%',
        background: 'rgba(201,168,76,0.08)',
        border: '1px solid rgba(201,168,76,0.25)',
        cursor: 'pointer',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        transition: 'background 0.3s, border-color 0.3s, transform 0.2s',
        flexShrink: 0,
        ...style,
      }}
      onMouseEnter={e => { (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1.08)'; }}
      onMouseLeave={e => { (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)'; }}
    >
      {isLight ? <IconSun /> : <IconMoon />}
    </button>
  );
};
