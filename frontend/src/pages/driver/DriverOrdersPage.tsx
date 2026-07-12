import React, { useEffect } from 'react';
import { useOrders } from '../../hooks/useOrders';
import { DriverOrderCard } from '../../components/driver/DriverOrderCard';
import { injectGreekStyles } from '../../styles/greekTheme';
import { useTheme } from '../../hooks/useTheme';

const IconTruck: React.FC<{ isLight?: boolean }> = ({ isLight }) => {
  const color = isLight ? "#A07814" : "#C9A84C";
  return (
    <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
      <path d="M1 4h11v9H1zM12 7h4l3 3v3h-7V7z" stroke={color} strokeWidth="1.2" strokeLinejoin="round" fill="none" opacity="0.6" />
      <circle cx="4.5" cy="14.5" r="1.5" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
      <circle cx="14.5" cy="14.5" r="1.5" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
    </svg>
  );
};

export const DriverOrdersPage: React.FC = () => {
  const { isLight } = useTheme();
  const { orders, pagination, isLoading, setPage } = useOrders();
  useEffect(() => { injectGreekStyles(); }, []);

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 36 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
            <IconTruck isLight={isLight} />
            <h1 className="greek-heading" style={{ fontSize: 22, margin: 0 }}>My Deliveries</h1>
          </div>
          <p className="greek-muted">
            {isLoading
              ? 'Loading your assigned orders…'
              : orders.length > 0
                ? `${orders.length} order${orders.length !== 1 ? 's' : ''} assigned to you`
                : 'No deliveries currently assigned'}
          </p>
          <div className="greek-divider" style={{ marginTop: 14 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── States ── */}
        {isLoading ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '48px 0' }}>
            <span className="greek-spinner" style={{ width: 20, height: 20 }} />
            <span className="greek-loading-text" style={{ fontSize: 11 }}>Loading deliveries…</span>
          </div>

        ) : orders.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '64px 0' }}>
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none" style={{ margin: '0 auto 20px', display: 'block', opacity: 0.25 }}>
              <rect x="4" y="14" width="28" height="22" rx="2" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.4" fill="none" />
              <path d="M32 20h8l4 6v8h-12V20z" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.4" strokeLinejoin="round" fill="none" />
              <circle cx="12" cy="37" r="3.5" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.2" fill="none" />
              <circle cx="38" cy="37" r="3.5" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.2" fill="none" />
            </svg>
            <p className="greek-muted" style={{ fontSize: 14, margin: 0 }}>No deliveries assigned to you yet.</p>
            <p className="greek-caption" style={{ marginTop: 8 }}>Check back later or contact your dispatcher.</p>
          </div>

        ) : (
          <div>
            {/* Order cards with stagger */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {orders.map((order, i) => (
                <div
                  key={order.id}
                  style={{ animation: 'greek-fade-in 0.35s ease both', animationDelay: `${i * 55}ms` }}
                >
                  <DriverOrderCard order={order} />
                </div>
              ))}
            </div>

            {/* Pagination */}
            {pagination.totalPages > 1 && (
              <div className="greek-card-static" style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginTop: 28,
                padding: '16px 20px',
                borderRadius: 4,
              }}>
                <button
                  className="greek-btn-secondary"
                  disabled={pagination.page <= 1}
                  onClick={() => setPage(pagination.page - 1)}
                >
                  ← Previous
                </button>

                <div style={{ textAlign: 'center' }}>
                  <span style={{ fontFamily: "'Cinzel', serif", fontSize: 11, fontWeight: 600, color: isLight ? '#A07814' : '#C9A84C', letterSpacing: '0.15em' }}>
                    {pagination.page}
                  </span>
                  <span className="greek-caption" style={{ margin: '0 8px' }}>/</span>
                  <span className="greek-caption">{pagination.totalPages}</span>
                </div>

                <button
                  className="greek-btn-secondary"
                  disabled={pagination.page >= pagination.totalPages}
                  onClick={() => setPage(pagination.page + 1)}
                >
                  Next →
                </button>
              </div>
            )}
          </div>
        )}

      </div>
    </div>
  );
};