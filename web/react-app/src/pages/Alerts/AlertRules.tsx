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
  metric: string;
  condition: string;
  threshold: number;
  duration: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  enabled: boolean;
  node_id?: number;
  process_name?: string;
  tags?: string;
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
      const response = await fetch('/api/alerts/rules', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch alert rules');
      }
      
      const data = await response.json();
      setRules(data.data || []);
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
      const response = await fetch(`/api/alerts/rules/${ruleId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete alert rule');
      }
      
      message.success('Alert rule deleted');
      loadRules();
    } catch (error) {
      console.error('Failed to delete alert rule:', error);
      message.error('Failed to delete alert rule');
    }
  };

  const handleToggle = async (ruleId: number, enabled: boolean) => {
    try {
      const response = await fetch(`/api/alerts/rules/${ruleId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ enabled }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to update alert rule');
      }
      
      message.success(`Alert rule ${enabled ? 'enabled' : 'disabled'}`);
      loadRules();
    } catch (error) {
      console.error('Failed to update alert rule:', error);
      message.error('Failed to update alert rule');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      const url = editingRule 
        ? `/api/alerts/rules/${editingRule.id}`
        : '/api/alerts/rules';
      
      const method = editingRule ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(values),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to ${editingRule ? 'update' : 'create'} alert rule`);
      }
      
      message.success(`Alert rule ${editingRule ? 'updated' : 'created'}`);
      setModalVisible(false);
      loadRules();
    } catch (error) {
      console.error('Failed to submit:', error);
      message.error(`Failed to ${editingRule ? 'update' : 'create'} alert rule`);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'red';
      case 'high': return 'orange';
      case 'medium': return 'gold';
      case 'low': return 'blue';
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
      title: 'Metric',
      dataIndex: 'metric',
      key: 'metric',
      render: (metric: string) => metric.replace(/_/g, ' ').toUpperCase(),
    },
    {
      title: 'Condition',
      key: 'condition',
      render: (_, record: AlertRule) => `${record.condition} ${record.threshold}`,
    },
    {
      title: 'Duration',
      dataIndex: 'duration',
      key: 'duration',
      render: (duration: number) => `${duration}s`,
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
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: AlertRule) => {
        const isSystemRule = record.id <= 2;
        return (
          <Switch
            checked={enabled}
            onChange={(checked) => handleToggle(record.id, checked)}
            disabled={isSystemRule}
            title={isSystemRule ? 'System rules are always enabled' : 'Toggle rule status'}
          />
        );
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record: AlertRule) => {
        const isSystemRule = record.id <= 2; // System rules have ID 1 and 2
        return (
          <Space>
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={isSystemRule}
              title={isSystemRule ? 'System rules cannot be edited' : 'Edit rule'}
            >
              Edit
            </Button>
            <Popconfirm
              title="Delete this rule?"
              onConfirm={() => handleDelete(record.id)}
              okText="Yes"
              cancelText="No"
              disabled={isSystemRule}
            >
              <Button
                size="small"
                icon={<DeleteOutlined />}
                danger
                disabled={isSystemRule}
                title={isSystemRule ? 'System rules cannot be deleted' : 'Delete rule'}
              >
                Delete
              </Button>
            </Popconfirm>
          </Space>
        );
      },
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
            icon={<ReloadOutlined />}
            onClick={loadRules}
            loading={loading}
          >
            Refresh
          </Button>
        </Space>
      </div>

      {/* System Rules Info */}
      <Card style={{ marginBottom: 16, backgroundColor: '#f0f5ff', borderColor: '#adc6ff' }}>
        <Space direction="vertical" size="small">
          <div style={{ fontWeight: 'bold', color: '#1890ff' }}>
            ℹ️ System Alert Rules
          </div>
          <div style={{ fontSize: '14px', color: '#595959' }}>
            This system uses automatic alert rules for monitoring:
          </div>
          <ul style={{ margin: '8px 0', paddingLeft: '20px', fontSize: '14px', color: '#595959' }}>
            <li><strong>Node Offline Alert</strong>: Automatically triggered when a node becomes unreachable</li>
            <li><strong>Process Stopped Alert</strong>: Automatically triggered when a process stops unexpectedly</li>
          </ul>
          <div style={{ fontSize: '13px', color: '#8c8c8c', fontStyle: 'italic' }}>
            System rules (ID 1-2) cannot be edited or deleted. Custom rules are not supported in this version.
          </div>
        </Space>
      </Card>

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
            severity: 'high',
            condition: '==',
            threshold: 0,
            duration: 1,
          }}
        >
          <Form.Item
            name="name"
            label="Rule Name"
            rules={[{ required: true, message: 'Please enter rule name' }]}
          >
            <Input placeholder="e.g., Node Offline Alert" />
          </Form.Item>

          <Form.Item
            name="description"
            label="Description"
          >
            <Input.TextArea rows={2} placeholder="Describe when this alert should trigger" />
          </Form.Item>

          <Form.Item
            name="metric"
            label="Metric"
            rules={[{ required: true, message: 'Please select metric' }]}
          >
            <Select
              options={[
                { label: 'Node Status', value: 'node_status' },
                { label: 'Process Status', value: 'process_status' },
                { label: 'CPU Usage', value: 'cpu' },
                { label: 'Memory Usage', value: 'memory' },
                { label: 'Disk Usage', value: 'disk' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="condition"
            label="Condition"
            rules={[{ required: true, message: 'Please select condition' }]}
          >
            <Select
              options={[
                { label: 'Equal (==)', value: '==' },
                { label: 'Not Equal (!=)', value: '!=' },
                { label: 'Greater Than (>)', value: '>' },
                { label: 'Greater or Equal (>=)', value: '>=' },
                { label: 'Less Than (<)', value: '<' },
                { label: 'Less or Equal (<=)', value: '<=' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="threshold"
            label="Threshold"
            rules={[{ required: true, message: 'Please enter threshold' }]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="duration"
            label="Duration (seconds)"
            rules={[{ required: true, message: 'Please enter duration' }]}
          >
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="severity"
            label="Severity"
            rules={[{ required: true, message: 'Please select severity' }]}
          >
            <Select
              options={[
                { label: 'Low', value: 'low' },
                { label: 'Medium', value: 'medium' },
                { label: 'High', value: 'high' },
                { label: 'Critical', value: 'critical' },
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
