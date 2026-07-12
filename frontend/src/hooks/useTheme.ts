import { useState, useEffect } from 'react';
import { THEME_KEY } from '../styles/greekTheme';

export const useTheme = () => {
  const [isLight, setIsLight] = useState(
    () => localStorage.getItem(THEME_KEY) === 'light'
  );
  useEffect(() => {
    const obs = new MutationObserver(() => {
      setIsLight(document.documentElement.getAttribute('data-theme') === 'light');
    });
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    return () => obs.disconnect();
  }, []);
  return { isLight, isDark: !isLight };
};
