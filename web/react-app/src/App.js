import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Container } from 'react-bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap-icons/font/bootstrap-icons.css';
import './App.css';

// Components
import Navbar from './components/Layout/Navbar';
import Sidebar from './components/Layout/Sidebar';
import ProtectedRoute from './components/Auth/ProtectedRoute';

// Pages
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import NodeDetail from './pages/NodeDetail';
import Environments from './pages/Environments';
import Groups from './pages/Groups';
import Users from './pages/Users';
import ActivityLogs from './pages/ActivityLogs';
import Profile from './pages/Profile';
import PermissionManagement from './pages/PermissionManagement';
import ProcessManagementEnhanced from './pages/ProcessManagementEnhanced';
import LogAnalysis from './pages/LogAnalysis';
import SystemMonitoring from './pages/SystemMonitoring';
import ConfigurationManagement from './pages/ConfigurationManagement';
import DataManagement from './pages/DataManagement';
import SystemSettings from './pages/SystemSettings';
import DeveloperTools from './pages/DeveloperTools';

// Context
import { AuthProvider } from './contexts/AuthContext';
import { WebSocketProvider } from './contexts/WebSocketContext';

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