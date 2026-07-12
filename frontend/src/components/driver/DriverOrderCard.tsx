import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { Order } from '../../services/order.service';
import { StatusBadge } from '../orders/StatusBadge';

interface DriverOrderCardProps {
  order: Order;
}

export const DriverOrderCard: React.FC<DriverOrderCardProps> = ({ order }) => {
  const navigate = useNavigate();

  return (
    <div
      className="greek-card-static mb-4"
      onClick={() => navigate(`/driver/orders/${order.id}`)}
      style={{ cursor: 'pointer', borderLeft: '3px solid rgba(201,168,76,0.45)' }}
    >
      <div className="flex justify-between items-start mb-2">
        <div>
          <h3 className="greek-subheading">Order #{order.id}</h3>
          <p className="greek-muted" style={{ marginTop: '4px' }}>
            {new Date(order.ordered_at).toLocaleDateString('vi-VN')}
          </p>
        </div>
        <StatusBadge status={order.status} />
      </div>

      <div style={{ marginTop: '16px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
        <div className="flex items-center gap-2">
          <span className="greek-muted" style={{ minWidth: '80px' }}>Customer:</span>
          <span className="greek-text" style={{ fontWeight: 500 }}>{order.username}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="greek-muted" style={{ minWidth: '80px' }}>Phone:</span>
          <span className="greek-text">{order.user_phone}</span>
        </div>
        <div className="flex items-start gap-2">
          <span className="greek-muted" style={{ minWidth: '80px' }}>Address:</span>
          <span className="greek-text">{order.shipping_address}</span>
        </div>
      </div>
    </div>
  );
};
