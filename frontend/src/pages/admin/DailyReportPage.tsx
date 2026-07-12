import React, { useState, useEffect } from 'react';
import { reportService, type DailyReportResponse } from '../../services/report.service';
import { injectGreekStyles } from '../../styles/greekTheme';

export const DailyReportPage: React.FC = () => {
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

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: '28px' }}>
        <div className="flex items-center justify-between">
          <h1 className="greek-heading" style={{ fontSize: '22px' }}>Daily Report</h1>
          <div className="flex items-center gap-2">
            <label className="greek-label" htmlFor="report-date" style={{ marginBottom: 0 }}>Date:</label>
            <input
              id="report-date"
              type="date"
              className="greek-input"
              style={{ width: 'auto' }}
              value={date}
              onChange={e => setDate(e.target.value)}
            />
          </div>
        </div>
        <div className="greek-divider" style={{ marginTop: '12px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      {isLoading ? (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '32px' }}>
          <span className="greek-spinner" />
          <span className="greek-muted">Loading report data…</span>
        </div>
      ) : error ? (
        <div className="greek-alert greek-alert-error">{error}</div>
      ) : !report ? (
        <div className="greek-card-static" style={{ padding: '40px', textAlign: 'center' }}>
          <span className="greek-muted">No data available for this date.</span>
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
          {[
            { label: 'Total Orders',           value: String(report.total_orders),                cls: 'greek-stat-value' },
            { label: 'Total Income',            value: formatCurrency(report.total_income),        cls: 'greek-stat-value-success' },
            { label: 'Delivered',               value: String(report.total_delivered),             cls: 'greek-stat-value-success' },
            { label: 'Cancelled',               value: String(report.total_canceled),              cls: 'greek-stat-value-danger' },
            { label: 'Refunded',                value: String(report.total_refunded),              cls: 'greek-stat-value-warning' },
            { label: 'Avg Deliver Time (hrs)',  value: report.avg_deliver_time.toFixed(1),         cls: 'greek-stat-value-neutral' },
          ].map(stat => (
            <div key={stat.label} className="greek-card-static" style={{ textAlign: 'center' }}>
              <p className="greek-muted" style={{ marginBottom: '8px' }}>{stat.label}</p>
              <p className={stat.cls} style={{ fontSize: stat.cls === 'greek-stat-value-success' && stat.label === 'Total Income' ? '18px' : undefined }}>
                {stat.value}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
