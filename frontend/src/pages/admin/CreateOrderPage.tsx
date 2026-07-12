import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { orderService } from '../../services/order.service';
import { toast } from 'react-hot-toast';
import { injectGreekStyles } from '../../styles/greekTheme';

const FIELDS = [
  { id: 'username', label: 'Customer Name', type: 'text', placeholder: 'Nguyen Van A' },
  { id: 'user_phone', label: 'Phone Number', type: 'text', placeholder: '0912 345 678' },
  { id: 'shipping_address', label: 'Shipping Address', type: 'text', placeholder: '123 ABC Street, HCMC' },
  { id: 'total_amount', label: 'Total Amount (VND)', type: 'number', placeholder: '150,000' },
] as const;

export const CreateOrderPage: React.FC = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    username: '', user_phone: '', shipping_address: '', total_amount: '',
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => { injectGreekStyles(); }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    try {
      const amount = parseInt(formData.total_amount, 10);
      if (isNaN(amount) || amount <= 0) throw new Error('Total amount must be a positive number');
      const newOrder = await orderService.createOrder({
        username: formData.username,
        user_phone: formData.user_phone,
        shipping_address: formData.shipping_address,
        total_amount: amount,
      });
      toast.success('Order created successfully!');
      navigate(`/admin/orders/${newOrder.id}`);
    } catch (err: any) {
      setError(err.message || 'Failed to create order');
      toast.error(err.message || 'Failed to create order');
      setIsSubmitting(false);
    }
  };

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 36 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 14 }}>
            <button
              className="greek-btn-secondary"
              onClick={() => navigate('/admin/orders')}
            >
              ← Back
            </button>

            <div>
              <h1 className="greek-heading" style={{ fontSize: 22, margin: 0 }}>
                Create New Order
              </h1>
              <p className="greek-muted" style={{ marginTop: 4 }}>
                Fill in the details below to register a new order
              </p>
            </div>
          </div>

          <div className="greek-divider">
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── Form card ── */}
        <div className="greek-card greek-card-corners" style={{ maxWidth: 560 }}>

          {/* Card title */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 28 }}>
            {/* Diamond icon */}
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path d="M9 1L17 9L9 17L1 9Z" stroke="#C9A84C" strokeWidth="1.2" fill="rgba(201,168,76,0.08)" />
            </svg>
            <h3 className="greek-subheading" style={{ margin: 0 }}>Order Information</h3>
          </div>

          <form onSubmit={handleSubmit} autoComplete="off">
            {error && (
              <div className="greek-alert greek-alert-error" style={{ marginBottom: 24 }}>
                {error}
              </div>
            )}

            {FIELDS.map((field, i) => (
              <div
                key={field.id}
                className="greek-field"
                style={{
                  marginBottom: i < FIELDS.length - 1 ? 20 : 28,
                }}
              >
                <label className="greek-label" htmlFor={field.id}>
                  {field.label}
                </label>
                <input
                  id={field.id}
                  name={field.id}
                  type={field.type}
                  className="greek-input"
                  placeholder={field.placeholder}
                  value={(formData as any)[field.id]}
                  onChange={handleChange}
                  required
                  min={field.type === 'number' ? '1' : undefined}
                />
              </div>
            ))}

            {/* Divider before actions */}
            <div className="greek-divider" style={{ margin: '0 0 24px' }}>
              <div className="greek-divider-line" />
              <div className="greek-divider-diamond" />
              <div className="greek-divider-line" />
            </div>

            {/* Action buttons */}
            <div style={{ display: 'flex', gap: 12 }}>
              <button
                type="button"
                className="greek-btn-secondary"
                style={{ flex: 1 }}
                onClick={() => navigate('/admin/orders')}
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="greek-btn-primary"
                style={{ flex: 2 }}
                disabled={isSubmitting}
              >
                {isSubmitting
                  ? <><span className="greek-spinner" />Creating…</>
                  : 'Create Order'
                }
              </button>
            </div>
          </form>
        </div>

      </div>
    </div>
  );
};