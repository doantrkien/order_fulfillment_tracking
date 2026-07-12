import React, { useState } from 'react';
import type { OrderStatus } from '../../services/order.service';

interface OrderStatusModalProps {
  currentStatus: OrderStatus;
  onClose: () => void;
  onConfirm: (newStatus: OrderStatus, driverId?: number) => Promise<void> | void;
  allowedRoles?: ('admin' | 'driver')[];
}

const allStatuses: { label: string; value: OrderStatus }[] = [
  { label: 'Created',   value: 'created' },
  { label: 'Paid',      value: 'paid' },
  { label: 'Packed',    value: 'packed' },
  { label: 'Shipped',   value: 'shipped' },
  { label: 'Delivered', value: 'delivered' },
  { label: 'Cancelled', value: 'cancelled' },
  { label: 'Refunded',  value: 'refunded' },
];

export const OrderStatusModal: React.FC<OrderStatusModalProps> = ({
  currentStatus, onClose, onConfirm, allowedRoles = ['admin'],
}) => {
  const [status, setStatus] = useState<OrderStatus>(currentStatus);
  const [driverId, setDriverId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const isAdmin = allowedRoles.includes('admin');
  const selectableStatuses = isAdmin
    ? allStatuses
    : allStatuses.filter(s => ['shipped', 'delivered'].includes(s.value));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (status === currentStatus) { setError('Please select a different status'); return; }
    setError('');
    setIsSubmitting(true);
    try {
      await onConfirm(status, driverId ? parseInt(driverId, 10) : undefined);
    } catch (err: any) {
      setError(err.message || 'Failed to update status');
      setIsSubmitting(false);
    }
  };

  return (
    <div className="greek-modal-overlay" onClick={onClose}>
      <div className="greek-modal" onClick={e => e.stopPropagation()}>
        <div className="greek-modal-header">
          <h2 className="greek-subheading">Update Order Status</h2>
          <button className="greek-close-btn" onClick={onClose}>×</button>
        </div>

        <div className="greek-modal-body">
          <form onSubmit={handleSubmit}>
            {error && <div className="greek-alert greek-alert-error">{error}</div>}

            <div style={{ marginBottom: '16px' }}>
              <label className="greek-label" htmlFor="statusSelect">New Status</label>
              <select
                id="statusSelect"
                className="greek-input greek-select"
                value={status}
                onChange={e => setStatus(e.target.value as OrderStatus)}
              >
                {selectableStatuses.map(s => (
                  <option key={s.value} value={s.value}>{s.label}</option>
                ))}
              </select>
            </div>

            {isAdmin && (
              <div style={{ marginBottom: '16px' }}>
                <label className="greek-label" htmlFor="driverId">Assign Driver (Optional)</label>
                <input
                  id="driverId"
                  type="number"
                  className="greek-input"
                  placeholder="Driver ID (e.g. 42)"
                  value={driverId}
                  onChange={e => setDriverId(e.target.value)}
                />
                <p className="greek-muted" style={{ marginTop: '6px' }}>
                  Required when setting status to shipped / delivered.
                </p>
              </div>
            )}

            <div className="greek-modal-footer" style={{ padding: '16px 0 0', border: 'none' }}>
              <button type="button" className="greek-btn-secondary" onClick={onClose} disabled={isSubmitting}>
                Cancel
              </button>
              <button type="submit" className="greek-btn-primary" disabled={isSubmitting || status === currentStatus}>
                {isSubmitting ? <><span className="greek-spinner" />Updating…</> : 'Confirm Update'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
