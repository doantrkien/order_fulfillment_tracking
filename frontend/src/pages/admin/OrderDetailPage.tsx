import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { orderService, type Order, type OrderStatus } from '../../services/order.service';
import { StatusBadge } from '../../components/orders/StatusBadge';
import { OrderStatusModal } from '../../components/orders/OrderStatusModal';
import { aiService, type AIExceptionAnalysis } from '../../services/ai.service';
import { ExceptionPanel } from '../../components/ai/ExceptionPanel';
import { DraftPanel } from '../../components/ai/DraftPanel';
import { toast } from 'react-hot-toast';
import { injectGreekStyles } from '../../styles/greekTheme';

const FIELD_LABEL: React.CSSProperties = {
  fontFamily: "'Cinzel', serif",
  fontSize: '9px',
  fontWeight: 600,
  letterSpacing: '0.22em',
  textTransform: 'uppercase',
  color: 'rgba(201,168,76,0.50)',
  marginBottom: '4px',
};
const FIELD_VALUE: React.CSSProperties = {
  fontFamily: "'Inter', sans-serif",
  fontSize: '14px',
  color: 'rgba(232,213,163,0.85)',
};

export const OrderDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [order, setOrder] = useState<Order | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [aiAnalysis, setAiAnalysis] = useState<AIExceptionAnalysis | null>(null);
  const [isAiLoading, setIsAiLoading] = useState(false);

  useEffect(() => { injectGreekStyles(); }, []);

  const fetchOrder = async () => {
    if (!id) return;
    setIsLoading(true);
    try {
      const data = await orderService.getOrder(parseInt(id, 10));
      setOrder(data);
      try {
        const aiData = await aiService.getLatestInsights(data.id);
        setAiAnalysis(aiData);
      } catch { setAiAnalysis(null); }
    } catch (err: any) {
      setError(err.message || 'Failed to fetch order details');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchOrder(); }, [id]);

  const handleUpdateStatus = async (newStatus: OrderStatus, driverId?: number) => {
    if (!id) return;
    await orderService.updateStatus(parseInt(id, 10), { status: newStatus, driver_id: driverId });
    fetchOrder();
    setIsModalOpen(false);
  };

  const handleTriggerAiAnalysis = async () => {
    if (!order) return;
    setIsAiLoading(true);
    try {
      const analysis = await aiService.analyzeException(order.id);
      setAiAnalysis(analysis);
      toast.success('AI Analysis complete!');
    } catch (err: any) {
      toast.error(err.message || 'Failed to run AI Analysis');
    } finally {
      setIsAiLoading(false);
    }
  };

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount);
  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleString('vi-VN');

  if (isLoading) {
    return (
      <div className="greek-loading-screen">
        <span className="greek-loading-text">Loading Order…</span>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="greek-card-static" style={{ padding: '40px', textAlign: 'center' }}>
        <p className="greek-muted">{error || 'Order not found'}</p>
        <button className="greek-btn-secondary" style={{ marginTop: '16px' }} onClick={() => navigate('/admin/orders')}>
          Back to Orders
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: '28px' }}>
        <div className="flex items-center gap-4">
          <button className="greek-btn-secondary" onClick={() => navigate('/admin/orders')}>← Back</button>
          <div>
            <h1 className="greek-heading" style={{ fontSize: '22px' }}>Order #{order.id}</h1>
          </div>
          <StatusBadge status={order.status} />
        </div>
        <button className="greek-btn-primary" onClick={() => setIsModalOpen(true)}>
          Update Status
        </button>
      </div>

      {/* Customer info */}
      <div className="greek-card-static mb-4">
        <h3 className="greek-subheading" style={{ marginBottom: '16px' }}>Customer Information</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
          <div><p style={FIELD_LABEL}>Name</p><p style={FIELD_VALUE}>{order.username}</p></div>
          <div><p style={FIELD_LABEL}>Phone</p><p style={FIELD_VALUE}>{order.user_phone}</p></div>
          <div style={{ gridColumn: '1 / -1' }}>
            <p style={FIELD_LABEL}>Shipping Address</p>
            <p style={FIELD_VALUE}>{order.shipping_address}</p>
          </div>
        </div>
      </div>

      {/* Order summary */}
      <div className="greek-card-static mb-4">
        <h3 className="greek-subheading" style={{ marginBottom: '16px' }}>Order Summary</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
          <div>
            <p style={FIELD_LABEL}>Total Amount</p>
            <p className="greek-stat-value" style={{ fontSize: '22px' }}>{formatCurrency(order.total_amount)}</p>
          </div>
          <div>
            <p style={FIELD_LABEL}>Ordered At</p>
            <p style={FIELD_VALUE}>{formatDate(order.ordered_at)}</p>
          </div>
          <div>
            <p style={FIELD_LABEL}>Status</p>
            <div style={{ marginTop: '4px' }}>
              <StatusBadge status={order.status} />
            </div>
          </div>
          {order.driver_note && (
            <div style={{ gridColumn: '1 / -1' }}>
              <p style={FIELD_LABEL}>Driver Note</p>
              <div style={{
                padding: '10px 14px',
                background: 'rgba(201,168,76,0.07)',
                borderLeft: '3px solid rgba(201,168,76,0.45)',
                borderRadius: '4px',
              }}>
                <p style={{ ...FIELD_VALUE, fontStyle: 'italic' }}>{order.driver_note}</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Event history placeholder */}
      <div className="greek-card-static mb-4">
        <h3 className="greek-subheading" style={{ marginBottom: '12px' }}>Event History</h3>
        <p className="greek-muted" style={{ textAlign: 'center', padding: '24px 0', fontStyle: 'italic' }}>
          Event history data is not yet available.{' '}
          <code style={{ fontSize: '11px' }}>GET /orders/:id/events</code> endpoint needed.
        </p>
      </div>

      {/* AI Section */}
      <div style={{ margin: '28px 0 16px' }}>
        <h2 className="greek-heading" style={{ fontSize: '18px' }}>AI Assistant</h2>
        <div className="greek-divider" style={{ marginTop: '10px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      <ExceptionPanel analysis={aiAnalysis} isLoading={isAiLoading} onTriggerAnalysis={handleTriggerAiAnalysis} />
      {aiAnalysis && <DraftPanel orderId={order.id} />}

      {isModalOpen && (
        <OrderStatusModal
          currentStatus={order.status}
          onClose={() => setIsModalOpen(false)}
          onConfirm={handleUpdateStatus}
          allowedRoles={['admin']}
        />
      )}
    </div>
  );
};
// Force TS server reload
