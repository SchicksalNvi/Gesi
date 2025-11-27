import { useState } from 'react';
import {
  Card,
  Tabs,
  Form,
  Input,
  Button,
  Switch,
  Select,
  InputNumber,
  message,
  Space,
  Divider,
  Alert,
} from 'antd';
import {
  SaveOutlined,
  UserOutlined,
  LockOutlined,
  MailOutlined,
  BellOutlined,
  SettingOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import { useStore } from '@/store';

const Settings: React.FC = () => {
  const { user } = useStore();
  const [loading, setLoading] = useState(false);
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [notificationForm] = Form.useForm();
  const [systemForm] = Form.useForm();

  const handleProfileUpdate = async () => {
    try {
      const values = await profileForm.validateFields();
      setLoading(true);
      
      // 实际应该调用 API
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success('Profile updated successfully');
    } catch (error) {
      console.error('Failed to update profile:', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async () => {
    try {
      const values = await passwordForm.validateFields();
      setLoading(true);
      
      // 实际应该调用 API
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success('Password changed successfully');
      passwordForm.resetFields();
    } catch (error) {
      console.error('Failed to change password:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNotificationUpdate = async () => {
    try {
      const values = await notificationForm.validateFields();
      setLoading(true);
      
      // 实际应该调用 API
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success('Notification settings updated');
    } catch (error) {
      console.error('Failed to update notification settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSystemUpdate = async () => {
    try {
      const values = await systemForm.validateFields();
      setLoading(true);
      
      // 实际应该调用 API
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      message.success('System settings updated');
    } catch (error) {
      console.error('Failed to update system settings:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1 style={{ marginBottom: 16 }}>Settings</h1>

      <Card>
        <Tabs
          items={[
            {
              key: 'profile',
              label: (
                <span>
                  <UserOutlined />
                  Profile
                </span>
              ),
              children: (
                <div style={{ maxWidth: 600 }}>
                  <Alert
                    message="Profile Information"
                    description="Update your personal information and preferences"
                    type="info"
                    showIcon
                    style={{ marginBottom: 24 }}
                  />
                  
                  <Form
                    form={profileForm}
                    layout="vertical"
                    initialValues={{
                      username: user?.username || '',
                      email: user?.email || '',
                      full_name: '',
                      timezone: 'UTC',
                    }}
                  >
                    <Form.Item
                      name="username"
                      label="Username"
                    >
                      <Input
                        prefix={<UserOutlined />}
                        disabled
                      />
                    </Form.Item>

                    <Form.Item
                      name="email"
                      label="Email"
                      rules={[
                        { required: true, message: 'Please enter email' },
                        { type: 'email', message: 'Please enter valid email' },
                      ]}
                    >
                      <Input
                        prefix={<MailOutlined />}
                        placeholder="Enter email"
                      />
                    </Form.Item>

                    <Form.Item
                      name="full_name"
                      label="Full Name"
                    >
                      <Input placeholder="Enter full name" />
                    </Form.Item>

                    <Form.Item
                      name="timezone"
                      label="Timezone"
                    >
                      <Select
                        options={[
                          { label: 'UTC', value: 'UTC' },
                          { label: 'America/New_York', value: 'America/New_York' },
                          { label: 'America/Los_Angeles', value: 'America/Los_Angeles' },
                          { label: 'Europe/London', value: 'Europe/London' },
                          { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
                          { label: 'Asia/Tokyo', value: 'Asia/Tokyo' },
                        ]}
                      />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        icon={<SaveOutlined />}
                        onClick={handleProfileUpdate}
                        loading={loading}
                      >
                        Save Changes
                      </Button>
                    </Form.Item>
                  </Form>
                </div>
              ),
            },
            {
              key: 'security',
              label: (
                <span>
                  <LockOutlined />
                  Security
                </span>
              ),
              children: (
                <div style={{ maxWidth: 600 }}>
                  <Alert
                    message="Change Password"
                    description="Ensure your password is strong and secure"
                    type="warning"
                    showIcon
                    style={{ marginBottom: 24 }}
                  />
                  
                  <Form
                    form={passwordForm}
                    layout="vertical"
                  >
                    <Form.Item
                      name="current_password"
                      label="Current Password"
                      rules={[
                        { required: true, message: 'Please enter current password' },
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Enter current password"
                      />
                    </Form.Item>

                    <Form.Item
                      name="new_password"
                      label="New Password"
                      rules={[
                        { required: true, message: 'Please enter new password' },
                        { min: 6, message: 'Password must be at least 6 characters' },
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Enter new password"
                      />
                    </Form.Item>

                    <Form.Item
                      name="confirm_password"
                      label="Confirm Password"
                      dependencies={['new_password']}
                      rules={[
                        { required: true, message: 'Please confirm password' },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            if (!value || getFieldValue('new_password') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(new Error('Passwords do not match'));
                          },
                        }),
                      ]}
                    >
                      <Input.Password
                        prefix={<LockOutlined />}
                        placeholder="Confirm new password"
                      />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        icon={<SaveOutlined />}
                        onClick={handlePasswordChange}
                        loading={loading}
                      >
                        Change Password
                      </Button>
                    </Form.Item>
                  </Form>
                </div>
              ),
            },
            {
              key: 'notifications',
              label: (
                <span>
                  <BellOutlined />
                  Notifications
                </span>
              ),
              children: (
                <div style={{ maxWidth: 600 }}>
                  <Alert
                    message="Notification Preferences"
                    description="Configure how you want to receive notifications"
                    type="info"
                    showIcon
                    style={{ marginBottom: 24 }}
                  />
                  
                  <Form
                    form={notificationForm}
                    layout="vertical"
                    initialValues={{
                      email_notifications: true,
                      process_alerts: true,
                      system_alerts: true,
                      weekly_report: false,
                    }}
                  >
                    <Form.Item
                      name="email_notifications"
                      label="Email Notifications"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Divider />

                    <Form.Item
                      name="process_alerts"
                      label="Process Alerts"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      name="system_alerts"
                      label="System Alerts"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      name="node_status_changes"
                      label="Node Status Changes"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Divider />

                    <Form.Item
                      name="weekly_report"
                      label="Weekly Summary Report"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        icon={<SaveOutlined />}
                        onClick={handleNotificationUpdate}
                        loading={loading}
                      >
                        Save Preferences
                      </Button>
                    </Form.Item>
                  </Form>
                </div>
              ),
            },
            {
              key: 'system',
              label: (
                <span>
                  <SettingOutlined />
                  System
                </span>
              ),
              children: user?.is_admin ? (
                <div style={{ maxWidth: 600 }}>
                  <Alert
                    message="System Configuration"
                    description="Configure global system settings (Admin only)"
                    type="warning"
                    showIcon
                    style={{ marginBottom: 24 }}
                  />
                  
                  <Form
                    form={systemForm}
                    layout="vertical"
                    initialValues={{
                      refresh_interval: 30,
                      log_retention_days: 30,
                      max_concurrent_connections: 100,
                      enable_websocket: true,
                      enable_activity_logging: true,
                    }}
                  >
                    <Form.Item
                      name="refresh_interval"
                      label="Refresh Interval (seconds)"
                      rules={[{ required: true }]}
                    >
                      <InputNumber min={5} max={300} style={{ width: '100%' }} />
                    </Form.Item>

                    <Form.Item
                      name="log_retention_days"
                      label="Log Retention (days)"
                      rules={[{ required: true }]}
                    >
                      <InputNumber min={1} max={365} style={{ width: '100%' }} />
                    </Form.Item>

                    <Form.Item
                      name="max_concurrent_connections"
                      label="Max Concurrent Connections"
                      rules={[{ required: true }]}
                    >
                      <InputNumber min={10} max={1000} style={{ width: '100%' }} />
                    </Form.Item>

                    <Divider />

                    <Form.Item
                      name="enable_websocket"
                      label="Enable WebSocket"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item
                      name="enable_activity_logging"
                      label="Enable Activity Logging"
                      valuePropName="checked"
                    >
                      <Switch />
                    </Form.Item>

                    <Form.Item>
                      <Space>
                        <Button
                          type="primary"
                          icon={<SaveOutlined />}
                          onClick={handleSystemUpdate}
                          loading={loading}
                        >
                          Save Settings
                        </Button>
                        <Button
                          icon={<DatabaseOutlined />}
                          onClick={() => message.info('Database backup started')}
                        >
                          Backup Database
                        </Button>
                      </Space>
                    </Form.Item>
                  </Form>
                </div>
              ) : (
                <Alert
                  message="Access Denied"
                  description="Only administrators can access system settings"
                  type="error"
                  showIcon
                />
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
};

export default Settings;
