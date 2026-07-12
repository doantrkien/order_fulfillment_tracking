import React, { useEffect } from 'react';
import { useOrders } from '../../hooks/useOrders';
import { DriverOrderCard } from '../../components/driver/DriverOrderCard';
import { injectGreekStyles } from '../../styles/greekTheme';

export const DriverOrdersPage: React.FC = () => {
  const { orders, pagination, isLoading, setPage } = useOrders();
  useEffect(() => { injectGreekStyles(); }, []);

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: '28px' }}>
        <h1 className="greek-heading" style={{ fontSize: '22px' }}>My Deliveries</h1>
        <div className="greek-divider" style={{ marginTop: '12px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      {isLoading ? (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '32px' }}>
          <span className="greek-spinner" />
          <span className="greek-muted">Loading deliveries…</span>
        </div>
      ) : orders.length === 0 ? (
        <div className="greek-card-static" style={{ padding: '40px', textAlign: 'center' }}>
          <span className="greek-muted">No deliveries assigned.</span>
        </div>
      ) : (
        <div>
          {orders.map(order => (
            <DriverOrderCard key={order.id} order={order} />
          ))}

          {pagination.totalPages > 1 && (
            <div className="flex justify-between items-center mt-4 mb-8">
              <button
                className="greek-btn-secondary"
                disabled={pagination.page <= 1}
                onClick={() => setPage(pagination.page - 1)}
              >
                Previous
              </button>
              <span className="greek-muted">Page {pagination.page} / {pagination.totalPages}</span>
              <button
                className="greek-btn-secondary"
                disabled={pagination.page >= pagination.totalPages}
                onClick={() => setPage(pagination.page + 1)}
              >
                Next
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
