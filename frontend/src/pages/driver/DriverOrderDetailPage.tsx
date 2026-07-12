import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { orderService, type Order, type OrderStatus } from '../../services/order.service';
import { eventService } from '../../services/event.service';
import { StatusBadge } from '../../components/orders/StatusBadge';
import { OrderStatusModal } from '../../components/orders/OrderStatusModal';
import { NoteModal } from '../../components/driver/NoteModal';
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
    fetchOrder(); // refresh to show the saved note
  };

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
        <button className="greek-btn-secondary" style={{ marginTop: '16px' }} onClick={() => navigate('/driver/orders')}>
          Back to Deliveries
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center gap-4" style={{ marginBottom: '28px' }}>
        <button className="greek-btn-secondary" onClick={() => navigate('/driver/orders')}>← Back</button>
        <div>
          <h1 className="greek-heading" style={{ fontSize: '22px' }}>Order #{order.id}</h1>
        </div>
      </div>

      {/* Order card */}
      <div className="greek-card-static mb-4">
        <div className="flex justify-between items-start" style={{ marginBottom: '20px' }}>
          <h3 className="greek-subheading">Delivery Details</h3>
          <StatusBadge status={order.status} />
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <div><p style={FIELD_LABEL}>Customer</p><p style={FIELD_VALUE}>{order.username}</p></div>
          <div><p style={FIELD_LABEL}>Phone</p><p style={{ ...FIELD_VALUE, color: '#C9A84C', fontWeight: 600 }}>{order.user_phone}</p></div>
          <div><p style={FIELD_LABEL}>Shipping Address</p><p style={FIELD_VALUE}>{order.shipping_address}</p></div>
          {order.driver_note && (
            <div style={{
              padding: '12px 16px',
              background: 'rgba(201,168,76,0.07)',
              borderLeft: '3px solid rgba(201,168,76,0.45)',
              borderRadius: '4px',
            }}>
              <p style={FIELD_LABEL}>Your Note</p>
              <p style={{ ...FIELD_VALUE, fontStyle: 'italic' }}>{order.driver_note}</p>
            </div>
          )}
        </div>
      </div>

      {/* Action buttons */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        <button
          className="greek-btn-primary"
          style={{ padding: '16px', width: '100%' }}
          onClick={() => setIsStatusModalOpen(true)}
        >
          Update Status
        </button>
        <button
          className="greek-btn-secondary"
          style={{ padding: '14px', width: '100%' }}
          onClick={() => setIsNoteModalOpen(true)}
        >
          Add Note / Report Issue
        </button>
      </div>

      {isStatusModalOpen && (
        <OrderStatusModal
          currentStatus={order.status}
          onClose={() => setIsStatusModalOpen(false)}
          onConfirm={handleUpdateStatus}
          allowedRoles={['driver']}
        />
      )}
      {isNoteModalOpen && (
        <NoteModal onClose={() => setIsNoteModalOpen(false)} onSave={handleAddNote} />
      )}
    </div>
  );
};
