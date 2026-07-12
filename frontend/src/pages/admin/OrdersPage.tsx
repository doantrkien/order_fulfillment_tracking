import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useOrders } from '../../hooks/useOrders';
import { OrderFilters } from '../../components/orders/OrderFilters';
import { OrderTable } from '../../components/orders/OrderTable';
import { injectGreekStyles } from '../../styles/greekTheme';

export const OrdersPage: React.FC = () => {
  const { orders, pagination, isLoading, setFilter, clearFilters, setPage } = useOrders();
  const navigate = useNavigate();

  useEffect(() => { injectGreekStyles(); }, []);

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 36 }}>
          <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>
            <div>
              <h1 className="greek-heading" style={{ fontSize: 22, margin: 0 }}>Orders Management</h1>
              <p className="greek-muted" style={{ marginTop: 5 }}>
                {pagination?.totalItems != null
                  ? `${pagination.totalItems} orders total`
                  : 'Manage and track all customer orders'}
              </p>
            </div>
            <button
              className="greek-btn-primary"
              onClick={() => navigate('/admin/orders/new')}
            >
              + Create Order
            </button>
          </div>

          <div className="greek-divider" style={{ marginTop: 16 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── Filters ── */}
        <div style={{ marginBottom: 20 }}>
          <OrderFilters onFilterChange={setFilter} onClear={clearFilters} />
        </div>

        {/* ── Table ── */}
        <div style={{ animation: 'greek-fade-in 0.4s ease both' }}>
          <OrderTable
            orders={orders}
            isLoading={isLoading}
            pagination={pagination}
            onPageChange={setPage}
            baseRoute="/admin/orders"
          />
        </div>

      </div>
    </div>
  );
};