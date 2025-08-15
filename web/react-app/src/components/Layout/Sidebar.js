import React from 'react';
import { Nav } from 'react-bootstrap';
import { LinkContainer } from 'react-router-bootstrap';
import { useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const Sidebar = () => {
  const location = useLocation();
  const { user } = useAuth();

  const menuItems = [
    {
      path: '/dashboard',
      icon: 'bi-speedometer2',
      label: 'Dashboard',
      adminOnly: false,
    },
    {
      path: '/nodes',
      icon: 'bi-hdd-stack',
      label: 'Nodes',
      adminOnly: false,
    },
    {
      path: '/environments',
      icon: 'bi-globe',
      label: 'Environments',
      adminOnly: false,
    },
    {
      path: '/groups',
      icon: 'bi-collection',
      label: 'Groups',
      adminOnly: false,
    },
    {
      path: '/process-enhanced',
      icon: 'bi-gear-wide-connected',
      label: 'Process Enhanced',
      adminOnly: false,
    },
    {
      path: '/log-analysis',
      icon: 'bi-graph-up',
      label: 'Log Analysis',
      adminOnly: false,
    },
    {
      path: '/system-monitoring',
      icon: 'bi-activity',
      label: 'System Monitoring',
      adminOnly: false,
    },
    {
      path: '/users',
      icon: 'bi-people',
      label: 'Users',
      adminOnly: true,
    },
    {
      path: '/permissions',
      icon: 'bi-shield-lock',
      label: 'Permissions',
      adminOnly: true,
    },
    {
      path: '/configuration',
      icon: 'bi-sliders',
      label: 'Configuration',
      adminOnly: true,
    },
    {
      path: '/data-management',
      icon: 'bi-database',
      label: 'Data Management',
      adminOnly: true,
    },
    {
      path: '/system-settings',
      icon: 'bi-gear',
      label: 'System Settings',
      adminOnly: true,
    },
    {
      path: '/developer-tools',
      icon: 'bi-code-slash',
      label: 'Developer Tools',
      adminOnly: true,
    },
    {
      path: '/activity-logs',
      icon: 'bi-journal-text',
      label: 'Activity Logs',
      adminOnly: false,
    },
  ];

  const filteredMenuItems = menuItems.filter(item => 
    !item.adminOnly || user?.is_admin
  );

  return (
    <div className="sidebar bg-light border-end" style={{ width: '250px', minHeight: '100vh' }}>
      <div className="p-3">
        <Nav className="flex-column">
          {filteredMenuItems.map((item) => (
            <LinkContainer key={item.path} to={item.path}>
              <Nav.Link 
                className={`d-flex align-items-center py-2 px-3 mb-1 rounded ${
                  location.pathname === item.path ? 'bg-primary text-white' : 'text-dark'
                }`}
              >
                <i className={`${item.icon} me-3`}></i>
                {item.label}
              </Nav.Link>
            </LinkContainer>
          ))}
        </Nav>
      </div>
    </div>
  );
};

export default Sidebar;