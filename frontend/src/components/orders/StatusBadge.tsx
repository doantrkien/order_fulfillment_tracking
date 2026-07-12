import React from 'react';
import type { OrderStatus } from '../../services/order.service';

interface StatusBadgeProps {
  status: OrderStatus;
}

const getGreekBadgeClass = (s: OrderStatus): string => {
  switch (s) {
    case 'delivered':
      return 'greek-badge greek-badge-green';
    case 'paid':
    case 'shipped':
    case 'packed':
      return 'greek-badge greek-badge-gold';
    case 'cancelled':
    case 'refunded':
      return 'greek-badge greek-badge-red';
    case 'created':
    default:
      return 'greek-badge greek-badge-muted';
  }
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => (
  <span className={getGreekBadgeClass(status)}>{status}</span>
);
