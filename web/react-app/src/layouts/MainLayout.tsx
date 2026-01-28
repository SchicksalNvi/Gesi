import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Space, Badge } from 'antd';
import {
  DashboardOutlined,
  ClusterOutlined,
  AppstoreOutlined,
  UserOutlined,
  BellOutlined,
  FileTextOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RadarChartOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { useStore } from '@/store';
import { GesiLogo } from '@/components/GesiLogo';

const { Header, Sider, Content } = Layout;

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useStore();
  const [collapsed, setCollapsed] = useState(false);

  // Menu items
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
      onClick: () => navigate('/dashboard'),
    },
    {
      key: '/environments',
      icon: <AppstoreOutlined />,
      label: 'Environments',
      onClick: () => navigate('/environments'),
    },
    {
      key: '/nodes',
      icon: <ClusterOutlined />,
      label: 'Nodes',
      onClick: () => navigate('/nodes'),
    },
    {
      key: '/discovery',
      icon: <RadarChartOutlined />,
      label: 'Discovery',
      onClick: () => navigate('/discovery'),
    },
    {
      key: '/processes',
      icon: <AppstoreOutlined />,
      label: 'Processes',
      onClick: () => navigate('/processes'),
    },
    {
      key: 'alerts-menu',
      icon: <BellOutlined />,
      label: 'Alerts',
      children: [
        {
          key: '/alerts',
          label: 'Alert List',
          onClick: () => navigate('/alerts'),
        },
        {
          key: '/alerts/rules',
          label: 'Alert Rules',
          onClick: () => navigate('/alerts/rules'),
        },
      ],
    },
    {
      key: '/logs',
      icon: <FileTextOutlined />,
      label: 'Logs',
      onClick: () => navigate('/logs'),
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: 'Users',
      onClick: () => navigate('/users'),
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: 'Settings',
      onClick: () => navigate('/settings'),
    },
  ];

  // User dropdown menu
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: 'Settings',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: () => {
        logout();
        navigate('/login');
      },
    },
  ];

  // Get current selected key from location
  const selectedKey = location.pathname;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            padding: '0 20px',
          }}
          onClick={() => navigate('/dashboard')}
        >
          <GesiLogo size={36} collapsed={collapsed} textColor="#fff" />
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
        />
      </Sider>
      
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'all 0.2s' }}>
        <Header
          style={{
            padding: '0 24px',
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 4px rgba(0,21,41,.08)',
          }}
        >
          <div>
            {collapsed ? (
              <MenuUnfoldOutlined
                style={{ fontSize: 18, cursor: 'pointer' }}
                onClick={() => setCollapsed(false)}
              />
            ) : (
              <MenuFoldOutlined
                style={{ fontSize: 18, cursor: 'pointer' }}
                onClick={() => setCollapsed(true)}
              />
            )}
          </div>
          
          <Space size="large">
            <Badge count={0} showZero={false}>
              <BellOutlined style={{ fontSize: 18, cursor: 'pointer' }} />
            </Badge>
            
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Space style={{ cursor: 'pointer' }}>
                <Avatar icon={<UserOutlined />} />
                <span>{user?.username || 'User'}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        
        <Content
          style={{
            margin: '24px 16px',
            padding: 24,
            minHeight: 280,
            background: '#f0f2f5',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
