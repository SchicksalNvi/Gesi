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
import { useStore } from '@/store';

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
  const { t } = useStore();
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
      title: t.common.name,
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <strong>{name}</strong>,
    },
    {
      title: t.common.description,
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: t.alertRules.metric,
      dataIndex: 'metric',
      key: 'metric',
      render: (metric: string) => metric.replace(/_/g, ' ').toUpperCase(),
    },
    {
      title: t.alerts.condition,
      key: 'condition',
      render: (_, record: AlertRule) => `${record.condition} ${record.threshold}`,
    },
    {
      title: t.alertRules.duration,
      dataIndex: 'duration',
      key: 'duration',
      render: (duration: number) => t.alertRules.durationSeconds.replace('{duration}', String(duration)),
    },
    {
      title: t.alerts.severity,
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => (
        <Tag color={getSeverityColor(severity)}>
          {severity.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: t.common.status,
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: AlertRule) => {
        const isSystemRule = record.id <= 2;
        return (
          <Switch
            checked={enabled}
            onChange={(checked) => handleToggle(record.id, checked)}
            disabled={isSystemRule}
          />
        );
      },
    },
    {
      title: t.common.actions,
      key: 'actions',
      render: (_, record: AlertRule) => {
        const isSystemRule = record.id <= 2;
        return (
          <Space>
            <Button
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
              disabled={isSystemRule}
            >
              {t.common.edit}
            </Button>
            <Popconfirm
              title={t.alerts.deleteRule + '?'}
              onConfirm={() => handleDelete(record.id)}
              okText={t.common.yes}
              cancelText={t.common.no}
              disabled={isSystemRule}
            >
              <Button
                size="small"
                icon={<DeleteOutlined />}
                danger
                disabled={isSystemRule}
              >
                {t.common.delete}
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
        <h1 style={{ margin: 0 }}>{t.alerts.alertRules}</h1>
        <Space>
          <Button onClick={() => navigate('/alerts')}>
            {t.common.back}
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadRules}
            loading={loading}
          >
            {t.common.refresh}
          </Button>
        </Space>
      </div>

      {/* System Rules Info */}
      <Card style={{ marginBottom: 16, backgroundColor: '#f0f5ff', borderColor: '#adc6ff' }}>
        <Space direction="vertical" size="small">
          <div style={{ fontWeight: 'bold', color: '#1890ff' }}>
            ℹ️ {t.alertRules.systemAlertRules}
          </div>
          <div style={{ fontSize: '14px', color: '#595959' }}>
            {t.alertRules.systemAlertDesc}
          </div>
          <ul style={{ margin: '8px 0', paddingLeft: '20px', fontSize: '14px', color: '#595959' }}>
            <li><strong>{t.alertRules.nodeOfflineAlert}</strong>: {t.alertRules.nodeOfflineDesc}</li>
            <li><strong>{t.alertRules.processStoppedAlert}</strong>: {t.alertRules.processStoppedDesc}</li>
          </ul>
          <div style={{ fontSize: '13px', color: '#8c8c8c', fontStyle: 'italic' }}>
            {t.alertRules.systemRulesNote}
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
            showTotal: (total) => t.alertRules.totalRules.replace('{total}', String(total)),
          }}
        />
      </Card>

      {/* 创建/编辑 Modal */}
      <Modal
        title={editingRule ? t.alertRules.editAlertRule : t.alertRules.createAlertRule}
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
            label={t.alertRules.ruleName}
            rules={[{ required: true, message: t.alertRules.enterRuleName }]}
          >
            <Input placeholder={t.alertRules.ruleNamePlaceholder} />
          </Form.Item>

          <Form.Item
            name="description"
            label={t.common.description}
          >
            <Input.TextArea rows={2} placeholder={t.alertRules.descriptionPlaceholder} />
          </Form.Item>

          <Form.Item
            name="metric"
            label={t.alertRules.metric}
            rules={[{ required: true, message: t.alertRules.selectMetric }]}
          >
            <Select
              options={[
                { label: t.alertRules.nodeStatus, value: 'node_status' },
                { label: t.alertRules.processStatus, value: 'process_status' },
                { label: t.alertRules.cpuUsage, value: 'cpu' },
                { label: t.alertRules.memoryUsage, value: 'memory' },
                { label: t.alertRules.diskUsage, value: 'disk' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="condition"
            label={t.alerts.condition}
            rules={[{ required: true, message: t.alertRules.selectCondition }]}
          >
            <Select
              options={[
                { label: t.alertRules.equal, value: '==' },
                { label: t.alertRules.notEqual, value: '!=' },
                { label: t.alertRules.greaterThan, value: '>' },
                { label: t.alertRules.greaterOrEqual, value: '>=' },
                { label: t.alertRules.lessThan, value: '<' },
                { label: t.alertRules.lessOrEqual, value: '<=' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="threshold"
            label={t.alerts.threshold}
            rules={[{ required: true, message: t.alertRules.enterThreshold }]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="duration"
            label={`${t.alertRules.duration} (${t.common.time})`}
            rules={[{ required: true, message: t.alertRules.enterDuration }]}
          >
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="severity"
            label={t.alerts.severity}
            rules={[{ required: true, message: t.alertRules.selectSeverity }]}
          >
            <Select
              options={[
                { label: t.alerts.low, value: 'low' },
                { label: t.alerts.medium, value: 'medium' },
                { label: t.alerts.high, value: 'high' },
                { label: t.alerts.critical, value: 'critical' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="enabled"
            label={t.common.enabled}
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
