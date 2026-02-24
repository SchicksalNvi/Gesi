import { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Button,
  Switch,
  InputNumber,
  message,
  Space,
  Divider,
  Alert,
} from 'antd';
import {
  SaveOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';
import { useStore } from '@/store';
import { settingsApi, SystemSettings } from '@/api/settings';

const Settings: React.FC = () => {
  const { user } = useStore();
  const [loading, setLoading] = useState(false);
  const [systemForm] = Form.useForm();

  // 加载系统设置
  useEffect(() => {
    const loadSystemSettings = async () => {
      if (!user?.is_admin) return;
      
      try {
        const response = await settingsApi.getSystemSettings();
        const settings = response.settings || {};
        const wsEnabled = settings.enable_websocket !== 'false';
        systemForm.setFieldsValue({
          refresh_interval: parseInt(settings.refresh_interval) || 30,
          process_refresh_interval: parseInt(settings.process_refresh_interval) || 5,
          log_retention_days: parseInt(settings.log_retention_days) || 30,
          max_concurrent_connections: parseInt(settings.max_concurrent_connections) || 100,
          enable_websocket: wsEnabled,
          enable_activity_logging: settings.enable_activity_logging !== 'false',
        });
        // Sync WebSocket setting to store
        const { setWebsocketEnabled } = useStore.getState();
        setWebsocketEnabled(wsEnabled);
      } catch (error) {
        console.error('Failed to load system settings:', error);
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

  const handleSystemUpdate = async () => {
    try {
      const values = await systemForm.validateFields();
      setLoading(true);
      
      const systemSettings: Partial<SystemSettings> = {
        refresh_interval: values.refresh_interval,
        process_refresh_interval: values.process_refresh_interval,
        log_retention_days: values.log_retention_days,
        max_concurrent_connections: values.max_concurrent_connections,
        enable_websocket: values.enable_websocket,
        enable_activity_logging: values.enable_activity_logging,
      };
      
      await settingsApi.updateSystemSettings(systemSettings);
      
      // Update local store for WebSocket setting
      const { setWebsocketEnabled } = useStore.getState();
      setWebsocketEnabled(values.enable_websocket);
      
      message.success('System settings updated');
    } catch (error: any) {
      console.error('Failed to update system settings:', error);
      message.error(error.response?.data?.message || 'Failed to update system settings');
    } finally {
      setLoading(false);
    }
  };

  if (!user?.is_admin) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          message="Access Denied"
          description="Only administrators can access system settings"
          type="error"
          showIcon
        />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 16 }}>System Settings</h1>

      <Card>
        <Alert
          message="System Configuration"
          description="Configure global system settings. Changes will affect all users."
          type="warning"
          showIcon
          style={{ marginBottom: 24 }}
        />
        
        <Form
          form={systemForm}
          layout="vertical"
          style={{ maxWidth: 600 }}
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
            help="When disabled, real-time updates will not be available"
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
      </Card>
    </div>
  );
};

export default Settings;
