import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { Order } from '../../services/order.service';
import { StatusBadge } from './StatusBadge';

interface PaginationData {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

interface OrderTableProps {
  orders: Order[];
  isLoading: boolean;
  pagination: PaginationData;
  onPageChange: (page: number) => void;
  baseRoute: string;
}

const formatCurrency = (amount: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount);

const formatDate = (dateString: string) =>
  new Date(dateString).toLocaleDateString('vi-VN', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit',
  });


export const OrderTable: React.FC<OrderTableProps> = ({ orders, isLoading, pagination, onPageChange, baseRoute }) => {
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="greek-card-static" style={{ padding: '40px', textAlign: 'center' }}>
        <span className="greek-spinner" />
        <span className="greek-muted">Loading orders…</span>
      </div>
    );
  }

  if (!orders || orders.length === 0) {
    return (
      <div className="greek-card-static" style={{ padding: '40px', textAlign: 'center' }}>
        <span className="greek-muted">No orders found matching your criteria.</span>
      </div>
    );
  }

  return (
    <div className="greek-card-static" style={{ padding: 0, overflow: 'hidden' }}>
      <div className="table-responsive">
        <table className="greek-table">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Customer</th>
              <th>Total Amount</th>
              <th>Date</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {orders.map((order, i) => (
              <tr
                key={order.id}
                onClick={() => navigate(`${baseRoute}/${order.id}`)}
                style={{ cursor: 'pointer', animationDelay: `${i * 40}ms` }}
              >
                <td><span style={{ fontWeight: 600, color: '#C9A84C' }}>#{order.id}</span></td>
                <td>
                  <div>{order.username}</div>
                  <div className="greek-muted">{order.user_phone}</div>
                </td>
                <td style={{ fontWeight: 600 }}>{formatCurrency(order.total_amount)}</td>
                <td className="greek-muted">{formatDate(order.ordered_at)}</td>
                <td><StatusBadge status={order.status} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination.totalPages > 1 && (
        <div style={{
          padding: '14px 20px',
          borderTop: '1px solid rgba(201,168,76,0.12)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <span className="greek-muted">
            Page {pagination.page} of {pagination.totalPages} &nbsp;·&nbsp; {pagination.totalItems} items
          </span>
          <div className="flex gap-2">
            <button
              className="greek-btn-secondary"
              style={{ padding: '8px 16px' }}
              disabled={pagination.page <= 1}
              onClick={() => onPageChange(pagination.page - 1)}
            >
              Previous
            </button>
            <button
              className="greek-btn-secondary"
              style={{ padding: '8px 16px' }}
              disabled={pagination.page >= pagination.totalPages}
              onClick={() => onPageChange(pagination.page + 1)}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
