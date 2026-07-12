import React, { useEffect, useState } from 'react';
import { orderService } from '../../services/order.service';
import { StatusBadge } from '../../components/orders/StatusBadge';
import { useNavigate } from 'react-router-dom';
import { injectGreekStyles } from '../../styles/greekTheme';

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
          created:   res.data.filter(o => o.status === 'created').length,
          delivered: res.data.filter(o => o.status === 'delivered').length,
          cancelled: res.data.filter(o => o.status === 'cancelled').length,
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
        <span className="greek-loading-text">Loading Dashboard…</span>
      </div>
    );
  }

  return (
    <div>
      {/* Page title */}
      <div style={{ marginBottom: '32px' }}>
        <h1 className="greek-heading" style={{ fontSize: '24px' }}>Overview Dashboard</h1>
        <div className="greek-divider" style={{ marginTop: '12px' }}>
          <div className="greek-divider-line" />
          <div className="greek-divider-diamond" />
          <div className="greek-divider-line" />
        </div>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '28px' }}>
        {[
          { label: 'Total Orders', value: stats.totalOrders, cls: 'greek-stat-value' },
          { label: 'Recent Created', value: stats.created, cls: 'greek-stat-value-neutral' },
          { label: 'Recent Delivered', value: stats.delivered, cls: 'greek-stat-value-success' },
          { label: 'Recent Cancelled', value: stats.cancelled, cls: 'greek-stat-value-danger' },
        ].map(stat => (
          <div key={stat.label} className="greek-card-static" style={{ textAlign: 'center' }}>
            <p className="greek-muted" style={{ marginBottom: '8px' }}>{stat.label}</p>
            <p className={stat.cls}>{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Recent Orders */}
      <div className="greek-card-static" style={{ padding: 0, overflow: 'hidden' }}>
        <div style={{
          padding: '14px 20px',
          borderBottom: '1px solid rgba(201,168,76,0.15)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: 'linear-gradient(90deg, rgba(201,168,76,0.04) 0%, transparent 100%)',
        }}>
          <h3 className="greek-subheading">Recent Orders</h3>
          <button className="greek-btn-secondary" style={{ padding: '6px 16px' }} onClick={() => navigate('/admin/orders')}>
            View All
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
            {recentOrders.length > 0 ? recentOrders.map((order, i) => (
              <tr
                key={order.id}
                onClick={() => navigate(`/admin/orders/${order.id}`)}
                style={{ cursor: 'pointer', animationDelay: `${i * 50}ms` }}
              >
                <td><span style={{ color: '#C9A84C', fontWeight: 600 }}>#{order.id}</span></td>
                <td>{order.username}</td>
                <td style={{ fontWeight: 600 }}>{formatCurrency(order.total_amount)}</td>
                <td><StatusBadge status={order.status} /></td>
              </tr>
            )) : (
              <tr>
                <td colSpan={4} style={{ textAlign: 'center', padding: '32px', color: 'rgba(201,168,76,0.35)', fontStyle: 'italic' }}>
                  No recent orders
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
