import React, { useEffect } from 'react';
import { useOrders } from '../../hooks/useOrders';
import { OrderFilters } from '../../components/orders/OrderFilters';
import { OrderTable } from '../../components/orders/OrderTable';
import { injectGreekStyles } from '../../styles/greekTheme';

export const OrdersPage: React.FC = () => {
  const { orders, pagination, isLoading, setFilter, clearFilters, setPage } = useOrders();
  useEffect(() => { injectGreekStyles(); }, []);

  return (
    <div>
      {/* Page header */}
      <div style={{ marginBottom: '28px' }}>
        <div className="flex items-center justify-between">
          <h1 className="greek-heading" style={{ fontSize: '22px' }}>Orders Management</h1>
          <button className="greek-btn-primary" onClick={() => window.location.href = '/admin/orders/new'}>
            + Create Order
          </button>
        </div>
        <div className="greek-divider" style={{ marginTop: '12px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      <OrderFilters onFilterChange={setFilter} onClear={clearFilters} />
      <OrderTable orders={orders} isLoading={isLoading} pagination={pagination} onPageChange={setPage} baseRoute="/admin/orders" />
    </div>
  );
};
