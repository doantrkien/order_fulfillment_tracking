import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { toast } from 'react-hot-toast';

// ── Theme CSS injection ────────────────────────────────────────────────────────
const injectStyles = () => {
  if (document.getElementById('greek-login-styles')) return;

  const link = document.createElement('link');
  link.id = 'greek-login-styles';
  link.rel = 'stylesheet';
  link.href = 'https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&family=Inter:wght@300;400;500&display=swap';
  document.head.appendChild(link);

  const style = document.createElement('style');
  style.id = 'greek-login-css';
  style.textContent = `

    @keyframes shimmer {
      0%   { background-position: -200% center; }
      100% { background-position:  200% center; }
    }
    @keyframes floatUp {
      0%, 100% { transform: translateY(0); }
      50%       { transform: translateY(-6px); }
    }
    @keyframes glowPulse {
      0%, 100% { box-shadow: 0 0 30px rgba(201,168,76,0.15), 0 0 60px rgba(201,168,76,0.05); }
      50%       { box-shadow: 0 0 40px rgba(201,168,76,0.30), 0 0 80px rgba(201,168,76,0.12); }
    }
    @keyframes glowPulseLight {
      0%, 100% { box-shadow: 0 0 30px rgba(160,120,20,0.12), 0 8px 40px rgba(0,0,0,0.10); }
      50%       { box-shadow: 0 0 50px rgba(160,120,20,0.22), 0 8px 48px rgba(0,0,0,0.14); }
    }
    @keyframes fadeSlideIn {
      from { opacity: 0; transform: translateY(24px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    @keyframes rotateSlow {
      from { transform: rotate(0deg); }
      to   { transform: rotate(360deg); }
    }
    @keyframes themeSwitch {
      from { opacity: 0.7; transform: scale(0.97); }
      to   { opacity: 1;   transform: scale(1); }
    }

    /* ── DARK: page bg ── */
    .greek-login-bg {
      min-height: 100vh;
      background: #0A0A0F;
      display: flex; align-items: center; justify-content: center;
      position: relative; overflow: hidden;
      font-family: 'Inter', sans-serif;
      transition: background 0.4s ease;
    }
    .greek-login-bg::before {
      content: ''; position: absolute; inset: 0; pointer-events: none;
      background:
        radial-gradient(ellipse 60% 40% at 50% 0%,   rgba(201,168,76,0.12) 0%, transparent 70%),
        radial-gradient(ellipse 40% 30% at 50% 100%, rgba(201,168,76,0.06) 0%, transparent 70%);
      transition: background 0.4s;
    }
    .greek-login-bg::after {
      content: ''; position: absolute; inset: 0; pointer-events: none;
      background-image:
        linear-gradient(rgba(201,168,76,0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(201,168,76,0.04) 1px, transparent 1px);
      background-size: 60px 60px;
      transition: background-image 0.4s;
    }

    /* ── DARK: card ── */
    .login-card {
      position: relative; z-index: 10;
      width: 100%; max-width: 420px; margin: 0 20px;
      background: linear-gradient(160deg, #13131C 0%, #0D0D15 60%, #111118 100%);
      border: 1px solid rgba(201,168,76,0.30);
      border-radius: 4px; padding: 48px 44px 44px;
      animation: fadeSlideIn 0.6s ease both, glowPulse 4s ease-in-out infinite;
      transition: background 0.4s, border-color 0.4s;
    }
    .login-card::before        { content:''; position:absolute; width:20px; height:20px; border-color:#C9A84C; border-style:solid; opacity:0.7; top:-1px;    left:-1px;  border-width:2px 0 0 2px; transition:border-color 0.4s; }
    .login-card::after         { content:''; position:absolute; width:20px; height:20px; border-color:#C9A84C; border-style:solid; opacity:0.7; bottom:-1px; right:-1px; border-width:0 2px 2px 0; transition:border-color 0.4s; }
    .login-card-corners::before{ content:''; position:absolute; width:20px; height:20px; border-color:#C9A84C; border-style:solid; opacity:0.7; bottom:-1px; left:-1px;  border-width:0 0 2px 2px; transition:border-color 0.4s; }
    .login-card-corners::after { content:''; position:absolute; width:20px; height:20px; border-color:#C9A84C; border-style:solid; opacity:0.7; top:-1px;    right:-1px; border-width:2px 2px 0 0; transition:border-color 0.4s; }

    .login-emblem { display:flex; flex-direction:column; align-items:center; margin-bottom:32px; animation:floatUp 5s ease-in-out infinite; }
    .login-emblem-title {
      font-family:'Cinzel',serif; font-size:22px; font-weight:700;
      letter-spacing:0.18em; text-transform:uppercase;
      background:linear-gradient(135deg,#C9A84C 0%,#F0C040 40%,#E8D5A3 60%,#C9A84C 100%);
      background-size:200% auto;
      -webkit-background-clip:text; -webkit-text-fill-color:transparent; background-clip:text;
      animation:shimmer 4s linear infinite; margin-top:14px;
    }
    .login-emblem-sub {
      font-family:'Inter',sans-serif; font-size:11px; font-weight:400;
      letter-spacing:0.30em; text-transform:uppercase;
      color:rgba(201,168,76,0.70); margin-top:5px; transition:color 0.4s;
    }

    .login-divider { display:flex; align-items:center; gap:10px; margin-bottom:28px; }
    .login-divider-line    { flex:1; height:1px; background:linear-gradient(90deg,transparent,rgba(201,168,76,0.35),transparent); transition:background 0.4s; }
    .login-divider-diamond { width:5px; height:5px; background:#C9A84C; transform:rotate(45deg); opacity:0.6; transition:background 0.4s; }

    .login-label {
      display:block; font-family:'Cinzel',serif; font-size:10px; font-weight:600;
      letter-spacing:0.22em; text-transform:uppercase;
      color:rgba(201,168,76,0.65); margin-bottom:8px; transition:color 0.4s;
    }
    .login-input {
      width:100%; background:rgba(201,168,76,0.04); border:1px solid rgba(201,168,76,0.18);
      border-radius:2px; padding:13px 16px;
      font-family:'Inter',sans-serif; font-size:14px; font-weight:400;
      color:#E8D5A3; outline:none;
      transition:border-color 0.25s, background 0.25s, box-shadow 0.25s, color 0.4s;
      box-sizing:border-box; letter-spacing:0.02em;
    }
    .login-input::placeholder { color:rgba(201,168,76,0.25); font-style:italic; }
    .login-input:focus {
      border-color:rgba(201,168,76,0.60); background:rgba(201,168,76,0.07);
      box-shadow:0 0 0 3px rgba(201,168,76,0.08), inset 0 1px 3px rgba(0,0,0,0.3);
    }
    .login-field { margin-bottom:20px; }

    .login-btn {
      width:100%; margin-top:8px; padding:15px;
      font-family:'Cinzel',serif; font-size:12px; font-weight:700;
      letter-spacing:0.30em; text-transform:uppercase; color:#0A0A0F;
      background:linear-gradient(135deg,#C9A84C 0%,#F0C040 45%,#C9A84C 100%);
      background-size:200% auto; border:none; border-radius:2px; cursor:pointer;
      transition:background-position 0.4s, transform 0.15s, box-shadow 0.2s;
      display:flex; align-items:center; justify-content:center; gap:8px;
    }
    .login-btn:hover:not(:disabled) {
      background-position:right center; transform:translateY(-1px);
      box-shadow:0 6px 24px rgba(201,168,76,0.35);
    }
    .login-btn:active:not(:disabled) { transform:translateY(0); }
    .login-btn:disabled { opacity:0.55; cursor:not-allowed; }

    .login-error {
      margin-bottom:20px; padding:12px 16px;
      background:rgba(180,50,50,0.12); border:1px solid rgba(180,50,50,0.30);
      border-radius:2px; font-size:13px; color:#E88080;
      font-family:'Inter',sans-serif; letter-spacing:0.01em;
      transition:background 0.4s, color 0.4s;
    }

    .login-hints { margin-top:28px; text-align:center; }
    .login-hints-title {
      font-family:'Cinzel',serif; font-size:9px; letter-spacing:0.30em;
      text-transform:uppercase; color:rgba(201,168,76,0.50); margin-bottom:10px; transition:color 0.4s;
    }
    .login-hints p { font-family:'Inter',sans-serif; font-size:12px; color:rgba(232,213,163,0.55); margin:5px 0; letter-spacing:0.02em; transition:color 0.4s; }
    .login-hints span { color:rgba(240,192,64,0.80); font-weight:500; }

    .login-spinner {
      display:inline-block; width:14px; height:14px;
      border:2px solid rgba(10,10,15,0.3); border-top-color:#0A0A0F;
      border-radius:50%; animation:rotateSlow 0.7s linear infinite; flex-shrink:0;
    }
    .login-loading {
      min-height:100vh; background:#0A0A0F;
      display:flex; align-items:center; justify-content:center; transition:background 0.4s;
    }
    .login-loading-text {
      font-family:'Cinzel',serif; font-size:11px; letter-spacing:0.35em;
      text-transform:uppercase; color:rgba(201,168,76,0.55);
      animation:glowPulse 2s ease-in-out infinite;
    }

    /* ── Theme toggle button ── */
    .login-theme-toggle {
      position:absolute; top:20px; right:20px; z-index:20;
      width:40px; height:40px; border-radius:50%;
      background:rgba(201,168,76,0.08); border:1px solid rgba(201,168,76,0.25);
      cursor:pointer; display:flex; align-items:center; justify-content:center;
      transition:background 0.3s, border-color 0.3s, transform 0.2s;
    }
    .login-theme-toggle:hover { background:rgba(201,168,76,0.16); border-color:rgba(201,168,76,0.50); transform:scale(1.08); }

    /* ============================================================
       LIGHT MODE overrides
       Palette: ivory (#F5F0E8) bg · warm white card · deep gold text
    ============================================================ */
    .greek-login-bg.light-mode { background:#F5F0E8; }
    .greek-login-bg.light-mode::before {
      background:
        radial-gradient(ellipse 60% 40% at 50% 0%,   rgba(160,120,20,0.10) 0%, transparent 70%),
        radial-gradient(ellipse 40% 30% at 50% 100%, rgba(160,120,20,0.05) 0%, transparent 70%);
    }
    .greek-login-bg.light-mode::after {
      background-image:
        linear-gradient(rgba(160,120,20,0.06) 1px, transparent 1px),
        linear-gradient(90deg, rgba(160,120,20,0.06) 1px, transparent 1px);
    }
    .greek-login-bg.light-mode .login-card {
      background:linear-gradient(160deg, #FFFDF7 0%, #FDF8EE 60%, #FAF4E6 100%);
      border-color:rgba(160,120,20,0.30);
      animation:fadeSlideIn 0.6s ease both, glowPulseLight 4s ease-in-out infinite;
    }
    .greek-login-bg.light-mode .login-card::before,
    .greek-login-bg.light-mode .login-card::after,
    .greek-login-bg.light-mode .login-card-corners::before,
    .greek-login-bg.light-mode .login-card-corners::after { border-color:#A07814; opacity:0.65; }

    .greek-login-bg.light-mode .login-emblem-sub        { color:rgba(110,75,8,0.75); }
    .greek-login-bg.light-mode .login-divider-line      { background:linear-gradient(90deg,transparent,rgba(160,120,20,0.38),transparent); }
    .greek-login-bg.light-mode .login-divider-diamond   { background:#A07814; opacity:0.65; }
    .greek-login-bg.light-mode .login-label             { color:rgba(90,60,5,0.80); }
    .greek-login-bg.light-mode .login-input             { background:rgba(160,120,20,0.05); border-color:rgba(160,120,20,0.22); color:#1E1300; }
    .greek-login-bg.light-mode .login-input::placeholder{ color:rgba(90,60,5,0.30); }
    .greek-login-bg.light-mode .login-input:focus       { border-color:rgba(160,120,20,0.60); background:rgba(160,120,20,0.08); box-shadow:0 0 0 3px rgba(160,120,20,0.09), inset 0 1px 3px rgba(0,0,0,0.04); }
    .greek-login-bg.light-mode .login-btn               { color:#1A0E00; }
    .greek-login-bg.light-mode .login-btn:hover:not(:disabled) { box-shadow:0 6px 24px rgba(160,120,20,0.28); }
    .greek-login-bg.light-mode .login-error             { background:rgba(180,50,50,0.07); border-color:rgba(160,40,40,0.28); color:#9B2020; }
    .greek-login-bg.light-mode .login-hints-title       { color:rgba(90,60,5,0.55); }
    .greek-login-bg.light-mode .login-hints p           { color:rgba(50,33,3,0.60); }
    .greek-login-bg.light-mode .login-hints span        { color:rgba(130,85,5,0.90); }
    .greek-login-bg.light-mode .login-theme-toggle      { background:rgba(160,120,20,0.10); border-color:rgba(160,120,20,0.30); }
    .greek-login-bg.light-mode .login-theme-toggle:hover{ background:rgba(160,120,20,0.18); border-color:rgba(160,120,20,0.55); }

    .theme-switching { animation:themeSwitch 0.3s ease both; }
  `;
  document.head.appendChild(style);
};

