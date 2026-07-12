import React, { useState, useEffect } from 'react';
import { reportService, type DailyReportResponse } from '../../services/report.service';
import { injectGreekStyles } from '../../styles/greekTheme';
import { useTheme } from '../../hooks/useTheme';

// ── Stat configs ─────────────────────────────────────────────────────────────
type StatConfig = {
  label: string;
  key: keyof DailyReportResponse | 'avg_deliver_time_fmt';
  valueColor: string;
  borderColor: string;
  bgColor: string;
};

// ── Component ────────────────────────────────────────────────────────────────
export const DailyReportPage: React.FC = () => {
  const { isLight } = useTheme();

  const STATS: StatConfig[] = [
    {
      label: 'Total Orders',
      key: 'total_orders',
      valueColor: isLight ? '#C9920A' : '#F0C040',
      borderColor: isLight ? 'rgba(160,120,20,0.28)' : 'rgba(201,168,76,0.25)',
      bgColor: isLight ? 'rgba(160,120,20,0.05)' : 'rgba(201,168,76,0.05)',
    },
    {
      label: 'Total Income',
      key: 'total_income',
      valueColor: isLight ? '#1E6B2E' : '#7EC88A',
      borderColor: isLight ? 'rgba(30,107,46,0.25)' : 'rgba(126,200,138,0.25)',
      bgColor: isLight ? 'rgba(30,107,46,0.05)' : 'rgba(126,200,138,0.05)',
    },
    {
      label: 'Delivered',
      key: 'total_delivered',
      valueColor: isLight ? '#1E6B2E' : '#7EC88A',
      borderColor: isLight ? 'rgba(30,107,46,0.25)' : 'rgba(126,200,138,0.25)',
      bgColor: isLight ? 'rgba(30,107,46,0.05)' : 'rgba(126,200,138,0.05)',
    },
    {
      label: 'Cancelled',
      key: 'total_cancelled',
      valueColor: isLight ? '#9B2020' : '#E88080',
      borderColor: isLight ? 'rgba(155,32,32,0.25)' : 'rgba(232,128,128,0.25)',
      bgColor: isLight ? 'rgba(155,32,32,0.05)' : 'rgba(232,128,128,0.05)',
    },
    {
      label: 'Refunded',
      key: 'total_refunded',
      valueColor: isLight ? '#C9920A' : '#F0C040',
      borderColor: isLight ? 'rgba(160,120,20,0.28)' : 'rgba(240,192,64,0.25)',
      bgColor: isLight ? 'rgba(160,120,20,0.05)' : 'rgba(240,192,64,0.05)',
    },
    {
      label: 'Avg Deliver Time',
      key: 'avg_deliver_time_fmt',
      valueColor: isLight ? '#1A4D80' : '#80B4E8',
      borderColor: isLight ? 'rgba(26,77,128,0.25)' : 'rgba(128,180,232,0.25)',
      bgColor: isLight ? 'rgba(26,77,128,0.05)' : 'rgba(128,180,232,0.05)',
    },
  ];
  const today = new Date().toISOString().split('T')[0];
  const [date, setDate] = useState(today);
  const [report, setReport] = useState<DailyReportResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => { injectGreekStyles(); }, []);

  const fetchReport = async (targetDate: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await reportService.getDailyReport(targetDate);
      setReport(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch report');
      setReport(null);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchReport(date); }, [date]);

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount);

  const getStatValue = (stat: StatConfig, r: DailyReportResponse): string => {
    if (stat.key === 'total_income') return formatCurrency(r.total_income);
    if (stat.key === 'avg_deliver_time_fmt') return `${r.avg_deliver_time.toFixed(1)} hrs`;
    return String(r[stat.key as keyof DailyReportResponse]);
  };

  // Formatted display date
  const displayDate = new Date(date + 'T00:00:00').toLocaleDateString('en-GB', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
  });

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 36 }}>
          <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 20, flexWrap: 'wrap' }}>
            <div>
              <h1 className="greek-heading" style={{ fontSize: 24, margin: 0 }}>Daily Report</h1>
              <p className="greek-muted" style={{ marginTop: 5 }}>
                {displayDate}
              </p>
            </div>

            {/* Date picker */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <label className="greek-label" htmlFor="report-date" style={{ marginBottom: 0, whiteSpace: 'nowrap' }}>
                Select Date
              </label>
              <input
                id="report-date"
                type="date"
                className="greek-input"
                style={{ width: 'auto', minWidth: 160 }}
                value={date}
                max={today}
                onChange={e => setDate(e.target.value)}
              />
            </div>
          </div>

          <div className="greek-divider" style={{ marginTop: 16 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── States ── */}
        {isLoading ? (
          <div style={{ padding: '60px 0', textAlign: 'center', animation: 'greek-fade-in 0.3s ease both' }}>
            <span className="greek-spinner" style={{ width: 20, height: 20 }} />
            <span className="greek-loading-text" style={{ fontSize: 11 }}>Loading report data…</span>
          </div>

        ) : error ? (
          <div className="greek-alert greek-alert-error">{error}</div>

        ) : !report ? (
          <div className="greek-card-static" style={{ padding: '48px', textAlign: 'center' }}>
            <svg width="36" height="36" viewBox="0 0 36 36" fill="none" style={{ margin: '0 auto 16px', display: 'block', opacity: 0.3 }}>
              <circle cx="18" cy="18" r="16" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.2" />
              <path d="M18 12v7M18 23v2" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.5" strokeLinecap="round" />
            </svg>
            <p className="greek-muted" style={{ margin: 0 }}>No data available for this date.</p>
          </div>

        ) : (
          <>
            {/* ── Stat grid ── */}
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: 16,
              marginBottom: 32,
            }}>
              {STATS.map((stat, i) => (
                <div
                  key={stat.label}
                  className="greek-card-static"
                  style={{
                    position: 'relative',
                    border: `1px solid ${stat.borderColor}`,
                    borderRadius: 4,
                    padding: '24px 20px 20px',
                    animation: `greek-fade-in 0.4s ease both`,
                    animationDelay: `${i * 60}ms`,
                    transition: 'border-color 0.25s, box-shadow 0.25s',
                  }}
                  onMouseEnter={e => {
                    (e.currentTarget as HTMLDivElement).style.borderColor = stat.valueColor + '55';
                    (e.currentTarget as HTMLDivElement).style.boxShadow = `0 4px 20px ${stat.bgColor}`;
                  }}
                  onMouseLeave={e => {
                    (e.currentTarget as HTMLDivElement).style.borderColor = stat.borderColor;
                    (e.currentTarget as HTMLDivElement).style.boxShadow = 'none';
                  }}
                >
                  {/* Icon + label row */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
                    <span className="greek-label" style={{
                      fontFamily: "'Cinzel', serif",
                      fontSize: 9,
                      fontWeight: 600,
                      letterSpacing: '0.22em',
                      textTransform: 'uppercase' as const,
                    }}>
                      {stat.label}
                    </span>
                  </div>

                  {/* Value */}
                  <div style={{
                    fontFamily: "'Cinzel', serif",
                    fontSize: 28,
                    fontWeight: 600,
                    color: stat.valueColor,
                    lineHeight: 1,
                  }}>
                    {getStatValue(stat, report)}
                  </div>

                  {/* Bottom accent line */}
                  <div style={{
                    position: 'absolute',
                    bottom: 0, left: 0, right: 0,
                    height: 2,
                    background: `linear-gradient(90deg, transparent, ${stat.valueColor}33, transparent)`,
                    borderRadius: '0 0 4px 4px',
                  }} />
                </div>
              ))}
            </div>

            {/* ── Summary card ── */}
            <div className="greek-card greek-card-corners">
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 20 }}>
                <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                  <path d="M9 1L17 9L9 17L1 9Z" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.2" fill={isLight ? "rgba(160,120,20,0.08)" : "rgba(201,168,76,0.08)"} />
                </svg>
                <h3 className="greek-subheading" style={{ margin: 0 }}>Summary</h3>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px 32px' }}>
                {[
                  { label: 'Completion Rate', value: report.total_orders > 0 ? `${((report.total_delivered / report.total_orders) * 100).toFixed(1)}%` : '—' },
                  { label: 'Cancellation Rate', value: report.total_orders > 0 ? `${((report.total_cancelled / report.total_orders) * 100).toFixed(1)}%` : '—' },
                  { label: 'Refund Rate', value: report.total_orders > 0 ? `${((report.total_refunded / report.total_orders) * 100).toFixed(1)}%` : '—' },
                  { label: 'Revenue per Order', value: report.total_delivered > 0 ? formatCurrency(report.total_income / report.total_delivered) : '—' },
                ].map(row => (
                  <div key={row.label} style={{ padding: '12px 0', borderBottom: '1px solid rgba(201,168,76,0.07)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span className="greek-muted">{row.label}</span>
                    <span className="greek-text" style={{ fontFamily: "'Inter', sans-serif", fontSize: 14, fontWeight: 500 }}>
                      {row.value}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}

      </div>
    </div>
  );
};