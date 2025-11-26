import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useStore } from '@/store';
import { authApi } from '@/api/auth';
import MainLayout from '@/layouts/MainLayout';
import Login from '@/pages/Login';
import Dashboard from '@/pages/Dashboard';
import NodeList from '@/pages/Nodes';
import NodeDetail from '@/pages/Nodes/NodeDetail';
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
  const { isAuthenticated, setUser } = useStore();

  // Load user info on mount
  useEffect(() => {
    if (isAuthenticated) {
      loadUserInfo();
    }
  }, [isAuthenticated]);

  const loadUserInfo = async () => {
    try {
      const response = await authApi.getCurrentUser();
      if (response.data) {
        setUser(response.data);
      }
    } catch (error) {
      console.error('Failed to load user info:', error);
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
