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
  const { user, t } = useStore();
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
          log_retention_days: parseInt(settings.log_retention_days) || 30,
          max_concurrent_connections: parseInt(settings.max_concurrent_connections) || 100,
          enable_websocket: wsEnabled,
          auto_refresh: settings.enable_activity_logging !== 'false',
        });
        // Sync settings to store
        const { setWebsocketEnabled, setAutoRefreshEnabled, setRefreshInterval } = useStore.getState();
        setWebsocketEnabled(wsEnabled);
        setAutoRefreshEnabled(settings.enable_activity_logging !== 'false');
        setRefreshInterval(parseInt(settings.refresh_interval) || 30);
      } catch (error) {
        console.error('Failed to load system settings:', error);
        systemForm.setFieldsValue({
          refresh_interval: 30,
          log_retention_days: 30,
          max_concurrent_connections: 100,
          enable_websocket: true,
          auto_refresh: true,
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
        log_retention_days: values.log_retention_days,
        max_concurrent_connections: values.max_concurrent_connections,
        enable_websocket: values.enable_websocket,
        enable_activity_logging: values.auto_refresh,
      };
      
      await settingsApi.updateSystemSettings(systemSettings);
      
      // Update local store
      const { setWebsocketEnabled, setAutoRefreshEnabled, setRefreshInterval } = useStore.getState();
      setWebsocketEnabled(values.enable_websocket);
      setAutoRefreshEnabled(values.auto_refresh);
      setRefreshInterval(values.refresh_interval);
      
      message.success(t.settings.systemSettingsUpdated);
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
          message={t.common.error}
          description={t.settings.adminOnlyAccess}
          type="error"
          showIcon
        />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <h1 style={{ marginBottom: 16 }}>{t.settings.title}</h1>

      <Card>
        <Alert
          message={t.settings.system}
          description={t.settings.systemSettingsDesc}
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
            log_retention_days: 30,
            max_concurrent_connections: 100,
            enable_websocket: true,
            auto_refresh: true,
          }}
        >
          <Form.Item
            name="refresh_interval"
            label={t.settings.refreshInterval + ' (seconds)'}
            help={t.settings.refreshIntervalHelp}
            rules={[{ required: true }]}
          >
            <InputNumber min={5} max={300} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="log_retention_days"
            label={t.settings.dataRetention + ' (days)'}
            rules={[{ required: true }]}
          >
            <InputNumber min={1} max={365} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="max_concurrent_connections"
            label={t.settings.maxConcurrentConnections}
            rules={[{ required: true }]}
          >
            <InputNumber min={10} max={1000} style={{ width: '100%' }} />
          </Form.Item>

          <Divider />

          <Form.Item
            name="enable_websocket"
            label={t.settings.websocketEnabled}
            valuePropName="checked"
            help={t.settings.websocketHelp}
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="auto_refresh"
            label={t.settings.autoRefresh}
            valuePropName="checked"
            help={t.settings.autoRefreshHelp}
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
                {t.settings.saveSettings}
              </Button>
              <Button
                icon={<DatabaseOutlined />}
                onClick={() => message.info(t.settings.databaseBackupStarted)}
              >
                {t.settings.backupNow}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
