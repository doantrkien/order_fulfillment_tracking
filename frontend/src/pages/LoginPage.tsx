import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { toast } from 'react-hot-toast';

// Inject Google Fonts + styles once
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
      0%, 100% { transform: translateY(0px); }
      50%       { transform: translateY(-6px); }
    }
    @keyframes glowPulse {
      0%, 100% { box-shadow: 0 0 30px rgba(201,168,76,0.15), 0 0 60px rgba(201,168,76,0.05); }
      50%       { box-shadow: 0 0 40px rgba(201,168,76,0.30), 0 0 80px rgba(201,168,76,0.12); }
    }
    @keyframes fadeSlideIn {
      from { opacity: 0; transform: translateY(24px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    @keyframes rotateSlow {
      from { transform: rotate(0deg); }
      to   { transform: rotate(360deg); }
    }

    .greek-bg {
      min-height: 100vh;
      background: #0A0A0F;
      display: flex;
      align-items: center;
      justify-content: center;
      position: relative;
      overflow: hidden;
      font-family: 'Inter', sans-serif;
    }

    /* Radial ambient glow */
    .greek-bg::before {
      content: '';
      position: absolute;
      inset: 0;
      background:
        radial-gradient(ellipse 60% 40% at 50% 0%,   rgba(201,168,76,0.12) 0%, transparent 70%),
        radial-gradient(ellipse 40% 30% at 50% 100%, rgba(201,168,76,0.06) 0%, transparent 70%);
      pointer-events: none;
    }

    /* Subtle grid texture */
    .greek-bg::after {
      content: '';
      position: absolute;
      inset: 0;
      background-image:
        linear-gradient(rgba(201,168,76,0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(201,168,76,0.04) 1px, transparent 1px);
      background-size: 60px 60px;
      pointer-events: none;
    }

    .greek-card {
      position: relative;
      z-index: 10;
      width: 100%;
      max-width: 420px;
      margin: 0 20px;
      background: linear-gradient(160deg, #13131C 0%, #0D0D15 60%, #111118 100%);
      border: 1px solid rgba(201,168,76,0.30);
      border-radius: 4px;
      padding: 48px 44px 44px;
      animation: fadeSlideIn 0.6s ease both, glowPulse 4s ease-in-out infinite;
    }

    /* Corner ornaments */
    .greek-card::before,
    .greek-card::after {
      content: '';
      position: absolute;
      width: 20px;
      height: 20px;
      border-color: #C9A84C;
      border-style: solid;
      opacity: 0.7;
    }
    .greek-card::before {
      top: -1px; left: -1px;
      border-width: 2px 0 0 2px;
    }
    .greek-card::after {
      bottom: -1px; right: -1px;
      border-width: 0 2px 2px 0;
    }

    /* Inner corner bottom-left & top-right via pseudo on a wrapper */
    .greek-card-inner::before,
    .greek-card-inner::after {
      content: '';
      position: absolute;
      width: 20px;
      height: 20px;
      border-color: #C9A84C;
      border-style: solid;
      opacity: 0.7;
    }
    .greek-card-inner::before {
      bottom: -1px; left: -1px;
      border-width: 0 0 2px 2px;
    }
    .greek-card-inner::after {
      top: -1px; right: -1px;
      border-width: 2px 2px 0 0;
    }

    .emblem {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin-bottom: 32px;
      animation: floatUp 5s ease-in-out infinite;
    }

    .emblem-title {
      font-family: 'Cinzel', serif;
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 0.18em;
      text-transform: uppercase;
      background: linear-gradient(135deg, #C9A84C 0%, #F0C040 40%, #E8D5A3 60%, #C9A84C 100%);
      background-size: 200% auto;
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      animation: shimmer 4s linear infinite;
      margin-top: 14px;
    }

    .emblem-sub {
      font-family: 'Inter', sans-serif;
      font-size: 11px;
      font-weight: 300;
      letter-spacing: 0.35em;
      text-transform: uppercase;
      color: rgba(201,168,76,0.45);
      margin-top: 5px;
    }

    .divider {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 28px;
    }
    .divider-line {
      flex: 1;
      height: 1px;
      background: linear-gradient(90deg, transparent, rgba(201,168,76,0.35), transparent);
    }
    .divider-diamond {
      width: 5px;
      height: 5px;
      background: #C9A84C;
      transform: rotate(45deg);
      opacity: 0.6;
    }

    .greek-label {
      display: block;
      font-family: 'Cinzel', serif;
      font-size: 10px;
      font-weight: 600;
      letter-spacing: 0.22em;
      text-transform: uppercase;
      color: rgba(201,168,76,0.65);
      margin-bottom: 8px;
    }

    .greek-input {
      width: 100%;
      background: rgba(201,168,76,0.04);
      border: 1px solid rgba(201,168,76,0.18);
      border-radius: 2px;
      padding: 13px 16px;
      font-family: 'Inter', sans-serif;
      font-size: 14px;
      font-weight: 400;
      color: #E8D5A3;
      outline: none;
      transition: border-color 0.25s, background 0.25s, box-shadow 0.25s;
      box-sizing: border-box;
      letter-spacing: 0.02em;
    }
    .greek-input::placeholder {
      color: rgba(201,168,76,0.20);
      font-style: italic;
    }
    .greek-input:focus {
      border-color: rgba(201,168,76,0.60);
      background: rgba(201,168,76,0.07);
      box-shadow: 0 0 0 3px rgba(201,168,76,0.08), inset 0 1px 3px rgba(0,0,0,0.3);
    }

    .form-field {
      margin-bottom: 20px;
    }

    .greek-btn {
      width: 100%;
      margin-top: 8px;
      padding: 15px;
      font-family: 'Cinzel', serif;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.30em;
      text-transform: uppercase;
      color: #0A0A0F;
      background: linear-gradient(135deg, #C9A84C 0%, #F0C040 45%, #C9A84C 100%);
      background-size: 200% auto;
      border: none;
      border-radius: 2px;
      cursor: pointer;
      transition: background-position 0.4s, opacity 0.2s, transform 0.15s;
      position: relative;
      overflow: hidden;
    }
    .greek-btn:hover:not(:disabled) {
      background-position: right center;
      transform: translateY(-1px);
      box-shadow: 0 6px 24px rgba(201,168,76,0.35);
    }
    .greek-btn:active:not(:disabled) {
      transform: translateY(0px);
    }
    .greek-btn:disabled {
      opacity: 0.55;
      cursor: not-allowed;
    }

    .greek-error {
      margin-bottom: 20px;
      padding: 12px 16px;
      background: rgba(180,50,50,0.12);
      border: 1px solid rgba(180,50,50,0.30);
      border-radius: 2px;
      font-size: 13px;
      color: #E88080;
      font-family: 'Inter', sans-serif;
      letter-spacing: 0.01em;
    }

    .demo-hints {
      margin-top: 28px;
      text-align: center;
    }
    .demo-hints-title {
      font-family: 'Cinzel', serif;
      font-size: 9px;
      letter-spacing: 0.30em;
      text-transform: uppercase;
      color: rgba(201,168,76,0.28);
      margin-bottom: 8px;
    }
    .demo-hints p {
      font-family: 'Inter', sans-serif;
      font-size: 11px;
      color: rgba(201,168,76,0.25);
      margin: 3px 0;
      letter-spacing: 0.02em;
    }
    .demo-hints span {
      color: rgba(201,168,76,0.45);
    }

    .loading-screen {
      min-height: 100vh;
      background: #0A0A0F;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .loading-text {
      font-family: 'Cinzel', serif;
      font-size: 11px;
      letter-spacing: 0.35em;
      text-transform: uppercase;
      color: rgba(201,168,76,0.45);
      animation: glowPulse 2s ease-in-out infinite;
    }

    /* Spinning ring loader for submit */
    .spin-ring {
      display: inline-block;
      width: 14px;
      height: 14px;
      border: 2px solid rgba(10,10,15,0.3);
      border-top-color: #0A0A0F;
      border-radius: 50%;
      animation: rotateSlow 0.7s linear infinite;
      vertical-align: middle;
      margin-right: 8px;
    }
  `;
  document.head.appendChild(style);
};

// Laurel wreath SVG emblem
const LaurelEmblem: React.FC = () => (
  <svg width="72" height="56" viewBox="0 0 72 56" fill="none" xmlns="http://www.w3.org/2000/svg">
    {/* Left branch */}
    <g opacity="0.85">
      <ellipse cx="12" cy="28" rx="5" ry="3" fill="#C9A84C" transform="rotate(-30 12 28)" />
      <ellipse cx="18" cy="20" rx="5" ry="3" fill="#C9A84C" transform="rotate(-50 18 20)" />
      <ellipse cx="10" cy="36" rx="4.5" ry="2.8" fill="#C9A84C" transform="rotate(-10 10 36)" />
      <ellipse cx="22" cy="13" rx="4.5" ry="2.8" fill="#C9A84C" transform="rotate(-65 22 13)" />
      <ellipse cx="8" cy="44" rx="4" ry="2.5" fill="#C9A84C" transform="rotate( 10 8 44)" />
      <path d="M36 48 Q20 40 8 20" stroke="#C9A84C" strokeWidth="1.2" fill="none" opacity="0.5" />
    </g>
    {/* Right branch (mirrored) */}
    <g opacity="0.85" transform="scale(-1,1) translate(-72,0)">
      <ellipse cx="12" cy="28" rx="5" ry="3" fill="#C9A84C" transform="rotate(-30 12 28)" />
      <ellipse cx="18" cy="20" rx="5" ry="3" fill="#C9A84C" transform="rotate(-50 18 20)" />
      <ellipse cx="10" cy="36" rx="4.5" ry="2.8" fill="#C9A84C" transform="rotate(-10 10 36)" />
      <ellipse cx="22" cy="13" rx="4.5" ry="2.8" fill="#C9A84C" transform="rotate(-65 22 13)" />
      <ellipse cx="8" cy="44" rx="4" ry="2.5" fill="#C9A84C" transform="rotate( 10 8 44)" />
      <path d="M36 48 Q20 40 8 20" stroke="#C9A84C" strokeWidth="1.2" fill="none" opacity="0.5" />
    </g>
    {/* Center star */}
    <polygon points="36,18 37.5,23 42,23 38.5,26 40,31 36,28 32,31 33.5,26 30,23 34.5,23" fill="#F0C040" opacity="0.9" />
  </svg>
);

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

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
      <div className="loading-screen">
        <span className="loading-text">Authenticating…</span>
      </div>
    );
  }

  return (
    <div className="greek-bg">
      <div className="greek-card greek-card-inner">

        {/* Emblem */}
        <div className="emblem">
          <LaurelEmblem />
          <div className="emblem-title">Tracking Portal</div>
          <div className="emblem-sub">Olympian Logistics Command</div>
        </div>

        {/* Divider */}
        <div className="divider">
          <div className="divider-line" />
          <div className="divider-diamond" />
          <div className="divider-line" />
        </div>

        <form onSubmit={handleSubmit} autoComplete="off">
          {error && <div className="greek-error">{error}</div>}

          <div className="form-field">
            <label className="greek-label" htmlFor="email">Email Address</label>
            <input
              id="email"
              type="email"
              className="greek-input"
              placeholder="herald@olympus.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </div>

          <div className="form-field">
            <label className="greek-label" htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              className="greek-input"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
            />
          </div>

          <button type="submit" className="greek-btn" disabled={isSubmitting}>
            {isSubmitting
              ? <><span className="spin-ring" />Entering…</>
              : 'Enter the Pantheon'
            }
          </button>
        </form>

        {/* Demo hints */}
        <div className="demo-hints">
          <div className="demo-hints-title">⸻ Demo Credentials ⸻</div>
          <p>Admin — <span>admin@demo.com</span> / password</p>
          <p>Driver — <span>driver@demo.com</span> / password</p>
        </div>
      </div>
    </div>
  );
};