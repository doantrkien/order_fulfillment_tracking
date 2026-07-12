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
import { useTheme } from '../../hooks/useTheme';

// ── Shared field styles ───────────────────────────────────────────────────────
const FL: React.CSSProperties = {
  fontFamily: "'Cinzel', serif",
  fontSize: 9,
  fontWeight: 600,
  letterSpacing: '0.22em',
  textTransform: 'uppercase',
  marginBottom: 5,
};
const FV: React.CSSProperties = {
  fontFamily: "'Inter', sans-serif",
  fontSize: 14,
  lineHeight: 1.5,
};

// ── Field block ───────────────────────────────────────────────────────────────
const Field: React.FC<{ label: string; children: React.ReactNode; wide?: boolean }> = ({ label, children, wide }) => (
  <div style={{ gridColumn: wide ? '1 / -1' : undefined }}>
    <p className="greek-label" style={FL}>{label}</p>
    <div className="greek-text" style={FV}>{children}</div>
  </div>
);

// ── Section card ─────────────────────────────────────────────────────────────
const Section: React.FC<{ title: string; icon?: React.ReactNode; children: React.ReactNode; delay?: number }> = ({ title, icon, children, delay = 0 }) => (
  <div
    className="greek-card-static"
    style={{ marginBottom: 16, animation: 'greek-fade-in 0.4s ease both', animationDelay: `${delay}ms` }}
  >
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 18, paddingBottom: 14, borderBottom: '1px solid rgba(201,168,76,0.10)' }}>
      {icon}
      <h3 className="greek-subheading" style={{ margin: 0 }}>{title}</h3>
    </div>
    {children}
  </div>
);

// ── Icons ─────────────────────────────────────────────────────────────────────
const IconPerson: React.FC<{ isLight?: boolean }> = ({ isLight }) => {
  const color = isLight ? "#A07814" : "#C9A84C";
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
      <circle cx="8" cy="5" r="3" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
      <path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6" stroke={color} strokeWidth="1.1" strokeLinecap="round" fill="none" opacity="0.6" />
    </svg>
  );
};
const IconBox: React.FC<{ isLight?: boolean }> = ({ isLight }) => {
  const color = isLight ? "#A07814" : "#C9A84C";
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
      <path d="M2 5l6-3 6 3v6l-6 3-6-3V5z" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
      <path d="M8 2v12M2 5l6 3 6-3" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
    </svg>
  );
};
const IconHistory: React.FC<{ isLight?: boolean }> = ({ isLight }) => {
  const color = isLight ? "#A07814" : "#C9A84C";
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
      <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.1" fill="none" opacity="0.6" />
      <path d="M8 5v3.5l2 2" stroke={color} strokeWidth="1.1" strokeLinecap="round" opacity="0.6" />
    </svg>
  );
};
const IconAI: React.FC<{ isLight?: boolean }> = ({ isLight }) => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path d="M8 1l1.5 4.5L14 7l-4.5 1.5L8 13l-1.5-4.5L2 7l4.5-1.5z" stroke={isLight ? "#C9920A" : "#F0C040"} strokeWidth="1.1" fill={isLight ? "rgba(201,146,10,0.10)" : "rgba(240,192,64,0.10)"} opacity="0.9" />
  </svg>
);

