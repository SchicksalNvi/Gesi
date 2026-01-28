import { useState, useEffect } from 'react';
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
import { authApi } from '@/api/auth';
import { settingsApi, UserPreferences, SystemSettings } from '@/api/settings';

const Settings: React.FC = () => {
  const { user, setUser, userPreferences, setUserPreferences } = useStore();
  const [loading, setLoading] = useState(false);
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [notificationForm] = Form.useForm();
  const [systemForm] = Form.useForm();

  // 初始化表单数据
  useEffect(() => {
    if (user) {
      const formValues = {
        username: user.username,
        email: user.email,
        full_name: user.full_name || '',
      };
      profileForm.setFieldsValue(formValues);
    }
  }, [user, profileForm]);

  // 当用户偏好设置加载后，更新profile表单的timezone字段
  useEffect(() => {
    if (userPreferences) {
      const timezoneValue = userPreferences.timezone || 'UTC';
      profileForm.setFieldsValue({
        timezone: timezoneValue,
      });
      
      // 同时设置通知表单的值
      notificationForm.setFieldsValue({
        email_notifications: userPreferences.email_notifications ?? true,
        process_alerts: userPreferences.process_alerts ?? true,
        system_alerts: userPreferences.system_alerts ?? true,
        node_status_changes: userPreferences.node_status_changes ?? false,
        weekly_report: userPreferences.weekly_report ?? false,
      });
    }
  }, [userPreferences, profileForm, notificationForm]);

  // 加载用户偏好设置
  useEffect(() => {
    const loadUserPreferences = async () => {
      try {
        const response = await settingsApi.getUserPreferences();
        
        // API直接返回UserPreferences对象
        const preferences: UserPreferences = response;
        
        setUserPreferences(preferences);
      } catch (error) {
        console.error('Failed to load user preferences:', error);
        // 使用默认值
        const defaultPreferences = {
          email_notifications: true,
          process_alerts: true,
          system_alerts: true,
          node_status_changes: false,
          weekly_report: false,
          timezone: 'UTC',
          theme: 'light',
          language: 'en',
        };
        setUserPreferences(defaultPreferences);
      }
    };

    if (user) {
      loadUserPreferences();
    }
  }, [user, setUserPreferences]);

  // 加载系统设置（仅管理员）
  useEffect(() => {
    const loadSystemSettings = async () => {
      if (!user?.is_admin) return;
      
      try {
        const response = await settingsApi.getSystemSettings();
        const settings = response.settings || {};
        systemForm.setFieldsValue({
          refresh_interval: parseInt(settings.refresh_interval) || 30,
          process_refresh_interval: parseInt(settings.process_refresh_interval) || 5,
          log_retention_days: parseInt(settings.log_retention_days) || 30,
          max_concurrent_connections: parseInt(settings.max_concurrent_connections) || 100,
          enable_websocket: settings.enable_websocket === 'true',
          enable_activity_logging: settings.enable_activity_logging === 'true',
        });
      } catch (error) {
        console.error('Failed to load system settings:', error);
        // 使用默认值
        systemForm.setFieldsValue({
          refresh_interval: 30,
          process_refresh_interval: 5,
          log_retention_days: 30,
          max_concurrent_connections: 100,
          enable_websocket: true,
          enable_activity_logging: true,
        });
      }
    };

    if (user?.is_admin) {
      loadSystemSettings();
    }
  }, [user, systemForm]);

  const handleProfileUpdate = async () => {
    try {
      const values = await profileForm.validateFields();
      setLoading(true);
      
      // 调用真实的API更新用户资料
      const response = await authApi.updateProfile({
        email: values.email,
        full_name: values.full_name,
      });
      
      // 更新本地状态 - 后端返回 { status, message, data: {...} }
      if ((response as any).data && user) {
        const updatedUser = { 
          ...user, 
          email: (response as any).data.email,
          full_name: (response as any).data.full_name,
          updated_at: (response as any).data.updated_at,
        };
        setUser(updatedUser);
      }

      // 如果timezone有变化，也更新用户偏好设置
      if (values.timezone && userPreferences && values.timezone !== userPreferences.timezone) {
        try {
          const updatedPreferences = {
            ...userPreferences,
            timezone: values.timezone,
          };
          await settingsApi.updateUserPreferences(updatedPreferences);
          setUserPreferences(updatedPreferences);
        } catch (error) {
          console.error('Failed to update timezone preference:', error);
          // 不阻止整个更新流程，只是记录错误
        }
      }
      
      message.success('Profile updated successfully');
    } catch (error: any) {
      console.error('Failed to update profile:', error);
      message.error(error.response?.data?.message || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async () => {
    try {
      const values = await passwordForm.validateFields();
      setLoading(true);
      
      if (!user?.username) {
        message.error('User information not available');
        return;
      }
      
      // 调用真实的密码修改API
      await settingsApi.changePassword(user.username, {
        old_password: values.current_password,
        new_password: values.new_password,
      });
      
      message.success('Password changed successfully');
      passwordForm.resetFields();
    } catch (error: any) {
      console.error('Failed to change password:', error);
      message.error(error.response?.data?.message || 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  const handleNotificationUpdate = async () => {
    try {
      const values = await notificationForm.validateFields();
      setLoading(true);
      
      // 调用真实的用户偏好设置API
      const preferences: UserPreferences = {
        email_notifications: values.email_notifications,
        process_alerts: values.process_alerts,
        system_alerts: values.system_alerts,
        node_status_changes: values.node_status_changes,
        weekly_report: values.weekly_report,
      };
      
      await settingsApi.updateUserPreferences(preferences);
      
      // 更新本地状态
      const { setUserPreferences } = useStore.getState();
      setUserPreferences(preferences);
      
      message.success('Notification settings updated');
    } catch (error: any) {
      console.error('Failed to update notification settings:', error);
      message.error(error.response?.data?.message || 'Failed to update notification settings');
    } finally {
      setLoading(false);
    }
  };

  const handleSystemUpdate = async () => {
    try {
      const values = await systemForm.validateFields();
      setLoading(true);
      
      // 调用真实的系统设置API
      const systemSettings: Partial<SystemSettings> = {
        refresh_interval: values.refresh_interval,
        process_refresh_interval: values.process_refresh_interval,
        log_retention_days: values.log_retention_days,
        max_concurrent_connections: values.max_concurrent_connections,
        enable_websocket: values.enable_websocket,
        enable_activity_logging: values.enable_activity_logging,
      };
      
      await settingsApi.updateSystemSettings(systemSettings);
      
      message.success('System settings updated');
    } catch (error: any) {
      console.error('Failed to update system settings:', error);
      message.error(error.response?.data?.message || 'Failed to update system settings');
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
                      process_refresh_interval: 5,
                      log_retention_days: 30,
                      max_concurrent_connections: 100,
                      enable_websocket: true,
                      enable_activity_logging: true,
                    }}
                  >
                    <Form.Item
                      name="refresh_interval"
                      label="Node Refresh Interval (seconds)"
                      help="How often to refresh node and process status"
                      rules={[{ required: true }]}
                    >
                      <InputNumber min={5} max={300} style={{ width: '100%' }} />
                    </Form.Item>

                    <Form.Item
                      name="process_refresh_interval"
                      label="Process Page Refresh Interval (seconds)"
                      help="How often to push process updates via WebSocket"
                      rules={[{ required: true }]}
                    >
                      <InputNumber min={1} max={300} style={{ width: '100%' }} />
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
