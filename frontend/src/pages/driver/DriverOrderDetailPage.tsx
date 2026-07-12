import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { orderService, type Order, type OrderStatus } from '../../services/order.service';
import { eventService } from '../../services/event.service';
import { StatusBadge } from '../../components/orders/StatusBadge';
import { OrderStatusModal } from '../../components/orders/OrderStatusModal';
import { NoteModal } from '../../components/driver/NoteModal';
import { toast } from 'react-hot-toast';
import { injectGreekStyles } from '../../styles/greekTheme';

// ── Shared field styles ───────────────────────────────────────────────────────
const FL: React.CSSProperties = {
  fontFamily: "'Cinzel', serif",
  fontSize: 9,
  fontWeight: 600,
  letterSpacing: '0.22em',
  textTransform: 'uppercase',
  color: 'rgba(201,168,76,0.60)',
  marginBottom: 5,
};
const FV: React.CSSProperties = {
  fontFamily: "'Inter', sans-serif",
  fontSize: 14,
  color: 'rgba(232,213,163,0.85)',
  lineHeight: 1.5,
};

// ── Icons ─────────────────────────────────────────────────────────────────────
const IconPin = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path d="M8 1.5C5.5 1.5 3.5 3.5 3.5 6c0 3.5 4.5 8.5 4.5 8.5S12.5 9.5 12.5 6c0-2.5-2-4.5-4.5-4.5z" stroke="#C9A84C" strokeWidth="1.1" fill="rgba(201,168,76,0.08)" opacity="0.7" />
    <circle cx="8" cy="6" r="1.5" stroke="#C9A84C" strokeWidth="1.1" fill="none" opacity="0.7" />
  </svg>
);
const IconPhone = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path d="M3 2h3l1.5 3.5-2 1.5A9 9 0 0 0 9 11l1.5-2L14 10.5V14c-6 0-11-5-11-12z" stroke="#C9A84C" strokeWidth="1.1" strokeLinejoin="round" fill="none" opacity="0.6" />
  </svg>
);
const IconNote = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <rect x="2" y="2" width="12" height="12" rx="1.5" stroke="#C9A84C" strokeWidth="1.1" fill="none" opacity="0.6" />
    <path d="M5 6h6M5 9h4" stroke="#C9A84C" strokeWidth="1.1" strokeLinecap="round" opacity="0.6" />
  </svg>
);