// ── SVG: Laurel wreath ────────────────────────────────────────────────────────
const LaurelEmblem: React.FC<{ isLight: boolean }> = ({ isLight }) => {
  const gold = isLight ? '#A07814' : '#C9A84C';
  const star = isLight ? '#8B6510' : '#F0C040';
  return (
    <svg width="72" height="56" viewBox="0 0 72 56" fill="none">
      <g opacity="0.85">
        <ellipse cx="12" cy="28" rx="5" ry="3" fill={gold} transform="rotate(-30 12 28)" />
        <ellipse cx="18" cy="20" rx="5" ry="3" fill={gold} transform="rotate(-50 18 20)" />
        <ellipse cx="10" cy="36" rx="4.5" ry="2.8" fill={gold} transform="rotate(-10 10 36)" />
        <ellipse cx="22" cy="13" rx="4.5" ry="2.8" fill={gold} transform="rotate(-65 22 13)" />
        <ellipse cx="8" cy="44" rx="4" ry="2.5" fill={gold} transform="rotate(10 8 44)" />
        <path d="M36 48 Q20 40 8 20" stroke={gold} strokeWidth="1.2" fill="none" opacity="0.5" />
      </g>
      <g opacity="0.85" transform="scale(-1,1) translate(-72,0)">
        <ellipse cx="12" cy="28" rx="5" ry="3" fill={gold} transform="rotate(-30 12 28)" />
        <ellipse cx="18" cy="20" rx="5" ry="3" fill={gold} transform="rotate(-50 18 20)" />
        <ellipse cx="10" cy="36" rx="4.5" ry="2.8" fill={gold} transform="rotate(-10 10 36)" />
        <ellipse cx="22" cy="13" rx="4.5" ry="2.8" fill={gold} transform="rotate(-65 22 13)" />
        <ellipse cx="8" cy="44" rx="4" ry="2.5" fill={gold} transform="rotate(10 8 44)" />
        <path d="M36 48 Q20 40 8 20" stroke={gold} strokeWidth="1.2" fill="none" opacity="0.5" />
      </g>
      <polygon points="36,18 37.5,23 42,23 38.5,26 40,31 36,28 32,31 33.5,26 30,23 34.5,23" fill={star} opacity="0.9" />
    </svg>
  );
};

