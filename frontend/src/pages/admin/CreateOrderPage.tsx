import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { orderService } from '../../services/order.service';
import { toast } from 'react-hot-toast';
import { injectGreekStyles } from '../../styles/greekTheme';

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
    <div>
      {/* Header */}
      <div style={{ marginBottom: '28px' }}>
        <div className="flex items-center gap-4">
          <button className="greek-btn-secondary" onClick={() => navigate('/admin/orders')}>← Back</button>
          <h1 className="greek-heading" style={{ fontSize: '22px' }}>Create New Order</h1>
        </div>
        <div className="greek-divider" style={{ marginTop: '12px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      <div className="greek-card-static" style={{ maxWidth: '600px' }}>
        <form onSubmit={handleSubmit}>
          {error && <div className="greek-alert greek-alert-error">{error}</div>}

          {[
            { id: 'username',         label: 'Customer Name',     type: 'text',   placeholder: 'Nguyen Van A' },
            { id: 'user_phone',       label: 'Phone Number',      type: 'text',   placeholder: '0912345678' },
            { id: 'shipping_address', label: 'Shipping Address',  type: 'text',   placeholder: '123 ABC Street, HCMC' },
            { id: 'total_amount',     label: 'Total Amount (VND)', type: 'number', placeholder: '150000' },
          ].map(field => (
            <div key={field.id} style={{ marginBottom: '16px' }}>
              <label className="greek-label" htmlFor={field.id}>{field.label}</label>
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

          <button type="submit" className="greek-btn-primary" style={{ width: '100%', marginTop: '8px' }} disabled={isSubmitting}>
            {isSubmitting ? <><span className="greek-spinner" />Creating…</> : 'Create Order'}
          </button>
        </form>
      </div>
    </div>
  );
};
