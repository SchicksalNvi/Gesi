import { useEffect, useState } from 'react';
import {
  Table,
  Tag,
  Button,
  Space,
  Card,
  message,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  InputNumber,
  Popconfirm,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';

interface AlertRule {
  id: number;
  name: string;
  description: string;
  condition_type: string;
  threshold: number;
  severity: 'info' | 'warning' | 'error' | 'critical';
  enabled: boolean;
  notification_channels: string[];
  created_at: string;
  updated_at: string;
}

const AlertRules: React.FC = () => {
  const navigate = useNavigate();
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadRules();
  }, []);

  const loadRules = async () => {
    setLoading(true);
    try {
      // 模拟数据
      const mockRules: AlertRule[] = [
        {
          id: 1,
          name: 'Process Stopped',
          description: 'Alert when a process stops unexpectedly',
          condition_type: 'process_state',
          threshold: 0,
          severity: 'critical',
          enabled: true,
          notification_channels: ['email', 'slack'],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 2,
          name: 'High Memory Usage',
          description: 'Alert when memory usage exceeds threshold',
          condition_type: 'memory_usage',
          threshold: 80,
          severity: 'warning',
          enabled: true,
          notification_channels: ['email'],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 3,
          name: 'High CPU Usage',
          description: 'Alert when CPU usage exceeds threshold',
          condition_type: 'cpu_usage',
          threshold: 90,
          severity: 'error',
          enabled: false,
          notification_channels: ['slack'],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ];
      
      setRules(mockRules);
    } catch (error) {
      console.error('Failed to load rules:', error);
      message.error('Failed to load alert rules');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingRule(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (rule: AlertRule) => {
    setEditingRule(rule);
    form.setFieldsValue(rule);
    setModalVisible(true);
  };

  const handleDelete = async (ruleId: number) => {
    try {
      // 实际应该调用 API
      message.success('Alert rule deleted');
      loadRules();
    } catch (error) {
      message.error('Failed to delete alert rule');
    }
  };

  const handleToggle = async (ruleId: number, enabled: boolean) => {
    try {
      // 实际应该调用 API
      message.success(`Alert rule ${enabled ? 'enabled' : 'disabled'}`);
      loadRules();
    } catch (error) {
      message.error('Failed to update alert rule');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingRule) {
        // 更新规则
        message.success('Alert rule updated');
      } else {
        // 创建规则
        message.success('Alert rule created');
      }
      
      setModalVisible(false);
      loadRules();
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'error': return 'orange';
      case 'warning': return 'gold';
      case 'info': return 'blue';
      default: return 'default';
    }
  };

  const columns: ColumnsType<AlertRule> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <strong>{name}</strong>,
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Condition',
      dataIndex: 'condition_type',
      key: 'condition_type',
      render: (type: string) => type.replace(/_/g, ' ').toUpperCase(),
    },
    {
      title: 'Threshold',
      dataIndex: 'threshold',
      key: 'threshold',
      render: (threshold: number, record: AlertRule) => {
        if (record.condition_type === 'process_state') return '-';
        return `${threshold}%`;
      },
    },
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => (
        <Tag color={getSeverityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: 'Channels',
      dataIndex: 'notification_channels',
      key: 'notification_channels',
      render: (channels: string[]) => (
        <Space>
          {channels.map(channel => (
            <Tag key={channel}>{channel}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: AlertRule) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggle(record.id, checked)}
        />
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record: AlertRule) => (
        <Space>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            Edit
          </Button>
          <Popconfirm
            title="Delete this rule?"
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              size="small"
              icon={<DeleteOutlined />}
              danger
            >
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>Alert Rules</h1>
        <Space>
          <Button onClick={() => navigate('/alerts')}>
            Back to Alerts
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            Create Rule
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadRules}
            loading={loading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={rules}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `Total ${total} rules`,
          }}
        />
      </Card>

      {/* 创建/编辑 Modal */}
      <Modal
        title={editingRule ? 'Edit Alert Rule' : 'Create Alert Rule'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            enabled: true,
            severity: 'warning',
            notification_channels: ['email'],
          }}
        >
          <Form.Item
            name="name"
            label="Rule Name"
            rules={[{ required: true, message: 'Please enter rule name' }]}
          >
            <Input placeholder="e.g., High Memory Usage" />
          </Form.Item>

          <Form.Item
            name="description"
            label="Description"
            rules={[{ required: true, message: 'Please enter description' }]}
          >
            <Input.TextArea rows={3} placeholder="Describe when this alert should trigger" />
          </Form.Item>

          <Form.Item
            name="condition_type"
            label="Condition Type"
            rules={[{ required: true, message: 'Please select condition type' }]}
          >
            <Select
              options={[
                { label: 'Process State', value: 'process_state' },
                { label: 'CPU Usage', value: 'cpu_usage' },
                { label: 'Memory Usage', value: 'memory_usage' },
                { label: 'Disk Usage', value: 'disk_usage' },
                { label: 'Process Restart', value: 'process_restart' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="threshold"
            label="Threshold (%)"
            rules={[{ required: true, message: 'Please enter threshold' }]}
          >
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="severity"
            label="Severity"
            rules={[{ required: true, message: 'Please select severity' }]}
          >
            <Select
              options={[
                { label: 'Info', value: 'info' },
                { label: 'Warning', value: 'warning' },
                { label: 'Error', value: 'error' },
                { label: 'Critical', value: 'critical' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="notification_channels"
            label="Notification Channels"
            rules={[{ required: true, message: 'Please select at least one channel' }]}
          >
            <Select
              mode="multiple"
              options={[
                { label: 'Email', value: 'email' },
                { label: 'Slack', value: 'slack' },
                { label: 'Webhook', value: 'webhook' },
                { label: 'SMS', value: 'sms' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="enabled"
            label="Enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AlertRules;