// ── Toggle icons ──────────────────────────────────────────────────────────────
const IconMoon = () => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path d="M15 10.5A6 6 0 0 1 7.5 3a6 6 0 1 0 7.5 7.5z" fill="#C9A84C" opacity="0.85" />
  </svg>
);
const IconSun = () => (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <circle cx="9" cy="9" r="3.5" fill="#A07814" opacity="0.85" />
    <g stroke="#A07814" strokeWidth="1.3" strokeLinecap="round" opacity="0.7">
      <line x1="9" y1="1.5" x2="9" y2="3" />
      <line x1="9" y1="15" x2="9" y2="16.5" />
      <line x1="1.5" y1="9" x2="3" y2="9" />
      <line x1="15" y1="9" x2="16.5" y2="9" />
      <line x1="3.6" y1="3.6" x2="4.7" y2="4.7" />
      <line x1="13.3" y1="13.3" x2="14.4" y2="14.4" />
      <line x1="14.4" y1="3.6" x2="13.3" y2="4.7" />
      <line x1="4.7" y1="13.3" x2="3.6" y2="14.4" />
    </g>
  </svg>
);

// ── Component ─────────────────────────────────────────────────────────────────
const THEME_KEY = 'greek-theme-mode';

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLight, setIsLight] = useState(() => localStorage.getItem(THEME_KEY) === 'light');
  const [switching, setSwitching] = useState(false);

  const { login, isAuthenticated, user, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => { injectStyles(); }, []);

  useEffect(() => {
    if (!isLoading && isAuthenticated && user) {
      const from = (location.state as any)?.from?.pathname;
      navigate(
        from ?? (user.role === 'admin' ? '/admin/orders' : '/driver/orders'),
        { replace: true }
      );
    }
  }, [isAuthenticated, isLoading, user, navigate, location]);

  // Persist + broadcast to the rest of the app via data-theme attribute
  useEffect(() => {
    localStorage.setItem(THEME_KEY, isLight ? 'light' : 'dark');
    document.documentElement.setAttribute('data-theme', isLight ? 'light' : 'dark');
  }, [isLight]);

  const toggleTheme = () => {
    setSwitching(true);
    setTimeout(() => { setIsLight(v => !v); setSwitching(false); }, 150);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    try {
      await login({ email, password });
      toast.success('Access granted.');
    } catch (err: any) {
      const msg = err.message || 'Authentication failed. Verify your credentials.';
      setError(msg);
      toast.error(msg);
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className={`login-loading${isLight ? ' light-mode' : ''}`}>
        <span className="login-loading-text">Authenticating…</span>
      </div>
    );
  }

  return (
    <div className={`greek-login-bg${isLight ? ' light-mode' : ''}`}>

      {/* Theme toggle */}
      <button className="login-theme-toggle" onClick={toggleTheme}
        title={isLight ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
        aria-label="Toggle theme">
        {isLight ? <IconSun /> : <IconMoon />}
      </button>

      {/* Card */}
      <div className={`login-card login-card-corners${switching ? ' theme-switching' : ''}`}>

        <div className="login-emblem">
          <LaurelEmblem isLight={isLight} />
          <div className="login-emblem-title">Tracking Portal</div>
          <div className="login-emblem-sub">Olympian Logistics Command</div>
        </div>

        <div className="login-divider">
          <div className="login-divider-line" />
          <div className="login-divider-diamond" />
          <div className="login-divider-line" />
        </div>

        <form onSubmit={handleSubmit} autoComplete="off">
          {error && <div className="login-error">{error}</div>}

          <div className="login-field">
            <label className="login-label" htmlFor="email">Email Address</label>
            <input id="email" type="email" className="login-input"
              placeholder="herald@olympus.com"
              value={email} onChange={e => setEmail(e.target.value)} required />
          </div>

          <div className="login-field">
            <label className="login-label" htmlFor="password">Password</label>
            <input id="password" type="password" className="login-input"
              placeholder="••••••••"
              value={password} onChange={e => setPassword(e.target.value)} required />
          </div>

          <button type="submit" className="login-btn" disabled={isSubmitting}>
            {isSubmitting
              ? <><span className="login-spinner" />Entering…</>
              : 'Enter the Pantheon'}
          </button>
        </form>

        <div className="login-hints">
          <div className="login-hints-title">⸻ Demo Credentials ⸻</div>
          <p>Admin — <span>admin@demo.com</span> / password</p>
          <p>Driver — <span>driver@demo.com</span> / password</p>
        </div>
      </div>
    </div>
  );
};