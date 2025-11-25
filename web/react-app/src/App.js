import React, { Suspense, lazy } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Container, Spinner } from 'react-bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap-icons/font/bootstrap-icons.css';
import './App.css';

// Components (不懒加载，因为它们在所有页面都需要)
import Navbar from './components/Layout/Navbar';
import Sidebar from './components/Layout/Sidebar';
import ProtectedRoute from './components/Auth/ProtectedRoute';

// Context (不懒加载，因为它们是全局的)
import { AuthProvider } from './contexts/AuthContext';
import { WebSocketProvider } from './contexts/WebSocketContext';

// 登录页面不懒加载，因为它是首次访问的页面
import Login from './pages/Login';

// 使用 React.lazy() 懒加载页面组件
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Nodes = lazy(() => import('./pages/Nodes'));
const NodeDetail = lazy(() => import('./pages/NodeDetail'));
const Environments = lazy(() => import('./pages/Environments'));
const Groups = lazy(() => import('./pages/Groups'));
const Users = lazy(() => import('./pages/Users'));
const ActivityLogs = lazy(() => import('./pages/ActivityLogs'));
const Profile = lazy(() => import('./pages/Profile'));
const PermissionManagement = lazy(() => import('./pages/PermissionManagement'));
const ProcessManagementEnhanced = lazy(() => import('./pages/ProcessManagementEnhanced'));
const LogAnalysis = lazy(() => import('./pages/LogAnalysis'));
const SystemMonitoring = lazy(() => import('./pages/SystemMonitoring'));
const ConfigurationManagement = lazy(() => import('./pages/ConfigurationManagement'));
const DataManagement = lazy(() => import('./pages/DataManagement'));
const SystemSettings = lazy(() => import('./pages/SystemSettings'));
const DeveloperTools = lazy(() => import('./pages/DeveloperTools'));

// 加载指示器组件
const LoadingFallback = () => (
  <div className="d-flex justify-content-center align-items-center" style={{ minHeight: '400px' }}>
    <Spinner animation="border" role="status" variant="primary">
      <span className="visually-hidden">Loading...</span>
    </Spinner>
  </div>
);

function App() {
  return (
    <AuthProvider>
      <WebSocketProvider>
        <Router>
          <div className="App">
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/*" element={
                <ProtectedRoute>
                  <div className="d-flex">
                    <Sidebar />
                    <div className="flex-grow-1">
                      <Navbar />
                      <Container fluid className="main-content p-4">
                        <Suspense fallback={<LoadingFallback />}>
                          <Routes>
                            <Route path="/" element={<Navigate to="/dashboard" replace />} />
                            <Route path="/dashboard" element={<Dashboard />} />
                            <Route path="/nodes" element={<Nodes />} />
                            <Route path="/nodes/:nodeName" element={<NodeDetail />} />
                            <Route path="/environments" element={<Environments />} />
                            <Route path="/groups" element={<Groups />} />
                            <Route path="/users" element={<Users />} />
                            <Route path="/activity-logs" element={<ActivityLogs />} />
                            <Route path="/profile" element={<Profile />} />
                            <Route path="/permissions" element={<PermissionManagement />} />
                            <Route path="/process-enhanced" element={<ProcessManagementEnhanced />} />
                            <Route path="/log-analysis" element={<LogAnalysis />} />
                            <Route path="/system-monitoring" element={<SystemMonitoring />} />
                            <Route path="/configuration" element={<ConfigurationManagement />} />
                            <Route path="/data-management" element={<DataManagement />} />
                            <Route path="/system-settings" element={<SystemSettings />} />
                            <Route path="/developer-tools" element={<DeveloperTools />} />
                          </Routes>
                        </Suspense>
                      </Container>
                    </div>
                  </div>
                </ProtectedRoute>
              } />
            </Routes>
          </div>
        </Router>
      </WebSocketProvider>
    </AuthProvider>
  );
}

export default App;