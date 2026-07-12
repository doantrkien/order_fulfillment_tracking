/**
 * Greek God Theme — shared inject helper
 * Gọi injectGreekStyles() trong useEffect của bất kỳ page nào cần theme.
 * Idempotent: kiểm tra id trước khi tạo — an toàn khi gọi nhiều lần.
 */

export const GREEK_CSS = `
/* =============================================
   GREEK GOD THEME — GLOBAL CSS
   Màu vàng thần thoại trên nền đen obsidian
   ============================================= */

/* ---------- Keyframe animations ---------- */
@keyframes greek-shimmer {
  0%   { background-position: -200% center; }
  100% { background-position:  200% center; }
}
@keyframes greek-float {
  0%, 100% { transform: translateY(0px); }
  50%       { transform: translateY(-6px); }
}
@keyframes greek-glow-pulse {
  0%, 100% { box-shadow: 0 0 30px rgba(201,168,76,0.15), 0 0 60px rgba(201,168,76,0.05); }
  50%       { box-shadow: 0 0 40px rgba(201,168,76,0.30), 0 0 80px rgba(201,168,76,0.12); }
}
@keyframes greek-fade-in {
  from { opacity: 0; transform: translateY(24px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes greek-spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}
@keyframes greek-row-in {
  from { opacity: 0; transform: translateX(-12px); }
  to   { opacity: 1; transform: translateX(0); }
}

/* ---------- Layout: full-page background ---------- */
.greek-bg {
  min-height: 100vh;
  background: #0A0A0F;
  position: relative;
  overflow: hidden;
  font-family: 'Inter', sans-serif;
  color: #E8D5A3;
}
.greek-bg::before {
  content: '';
  position: fixed;
  inset: 0;
  background:
    radial-gradient(ellipse 60% 40% at 50% 0%,   rgba(201,168,76,0.10) 0%, transparent 70%),
    radial-gradient(ellipse 40% 30% at 50% 100%, rgba(201,168,76,0.05) 0%, transparent 70%);
  pointer-events: none;
  z-index: 0;
}
.greek-bg::after {
  content: '';
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(201,168,76,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(201,168,76,0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  pointer-events: none;
  z-index: 0;
}

/* ---------- Page content wrapper ---------- */
.greek-page {
  position: relative;
  z-index: 1;
  padding: 32px 36px;
  min-height: 100vh;
}

/* ---------- Card ---------- */
.greek-card {
  position: relative;
  background: linear-gradient(160deg, #13131C 0%, #0D0D15 60%, #111118 100%);
  border: 1px solid rgba(201,168,76,0.28);
  border-radius: 4px;
  padding: 32px 36px;
  animation: greek-fade-in 0.5s ease both, greek-glow-pulse 4s ease-in-out infinite;
}
.greek-card-corners::before,
.greek-card-corners::after,
.greek-card::before,
.greek-card::after {
  content: '';
  position: absolute;
  width: 18px;
  height: 18px;
  border-color: #C9A84C;
  border-style: solid;
  opacity: 0.65;
}
.greek-card::before { top: -1px;    left: -1px;  border-width: 2px 0 0 2px; }
.greek-card::after  { bottom: -1px; right: -1px; border-width: 0 2px 2px 0; }
.greek-card-corners::before { bottom: -1px; left: -1px;  border-width: 0 0 2px 2px; }
.greek-card-corners::after  { top: -1px;    right: -1px; border-width: 2px 2px 0 0; }

.greek-card-static {
  position: relative;
  background: linear-gradient(160deg, #13131C 0%, #0D0D15 60%, #111118 100%);
  border: 1px solid rgba(201,168,76,0.20);
  border-radius: 4px;
  padding: 20px 24px;
  transition: border-color 0.25s, box-shadow 0.25s;
}
.greek-card-static:hover {
  border-color: rgba(201,168,76,0.45);
  box-shadow: 0 4px 20px rgba(201,168,76,0.10);
}

/* ---------- Typography ---------- */
.greek-heading {
  font-family: 'Cinzel', serif;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  background: linear-gradient(135deg, #C9A84C 0%, #F0C040 40%, #E8D5A3 60%, #C9A84C 100%);
  background-size: 200% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: greek-shimmer 5s linear infinite;
}
.greek-subheading {
  font-family: 'Cinzel', serif;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #C9A84C;
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
.greek-text {
  font-family: 'Inter', sans-serif;
  color: rgba(232,213,163,0.75);
  font-size: 14px;
  line-height: 1.6;
}
.greek-muted {
  font-family: 'Inter', sans-serif;
  color: rgba(232,213,163,0.55);
  font-size: 12px;
  line-height: 1.5;
}

/* ---------- Divider ---------- */
.greek-divider {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 20px 0;
}
.greek-divider-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(201,168,76,0.30), transparent);
}
.greek-divider-diamond {
  width: 5px;
  height: 5px;
  background: #C9A84C;
  transform: rotate(45deg);
  opacity: 0.55;
  flex-shrink: 0;
}

/* ---------- Form elements ---------- */
.greek-input {
  width: 100%;
  background: rgba(201,168,76,0.04);
  border: 1px solid rgba(201,168,76,0.18);
  border-radius: 2px;
  padding: 12px 16px;
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  color: #E8D5A3;
  outline: none;
  transition: border-color 0.25s, background 0.25s, box-shadow 0.25s;
  box-sizing: border-box;
}
.greek-input::placeholder { color: rgba(201,168,76,0.20); font-style: italic; }
.greek-input:focus {
  border-color: rgba(201,168,76,0.55);
  background: rgba(201,168,76,0.07);
  box-shadow: 0 0 0 3px rgba(201,168,76,0.08);
}
.greek-input:disabled { opacity: 0.45; cursor: not-allowed; }

.greek-select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6'%3E%3Cpath d='M0 0l5 6 5-6z' fill='%23C9A84C' opacity='.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 14px center;
  padding-right: 36px;
  cursor: pointer;
}
.greek-select option { background: #13131C; color: #E8D5A3; }

/* ---------- Buttons ---------- */
.greek-btn-primary {
  font-family: 'Cinzel', serif;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: #0A0A0F;
  background: linear-gradient(135deg, #C9A84C 0%, #F0C040 50%, #C9A84C 100%);
  background-size: 200% auto;
  border: none;
  border-radius: 2px;
  padding: 13px 28px;
  cursor: pointer;
  transition: background-position 0.4s, transform 0.15s, box-shadow 0.2s;
}
.greek-btn-primary:hover:not(:disabled) {
  background-position: right center;
  transform: translateY(-1px);
  box-shadow: 0 6px 24px rgba(201,168,76,0.35);
}
.greek-btn-primary:active:not(:disabled) { transform: translateY(0); }
.greek-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.greek-btn-secondary {
  font-family: 'Cinzel', serif;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: rgba(201,168,76,0.75);
  background: transparent;
  border: 1px solid rgba(201,168,76,0.30);
  border-radius: 2px;
  padding: 10px 20px;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s, background 0.2s;
}
.greek-btn-secondary:hover:not(:disabled) {
  border-color: rgba(201,168,76,0.65);
  color: #F0C040;
  background: rgba(201,168,76,0.06);
}
.greek-btn-secondary:disabled { opacity: 0.4; cursor: not-allowed; }

.greek-btn-danger {
  font-family: 'Cinzel', serif;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.20em;
  text-transform: uppercase;
  color: rgba(232,128,128,0.75);
  background: transparent;
  border: 1px solid rgba(180,50,50,0.35);
  border-radius: 2px;
  padding: 10px 20px;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s, background 0.2s;
}
.greek-btn-danger:hover:not(:disabled) {
  border-color: rgba(220,80,80,0.65);
  color: #E88080;
  background: rgba(180,50,50,0.08);
}

/* ---------- Badge / Status pill ---------- */
.greek-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-family: 'Cinzel', serif;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  padding: 4px 10px;
  border-radius: 2px;
  border: 1px solid;
}
.greek-badge-gold  { color: #F0C040; border-color: rgba(240,192,64,0.35); background: rgba(240,192,64,0.08); }
.greek-badge-green { color: #7EC88A; border-color: rgba(126,200,138,0.35); background: rgba(126,200,138,0.08); }
.greek-badge-red   { color: #E88080; border-color: rgba(232,128,128,0.35); background: rgba(232,128,128,0.08); }
.greek-badge-muted { color: rgba(201,168,76,0.40); border-color: rgba(201,168,76,0.15); background: transparent; }

/* ---------- Table ---------- */
.greek-table { width: 100%; border-collapse: collapse; }
.greek-table thead tr { border-bottom: 1px solid rgba(201,168,76,0.20); }
.greek-table th {
  font-family: 'Cinzel', serif;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.25em;
  text-transform: uppercase;
  color: rgba(201,168,76,0.60);
  padding: 10px 16px;
  text-align: left;
}
.greek-table td {
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  color: rgba(232,213,163,0.80);
  padding: 14px 16px;
  border-bottom: 1px solid rgba(201,168,76,0.07);
}
.greek-table tbody tr {
  transition: background 0.15s;
  animation: greek-row-in 0.3s ease both;
}
.greek-table tbody tr:hover { background: rgba(201,168,76,0.04); }

/* ---------- Sidebar ---------- */
.greek-sidebar {
  background: linear-gradient(180deg, #0D0D15 0%, #0A0A0F 100%);
  border-right: 1px solid rgba(201,168,76,0.15);
  height: 100vh;
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  z-index: 100;
}
.greek-sidebar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  font-family: 'Cinzel', serif;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: rgba(201,168,76,0.55);
  cursor: pointer;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
  border-left: 2px solid transparent;
  text-decoration: none;
  white-space: nowrap;
}
.greek-sidebar-item:hover {
  color: rgba(201,168,76,0.80);
  background: rgba(201,168,76,0.05);
  border-left-color: rgba(201,168,76,0.30);
}
.greek-sidebar-item.active {
  color: #F0C040;
  background: rgba(201,168,76,0.08);
  border-left-color: #C9A84C;
}

/* ---------- Top navbar ---------- */
.greek-navbar {
  background: rgba(10,10,15,0.92);
  border-bottom: 1px solid rgba(201,168,76,0.15);
  backdrop-filter: blur(12px);
  padding: 0 32px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 50;
}

/* ---------- Alert / notification ---------- */
.greek-alert { padding: 12px 16px; border-radius: 2px; font-family: 'Inter', sans-serif; font-size: 13px; margin-bottom: 16px; }
.greek-alert-error   { background: rgba(180,50,50,0.12);  border: 1px solid rgba(180,50,50,0.30);  color: #E88080; }
.greek-alert-success { background: rgba(50,150,80,0.10);  border: 1px solid rgba(50,150,80,0.30);  color: #7EC88A; }
.greek-alert-info    { background: rgba(201,168,76,0.08); border: 1px solid rgba(201,168,76,0.25); color: rgba(201,168,76,0.80); }

/* ---------- Loading states ---------- */
.greek-loading-screen {
  min-height: 100vh;
  background: #0A0A0F;
  display: flex;
  align-items: center;
  justify-content: center;
}
.greek-loading-text {
  font-family: 'Cinzel', serif;
  font-size: 11px;
  letter-spacing: 0.35em;
  text-transform: uppercase;
  color: rgba(201,168,76,0.55);
  animation: greek-glow-pulse 2s ease-in-out infinite;
}
.greek-spinner {
  display: inline-block;
  width: 16px; height: 16px;
  border: 2px solid rgba(201,168,76,0.20);
  border-top-color: #C9A84C;
  border-radius: 50%;
  animation: greek-spin 0.8s linear infinite;
  vertical-align: middle;
  margin-right: 8px;
}

/* ---------- Emblem ---------- */
.greek-emblem {
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: greek-float 5s ease-in-out infinite;
}

/* ---------- Modal overlay ---------- */
.greek-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.75);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  animation: greek-fade-in 0.2s ease;
}
.greek-modal {
  position: relative;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  margin: 0 20px;
  background: linear-gradient(160deg, #13131C 0%, #0D0D15 100%);
  border: 1px solid rgba(201,168,76,0.30);
  border-radius: 4px;
  animation: greek-fade-in 0.3s cubic-bezier(0.16,1,0.3,1);
}
.greek-modal-header {
  padding: 20px 28px;
  border-bottom: 1px solid rgba(201,168,76,0.15);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.greek-modal-body { padding: 24px 28px; }
.greek-modal-footer {
  padding: 16px 28px;
  border-top: 1px solid rgba(201,168,76,0.12);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* ---------- Utility ---------- */
.greek-close-btn {
  background: none;
  border: none;
  color: rgba(201,168,76,0.40);
  font-size: 22px;
  cursor: pointer;
  line-height: 1;
  transition: color 0.2s;
}
.greek-close-btn:hover { color: #C9A84C; }

.greek-stat-value {
  font-family: 'Cinzel', serif;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.05em;
  background: linear-gradient(135deg, #C9A84C 0%, #F0C040 50%, #C9A84C 100%);
  background-size: 200% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: greek-shimmer 5s linear infinite;
}
.greek-stat-value-success {
  font-family: 'Cinzel', serif;
  font-size: 28px;
  font-weight: 700;
  color: #7EC88A;
}
.greek-stat-value-danger {
  font-family: 'Cinzel', serif;
  font-size: 28px;
  font-weight: 700;
  color: #E88080;
}
.greek-stat-value-warning {
  font-family: 'Cinzel', serif;
  font-size: 28px;
  font-weight: 700;
  color: #fbbf24;
}
.greek-stat-value-neutral {
  font-family: 'Cinzel', serif;
  font-size: 28px;
  font-weight: 700;
  color: rgba(232,213,163,0.70);
}

/* ============================================================
   LIGHT MODE OVERRIDES
   Applied when <html data-theme="light">
   All selectors use [data-theme="light"] prefix so they win
   over the dark defaults with zero specificity tricks.
============================================================ */

/* Page background */
[data-theme="light"] .greek-bg {
  background: #F5F0E8;
}
[data-theme="light"] .greek-bg::before {
  background:
    radial-gradient(ellipse 60% 40% at 50% 0%,   rgba(160,120,20,0.10) 0%, transparent 70%),
    radial-gradient(ellipse 40% 30% at 50% 100%, rgba(160,120,20,0.05) 0%, transparent 70%);
}
[data-theme="light"] .greek-bg::after {
  background-image:
    linear-gradient(rgba(160,120,20,0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(160,120,20,0.06) 1px, transparent 1px);
}

/* Main card */
[data-theme="light"] .greek-card,
[data-theme="light"] .greek-card-static {
  background: linear-gradient(160deg, #FFFDF7 0%, #FDF8EE 60%, #FAF4E6 100%);
  border-color: rgba(160,120,20,0.28);
}
[data-theme="light"] .greek-card-static:hover {
  border-color: rgba(160,120,20,0.50);
  box-shadow: 0 4px 20px rgba(160,120,20,0.10);
}

/* Corner ornaments */
[data-theme="light"] .greek-card::before,
[data-theme="light"] .greek-card::after,
[data-theme="light"] .greek-card-corners::before,
[data-theme="light"] .greek-card-corners::after {
  border-color: #A07814;
  opacity: 0.60;
}

/* Body text */
[data-theme="light"] .greek-bg,
[data-theme="light"] .greek-page {
  color: #2C1F05;
}

/* Headings — shimmer stays, but gradient anchors darken */
[data-theme="light"] .greek-heading {
  background: linear-gradient(135deg, #8B6510 0%, #C9920A 40%, #6B4D08 60%, #8B6510 100%);
  background-size: 200% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
[data-theme="light"] .greek-subheading { color: #7A5C0A; }
[data-theme="light"] .greek-label      { color: rgba(80,55,5,0.80); }
[data-theme="light"] .greek-text       { color: rgba(44,31,5,0.85); }
[data-theme="light"] .greek-muted      { color: rgba(44,31,5,0.58); }
[data-theme="light"] .greek-caption    { color: rgba(90,60,5,0.52); }

/* Divider */
[data-theme="light"] .greek-divider-line {
  background: linear-gradient(90deg, transparent, rgba(160,120,20,0.35), transparent);
}
[data-theme="light"] .greek-divider-diamond { background: #A07814; opacity: 0.60; }

/* Inputs */
[data-theme="light"] .greek-input,
[data-theme="light"] .greek-select {
  background: rgba(160,120,20,0.05);
  border-color: rgba(160,120,20,0.22);
  color: #1E1300;
}
[data-theme="light"] .greek-input::placeholder,
[data-theme="light"] .greek-select::placeholder {
  color: rgba(90,60,5,0.30);
}
[data-theme="light"] .greek-input:focus,
[data-theme="light"] .greek-select:focus {
  border-color: rgba(160,120,20,0.60);
  background: rgba(160,120,20,0.08);
  box-shadow: 0 0 0 3px rgba(160,120,20,0.09), inset 0 1px 2px rgba(0,0,0,0.04);
}

/* Buttons — gold fill is unchanged; only text color adjusts */
[data-theme="light"] .greek-btn-primary { color: #1A0E00; }
[data-theme="light"] .greek-btn-primary:hover:not(:disabled) {
  box-shadow: 0 6px 24px rgba(160,120,20,0.28);
}
[data-theme="light"] .greek-btn-secondary {
  color: rgba(100,70,5,0.85);
  border-color: rgba(160,120,20,0.32);
}
[data-theme="light"] .greek-btn-secondary:hover:not(:disabled) {
  border-color: rgba(160,120,20,0.65);
  color: #5A3C05;
  background: rgba(160,120,20,0.07);
}
[data-theme="light"] .greek-btn-danger {
  color: rgba(160,40,40,0.85);
  border-color: rgba(160,40,40,0.35);
}
[data-theme="light"] .greek-btn-danger:hover:not(:disabled) {
  color: #9B2020;
  border-color: rgba(160,40,40,0.65);
  background: rgba(160,40,40,0.06);
}

/* Spinner inside primary button */
[data-theme="light"] .greek-btn-primary .greek-spinner {
  border-color: rgba(26,14,0,0.25);
  border-top-color: #1A0E00;
}

/* Badges — keep same hues, just slightly more saturated */
[data-theme="light"] .greek-badge-gold  { color: #7A5C00; border-color: rgba(160,120,0,0.40);  background: rgba(160,120,0,0.10); }
[data-theme="light"] .greek-badge-green { color: #1E6B2E; border-color: rgba(30,107,46,0.35);  background: rgba(30,107,46,0.08); }
[data-theme="light"] .greek-badge-red   { color: #9B2020; border-color: rgba(155,32,32,0.35);  background: rgba(155,32,32,0.07); }
[data-theme="light"] .greek-badge-blue  { color: #1A4D80; border-color: rgba(26,77,128,0.35);  background: rgba(26,77,128,0.07); }
[data-theme="light"] .greek-badge-muted { color: rgba(80,55,5,0.55); border-color: rgba(80,55,5,0.20); }

/* Alerts */
[data-theme="light"] .greek-alert-error   { background: rgba(180,50,50,0.07);  border-color: rgba(155,32,32,0.28);  color: #9B2020; }
[data-theme="light"] .greek-alert-success { background: rgba(30,107,46,0.07);  border-color: rgba(30,107,46,0.28);  color: #1E6B2E; }
[data-theme="light"] .greek-alert-info    { background: rgba(160,120,20,0.07); border-color: rgba(160,120,20,0.28); color: rgba(80,55,5,0.90); }
[data-theme="light"] .greek-alert-warning { background: rgba(180,120,10,0.08); border-color: rgba(180,120,10,0.30); color: rgba(100,65,0,0.90); }

/* Table */
[data-theme="light"] .greek-table thead tr { border-bottom-color: rgba(160,120,20,0.22); }
[data-theme="light"] .greek-table th       { color: rgba(80,55,5,0.65); }
[data-theme="light"] .greek-table td       { color: rgba(44,31,5,0.80); border-bottom-color: rgba(160,120,20,0.09); }
[data-theme="light"] .greek-table tbody tr:hover { background: rgba(160,120,20,0.04); }

/* Sidebar */
[data-theme="light"] .greek-sidebar {
  background: linear-gradient(180deg, #F0EAD8 0%, #EDE5CF 100%);
  border-right-color: rgba(160,120,20,0.18);
}
[data-theme="light"] .greek-sidebar-item       { color: rgba(80,55,5,0.55); }
[data-theme="light"] .greek-sidebar-item:hover { color: rgba(80,55,5,0.85); background: rgba(160,120,20,0.07); border-left-color: rgba(160,120,20,0.40); }
[data-theme="light"] .greek-sidebar-item.active{ color: #5A3C05; background: rgba(160,120,20,0.10); border-left-color: #A07814; }
[data-theme="light"] .greek-sidebar-section    { color: rgba(80,55,5,0.35); }

/* Navbar */
[data-theme="light"] .greek-navbar {
  background: rgba(245,240,232,0.92);
  border-bottom-color: rgba(160,120,20,0.18);
}

/* Loading screen */
[data-theme="light"] .greek-loading-screen { background: #F5F0E8; }
[data-theme="light"] .greek-loading-text   { color: rgba(80,55,5,0.55); }
[data-theme="light"] .greek-spinner        { border-color: rgba(160,120,20,0.20); border-top-color: #A07814; }

/* Scrollbar */
[data-theme="light"] ::-webkit-scrollbar-track { background: #EDE5CF; }
[data-theme="light"] ::-webkit-scrollbar-thumb { background: rgba(160,120,20,0.28); }
[data-theme="light"] ::-webkit-scrollbar-thumb:hover { background: rgba(160,120,20,0.50); }
`;