// ── Page ─────────────────────────────────────────────────────────────────────
export const OrderDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isLight } = useTheme();
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

  // ── Loading ──
  if (isLoading) {
    return (
      <div className="greek-loading-screen">
        <span className="greek-spinner" style={{ width: 24, height: 24 }} />
        <span className="greek-loading-text" style={{ marginTop: 16 }}>Loading Order…</span>
      </div>
    );
  }

  // ── Error ──
  if (error || !order) {
    return (
      <div className="greek-bg">
        <div className="greek-page" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', gap: 20 }}>
          <svg width="40" height="40" viewBox="0 0 40 40" fill="none" style={{ opacity: 0.35 }}>
            <circle cx="20" cy="20" r="18" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.2" />
            <path d="M20 12v9M20 25v3" stroke={isLight ? "#A07814" : "#C9A84C"} strokeWidth="1.5" strokeLinecap="round" />
          </svg>
          <p className="greek-muted" style={{ fontSize: 14 }}>{error || 'Order not found'}</p>
          <button className="greek-btn-secondary" onClick={() => navigate('/admin/orders')}>
            ← Back to Orders
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 32 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 16, flexWrap: 'wrap' }}>

            <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
              <button className="greek-btn-secondary" onClick={() => navigate('/admin/orders')}>
                ← Back
              </button>
              <div>
                <h1 className="greek-heading" style={{ fontSize: 22, margin: 0 }}>
                  Order #{order.id}
                </h1>
                <p className="greek-muted" style={{ marginTop: 4 }}>
                  Placed {formatDate(order.ordered_at)}
                </p>
              </div>
              <StatusBadge status={order.status} />
            </div>

            <button className="greek-btn-primary" onClick={() => setIsModalOpen(true)}>
              Update Status
            </button>
          </div>

          <div className="greek-divider" style={{ marginTop: 16 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── Customer info ── */}
        <Section title="Customer Information" icon={<IconPerson isLight={isLight} />} delay={0}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <Field label="Name">{order.username}</Field>
            <Field label="Phone">{order.user_phone}</Field>
            <Field label="Shipping Address" wide>{order.shipping_address}</Field>
          </div>
        </Section>

        {/* ── Order summary ── */}
        <Section title="Order Summary" icon={<IconBox isLight={isLight} />} delay={80}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>

            <div>
              <p className="greek-label" style={FL}>Total Amount</p>
              <div style={{ fontFamily: "'Inter', sans-serif", fontSize: 24, fontWeight: 600, color: isLight ? '#C9920A' : '#F0C040', letterSpacing: '-0.01em', lineHeight: 1 }}>
                {formatCurrency(order.total_amount)}
              </div>
            </div>

            <div>
              <p className="greek-label" style={FL}>Status</p>
              <div style={{ marginTop: 2 }}>
                <StatusBadge status={order.status} />
              </div>
            </div>

            {order.driver_note && (
              <div style={{ gridColumn: '1 / -1' }}>
                <p className="greek-label" style={FL}>Driver Note</p>
                <div style={{ padding: '10px 14px', background: isLight ? 'rgba(160,120,20,0.06)' : 'rgba(201,168,76,0.06)', borderLeft: `2px solid ${isLight ? 'rgba(160,120,20,0.40)' : 'rgba(201,168,76,0.40)'}`, borderRadius: '0 2px 2px 0' }}>
                  <p className="greek-text" style={{ ...FV, fontStyle: 'italic' }}>{order.driver_note}</p>
                </div>
              </div>
            )}
          </div>
        </Section>

        {/* ── Event history ── */}
        <Section title="Event History" icon={<IconHistory isLight={isLight} />} delay={160}>
          <div style={{ padding: '20px 0', textAlign: 'center' }}>
            <p className="greek-muted" style={{ fontStyle: 'italic', marginBottom: 6 }}>
              Event history is not yet available.
            </p>
            <code style={{ fontFamily: "'Inter', monospace", fontSize: 11, color: isLight ? 'rgba(160,120,20,0.60)' : 'rgba(201,168,76,0.40)', background: isLight ? 'rgba(160,120,20,0.06)' : 'rgba(201,168,76,0.06)', padding: '2px 8px', borderRadius: 2 }}>
              GET /orders/:id/events
            </code>
          </div>
        </Section>

        {/* ── AI Assistant ── */}
        <div style={{ marginTop: 36, marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 10 }}>
            <IconAI isLight={isLight} />
            <h2 className="greek-heading" style={{ fontSize: 18, margin: 0 }}>AI Assistant</h2>
          </div>
          <div className="greek-divider" style={{ marginTop: 10 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        <ExceptionPanel
          analysis={aiAnalysis}
          isLoading={isAiLoading}
          onTriggerAnalysis={handleTriggerAiAnalysis}
        />
        {aiAnalysis && <DraftPanel orderId={order.id} />}

        {/* ── Status modal ── */}
        {isModalOpen && (
          <OrderStatusModal
            currentStatus={order.status}
            onClose={() => setIsModalOpen(false)}
            onConfirm={handleUpdateStatus}
            allowedRoles={['admin']}
          />
        )}

      </div>
    </div>
  );
};