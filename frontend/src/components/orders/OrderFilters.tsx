import React, { useState } from 'react';
import type { OrderQuery, OrderStatus } from '../../services/order.service';

interface OrderFiltersProps {
  initialFilters?: Partial<OrderQuery>;
  onFilterChange: (filters: Partial<OrderQuery>) => void;
  onClear: () => void;
}

const statuses: { label: string; value: OrderStatus | '' }[] = [
  { label: 'All Statuses', value: '' },
  { label: 'Created', value: 'created' },
  { label: 'Paid', value: 'paid' },
  { label: 'Shipped', value: 'shipped' },
  { label: 'Delivered', value: 'delivered' },
  { label: 'Cancelled', value: 'cancelled' },
  { label: 'Refunded', value: 'refunded' },
];

export const OrderFilters: React.FC<OrderFiltersProps> = ({ initialFilters, onFilterChange, onClear }) => {
  const [status, setStatus] = useState<string>(initialFilters?.status || '');
  const [date, setDate] = useState<string>(initialFilters?.date || '');

  const handleApply = () => onFilterChange({ status: status || undefined, date: date || undefined });
  const handleClear = () => { setStatus(''); setDate(''); onClear(); };

  return (
    <div className="greek-card-static mb-4" style={{ padding: '16px 24px' }}>
      <div className="flex items-end gap-4 flex-wrap">
        <div style={{ minWidth: '200px' }}>
          <label className="greek-label">Status</label>
          <select
            className="greek-input greek-select"
            value={status}
            onChange={e => setStatus(e.target.value)}
          >
            {statuses.map(s => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
        </div>

        <div style={{ minWidth: '200px' }}>
          <label className="greek-label">Date</label>
          <input
            type="date"
            className="greek-input"
            value={date}
            onChange={e => setDate(e.target.value)}
          />
        </div>

        <div className="flex gap-2">
          <button className="greek-btn-primary" onClick={handleApply}>
            Apply
          </button>
          <button className="greek-btn-secondary" onClick={handleClear}>
            Clear
          </button>
        </div>
      </div>
    </div>
  );
};
