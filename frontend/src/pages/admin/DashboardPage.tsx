import React, { useEffect, useState } from 'react';
import { orderService } from '../../services/order.service';
import { StatusBadge } from '../../components/orders/StatusBadge';
import { useNavigate } from 'react-router-dom';
import { injectGreekStyles } from '../../styles/greekTheme';



const STAT_CONFIGS = [
  { key: 'totalOrders', label: 'Total Orders', color: '#F0C040', border: 'rgba(201,168,76,0.25)', bg: 'rgba(201,168,76,0.05)' },
  { key: 'created', label: 'Recent Created', color: '#80B4E8', border: 'rgba(128,180,232,0.25)', bg: 'rgba(128,180,232,0.05)' },
  { key: 'delivered', label: 'Recent Delivered', color: '#7EC88A', border: 'rgba(126,200,138,0.25)', bg: 'rgba(126,200,138,0.05)' },
  { key: 'cancelled', label: 'Recent Cancelled', color: '#E88080', border: 'rgba(232,128,128,0.25)', bg: 'rgba(232,128,128,0.05)' },
] as const;

export const DashboardPage: React.FC = () => {
  const [stats, setStats] = useState({ totalOrders: 0, created: 0, delivered: 0, cancelled: 0 });
  const [recentOrders, setRecentOrders] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => { injectGreekStyles(); }, []);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const res = await orderService.getOrders({ limit: 10 });
        setRecentOrders(res.data.slice(0, 5));
        setStats({
          totalOrders: res.total_items || 0,
          created: res.data.filter((o: any) => o.status === 'created').length,
          delivered: res.data.filter((o: any) => o.status === 'delivered').length,
          cancelled: res.data.filter((o: any) => o.status === 'cancelled').length,
        });
      } catch (error) {
        console.error('Failed to load dashboard data', error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchDashboardData();
  }, []);

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount);

  if (isLoading) {
    return (
      <div className="greek-loading-screen">
        <span className="greek-spinner" style={{ width: 24, height: 24, borderTopColor: '#C9A84C' }} />
        <span className="greek-loading-text" style={{ marginTop: 16 }}>Loading Dashboard…</span>
      </div>
    );
  }

  return (
    <div className="greek-bg">
      <div className="greek-page">

        {/* ── Page header ── */}
        <div style={{ marginBottom: 36 }}>
          <h1 className="greek-heading" style={{ fontSize: 24, margin: 0 }}>Overview Dashboard</h1>
          <p className="greek-muted" style={{ marginTop: 6 }}>
            Snapshot of the last 10 orders
          </p>
          <div className="greek-divider" style={{ marginTop: 14 }}>
            <div className="greek-divider-line" />
            <div className="greek-divider-diamond" />
            <div className="greek-divider-line" />
          </div>
        </div>

        {/* ── Stat cards ── */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16, marginBottom: 28 }}>
          {STAT_CONFIGS.map((s, i) => (
            <div
              key={s.key}
              style={{
                position: 'relative',
                background: 'linear-gradient(160deg, #13131C 0%, #0D0D15 100%)',
                border: `1px solid ${s.border}`,
                borderRadius: 4,
                padding: '22px 20px 18px',
                animation: 'greek-fade-in 0.4s ease both',
                animationDelay: `${i * 60}ms`,
                transition: 'border-color 0.25s, box-shadow 0.25s',
                cursor: 'default',
              }}
              onMouseEnter={e => {
                (e.currentTarget as HTMLDivElement).style.borderColor = s.color + '55';
                (e.currentTarget as HTMLDivElement).style.boxShadow = `0 4px 20px ${s.bg}`;
              }}
              onMouseLeave={e => {
                (e.currentTarget as HTMLDivElement).style.borderColor = s.border;
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'none';
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 14 }}>
                <span style={{ fontFamily: "'Cinzel', serif", fontSize: 9, fontWeight: 600, letterSpacing: '0.22em', textTransform: 'uppercase', color: 'rgba(201,168,76,0.60)' }}>
                  {s.label}
                </span>
              </div>
              <div style={{ fontFamily: "'Cinzel', serif", fontSize: 32, fontWeight: 600, color: s.color, lineHeight: 1 }}>
                {stats[s.key]}
              </div>
              <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 2, background: `linear-gradient(90deg, transparent, ${s.color}33, transparent)`, borderRadius: '0 0 4px 4px' }} />
            </div>
          ))}
        </div>

        {/* ── Recent Orders table ── */}
        <div style={{ position: 'relative', background: 'linear-gradient(160deg, #13131C 0%, #0D0D15 60%, #111118 100%)', border: '1px solid rgba(201,168,76,0.22)', borderRadius: 4, overflow: 'hidden', animation: 'greek-fade-in 0.5s ease 0.25s both' }}>

          {/* Table header bar */}
          <div style={{ padding: '16px 24px', borderBottom: '1px solid rgba(201,168,76,0.15)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'linear-gradient(90deg, rgba(201,168,76,0.05) 0%, transparent 100%)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M8 1L15 8L8 15L1 8Z" stroke="#C9A84C" strokeWidth="1.1" fill="rgba(201,168,76,0.08)" />
              </svg>
              <h3 className="greek-subheading" style={{ margin: 0 }}>Recent Orders</h3>
            </div>
            <button
              className="greek-btn-secondary"
              style={{ padding: '6px 16px', fontSize: 10 }}
              onClick={() => navigate('/admin/orders')}
            >
              View All →
            </button>
          </div>

          <table className="greek-table">
            <thead>
              <tr>
                <th>Order ID</th>
                <th>Customer</th>
                <th>Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {recentOrders.length > 0 ? (
                recentOrders.map((order, i) => (
                  <tr
                    key={order.id}
                    onClick={() => navigate(`/admin/orders/${order.id}`)}
                    style={{ cursor: 'pointer', animationDelay: `${i * 50}ms` }}
                  >
                    <td>
                      <span style={{ fontFamily: "'Cinzel', serif", fontSize: 12, fontWeight: 600, color: '#C9A84C', letterSpacing: '0.05em' }}>
                        #{order.id}
                      </span>
                    </td>
                    <td style={{ color: 'rgba(232,213,163,0.85)' }}>{order.username}</td>
                    <td style={{ fontFamily: "'Inter', sans-serif", fontWeight: 600, color: '#F0C040' }}>
                      {formatCurrency(order.total_amount)}
                    </td>
                    <td><StatusBadge status={order.status} /></td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={4} style={{ textAlign: 'center', padding: '40px', color: 'rgba(232,213,163,0.45)', fontStyle: 'italic', fontFamily: "'Inter', sans-serif", fontSize: 13 }}>
                    No recent orders found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

      </div>
    </div>
  );
};