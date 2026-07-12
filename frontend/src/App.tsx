import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ProtectedRoute } from './components/layout/ProtectedRoute';
import { DashboardLayout } from './components/layout/DashboardLayout';
import { LoginPage } from './pages/LoginPage';
import { OrdersPage } from './pages/admin/OrdersPage';
import { OrderDetailPage } from './pages/admin/OrderDetailPage';
import { CreateOrderPage } from './pages/admin/CreateOrderPage';
import { DashboardPage } from './pages/admin/DashboardPage';
import { DailyReportPage } from './pages/admin/DailyReportPage';
import { DriverOrdersPage } from './pages/driver/DriverOrdersPage';
import { DriverOrderDetailPage } from './pages/driver/DriverOrderDetailPage';
import { Toaster } from 'react-hot-toast';
// Note: We'll create these actual pages in Phase 2 & 3.
// For now, we use placeholders to test the routing.

function App() {
  return (
    <AuthProvider>
      <Toaster
        position="top-right"
        toastOptions={{
          style: {
            background: '#13131C',
            color: '#E8D5A3',
            border: '1px solid rgba(201,168,76,0.25)',
            fontFamily: "'Inter', sans-serif",
            fontSize: '13px',
          },
          success: {
            iconTheme: { primary: '#34d399', secondary: '#13131C' },
          },
          error: {
            iconTheme: { primary: '#f87171', secondary: '#13131C' },
          },
        }}
      />
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          
          {/* Admin Routes */}
          <Route path="/admin" element={<ProtectedRoute allowedRoles={['admin']} />}>
            <Route element={<DashboardLayout />}>
              <Route path="dashboard" element={<DashboardPage />} />
              <Route path="orders" element={<OrdersPage />} />
              <Route path="orders/new" element={<CreateOrderPage />} />
              <Route path="orders/:id" element={<OrderDetailPage />} />
              <Route path="reports" element={<DailyReportPage />} />
              <Route index element={<Navigate to="/admin/dashboard" replace />} />
            </Route>
          </Route>

          {/* Driver Routes */}
          <Route path="/driver" element={<ProtectedRoute allowedRoles={['driver']} />}>
            <Route element={<DashboardLayout />}>
              <Route path="orders" element={<DriverOrdersPage />} />
              <Route path="orders/:id" element={<DriverOrderDetailPage />} />
              <Route index element={<Navigate to="/driver/orders" replace />} />
            </Route>
          </Route>

          {/* Fallback */}
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