// ── Page ─────────────────────────────────────────────────────────────────────
export const DriverOrderDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [order, setOrder] = useState<Order | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isStatusModalOpen, setIsStatusModalOpen] = useState(false);
  const [isNoteModalOpen, setIsNoteModalOpen] = useState(false);

  useEffect(() => { injectGreekStyles(); }, []);

  const fetchOrder = async () => {
    if (!id) return;
    setIsLoading(true);
    try {
      const data = await orderService.getOrder(parseInt(id, 10));
      setOrder(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch order details');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { fetchOrder(); }, [id]);

  const handleUpdateStatus = async (newStatus: OrderStatus) => {
    if (!id) return;
    await orderService.updateStatus(parseInt(id, 10), { status: newStatus });
    fetchOrder();
    setIsStatusModalOpen(false);
  };

  const handleAddNote = async (note: string) => {
    if (!order) return;
    await eventService.addDriverNote(order.id, note);
    setIsNoteModalOpen(false);
    toast.success('Note added successfully!');
    fetchOrder();
  };

  // ── Loading ──
  if (isLoading) {
    return (
      <div className="greek-loading-screen">
        <span className="greek-spinner" style={{ width: 24, height: 24, borderTopColor: '#C9A84C' }} />
        <span className="greek-loading-text" style={{ marginTop: 16 }}>Loading Order…</span>
      </div>
    );
  }

  // ── Error ──
  if (error || !order) {
    return (
      <div className="greek-bg">
        <div className="greek-page" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', gap: 20 }}>
          <svg width="40" height="40" viewBox="0 0 40 40" fill="none" style={{ opacity: 0.30 }}>
            <circle cx="20" cy="20" r="18" stroke="#C9A84C" strokeWidth="1.2" />
            <path d="M20 12v9M20 25v3" stroke="#C9A84C" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
          <p className="greek-muted" style={{ fontSize: 14 }}>{error || 'Order not found'}</p>
          <button className="greek-btn-secondary" onClick={() => navigate('/driver/orders')}>
            ← Back to Deliveries
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="greek-bg">
      <div className="greek-page" style={{ maxWidth: 640, margin: '0 auto' }}>

        {/* ── Header ── */}
        <div style={{ marginBottom: 32 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 6 }}>
            <button className="greek-btn-secondary" onClick={() => navigate('/driver/orders')}>
              ← Back
            </button>
            <h1 className="greek-heading" style={{ fontSize: 22, margin: 0 }}>
              Order #{order.id}
            </h1>
            <StatusBadge status={order.status} />
          </div>
          <div className="greek-divider" style={{ marginTop: 14 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── Delivery details card ── */}
        <div
          className="greek-card greek-card-corners"
          style={{ marginBottom: 20, animation: 'greek-fade-in 0.4s ease both' }}
        >
          {/* Card title */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 24, paddingBottom: 14, borderBottom: '1px solid rgba(201,168,76,0.10)' }}>
            <IconPin />
            <h3 className="greek-subheading" style={{ margin: 0 }}>Delivery Details</h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

            {/* Customer */}
            <div>
              <p style={FL}>Customer</p>
              <p style={{ ...FV, fontWeight: 500 }}>{order.username}</p>
            </div>

            {/* Phone — highlighted gold, tappable on mobile */}
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 5 }}>
                <IconPhone />
                <p style={{ ...FL, marginBottom: 0 }}>Phone</p>
              </div>
              <a
                href={`tel:${order.user_phone}`}
                style={{ fontFamily: "'Inter', sans-serif", fontSize: 16, fontWeight: 600, color: '#F0C040', letterSpacing: '0.03em', textDecoration: 'none' }}
              >
                {order.user_phone}
              </a>
            </div>

            {/* Address */}
            <div>
              <p style={FL}>Shipping Address</p>
              <p style={{ ...FV, lineHeight: 1.6 }}>{order.shipping_address}</p>
            </div>

            {/* Driver note (if exists) */}
            {order.driver_note && (
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
                  <IconNote />
                  <p style={{ ...FL, marginBottom: 0 }}>Your Note</p>
                </div>
                <div style={{ padding: '10px 14px', background: 'rgba(201,168,76,0.06)', borderLeft: '2px solid rgba(201,168,76,0.40)', borderRadius: '0 2px 2px 0' }}>
                  <p style={{ ...FV, fontStyle: 'italic', color: 'rgba(232,213,163,0.75)', margin: 0 }}>
                    {order.driver_note}
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* ── Action buttons ── */}
        <div
          style={{ display: 'flex', flexDirection: 'column', gap: 12, animation: 'greek-fade-in 0.4s ease 0.1s both' }}
        >
          <button
            className="greek-btn-primary"
            style={{ width: '100%', padding: '15px' }}
            onClick={() => setIsStatusModalOpen(true)}
          >
            Update Delivery Status
          </button>
          <button
            className="greek-btn-secondary"
            style={{ width: '100%', padding: '13px' }}
            onClick={() => setIsNoteModalOpen(true)}
          >
            {order.driver_note ? 'Edit Note / Report Issue' : 'Add Note / Report Issue'}
          </button>
        </div>

        {/* ── Modals ── */}
        {isStatusModalOpen && (
          <OrderStatusModal
            currentStatus={order.status}
            onClose={() => setIsStatusModalOpen(false)}
            onConfirm={handleUpdateStatus}
            allowedRoles={['driver']}
          />
        )}
        {isNoteModalOpen && (
          <NoteModal
            onClose={() => setIsNoteModalOpen(false)}
            onSave={handleAddNote}
          />
        )}

      </div>
    </div>
  );
};