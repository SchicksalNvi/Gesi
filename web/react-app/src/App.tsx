import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from '@/store';
import { authApi } from '@/api/auth';
import MainLayout from '@/layouts/MainLayout';
import Login from '@/pages/Login';
import Dashboard from '@/pages/Dashboard';
import NodeList from '@/pages/Nodes';
import NodeDetail from '@/pages/Nodes/NodeDetail';
import ProcessesPage from '@/pages/Processes';
import EnvironmentList from '@/pages/Environments';
import EnvironmentDetail from '@/pages/Environments/EnvironmentDetail';
import UserList from '@/pages/Users';
import AlertList from '@/pages/Alerts';
import AlertRules from '@/pages/Alerts/AlertRules';
import LogList from '@/pages/Logs';
import Settings from '@/pages/Settings';

// Protected Route Component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
}

function App() {
  const { isAuthenticated, setUser, logout } = useStore();

  // Listen for logout events from API interceptor
  useEffect(() => {
    const handleLogout = () => {
      logout();
    };

    window.addEventListener('auth:logout', handleLogout);
    return () => window.removeEventListener('auth:logout', handleLogout);
  }, [logout]);

  // Validate token on mount only if we have both token and user
  useEffect(() => {
    const token = localStorage.getItem('token');
    const user = localStorage.getItem('user');
    
    if (token && user) {
      // We have stored credentials, verify they're still valid
      loadUserInfo();
    } else if (token || user) {
      // Partial credentials, clean up
      logout();
    }
  }, []);

  const loadUserInfo = async () => {
    try {
      const response = await authApi.getCurrentUser();
      // 后端返回 { status, data: { user: {...} } }
      if (response.status === 'success' && (response as any).data?.user) {
        setUser((response as any).data.user);
      } else {
        // Invalid response, logout
        logout();
      }
    } catch (error) {
      console.error('Failed to load user info:', error);
      // Token is invalid, logout
      logout();
    }
  };

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <MainLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="nodes" element={<NodeList />} />
        <Route path="nodes/:nodeName" element={<NodeDetail />} />
        <Route path="processes" element={<ProcessesPage />} />
        <Route path="environments" element={<EnvironmentList />} />
        <Route path="environments/:environmentName" element={<EnvironmentDetail />} />
        <Route path="users" element={<UserList />} />
        <Route path="alerts" element={<AlertList />} />
        <Route path="alerts/rules" element={<AlertRules />} />
        <Route path="logs" element={<LogList />} />
        <Route path="settings" element={<Settings />} />
      </Route>
      
      {/* 未匹配的路由：已登录去 dashboard，未登录去 login */}
      <Route 
        path="*" 
        element={
          isAuthenticated ? (
            <Navigate to="/dashboard" replace />
          ) : (
            <Navigate to="/login" replace />
          )
        } 
      />
    </Routes>
  );
}

export default App;