export const injectGreekStyles = () => {
  // Fonts
  if (!document.getElementById('greek-fonts')) {
    const link = document.createElement('link');
    link.id = 'greek-fonts';
    link.rel = 'stylesheet';
    link.href = 'https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&family=Inter:wght@300;400;500&display=swap';
    document.head.appendChild(link);
  }
  // CSS
  if (!document.getElementById('greek-theme-css')) {
    const style = document.createElement('style');
    style.id = 'greek-theme-css';
    style.textContent = GREEK_CSS;
    document.head.appendChild(style);
  }
};

// ── Theme persistence key ─────────────────────────────────────────────────────
export const THEME_KEY = 'greek-theme-mode';

// ── Read current theme ────────────────────────────────────────────────────────
export const getStoredTheme = (): 'light' | 'dark' =>
  (localStorage.getItem(THEME_KEY) as 'light' | 'dark') || 'dark';

// ── Apply theme to <html> (call this once on app boot in main.tsx / App.tsx) ──
export const applyTheme = (mode: 'light' | 'dark') => {
  localStorage.setItem(THEME_KEY, mode);
  document.documentElement.setAttribute('data-theme', mode);
};

// ── Toggle helper ─────────────────────────────────────────────────────────────
export const toggleTheme = () => {
  const next = getStoredTheme() === 'light' ? 'dark' : 'light';
  applyTheme(next);
  return next;
};

