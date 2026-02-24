import { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Table, Tag, Button } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlayCircleOutlined,
  BellOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { nodesApi } from '@/api/nodes';
import { useStore } from '@/store';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Node } from '@/types';

export default function Dashboard() {
  const navigate = useNavigate();
  const { nodes, setNodes, systemStats, setSystemStats, t } = useStore();
  const [loading, setLoading] = useState(false);
  const [activeAlerts, setActiveAlerts] = useState(0);

  // Load initial data
  useEffect(() => {
    loadNodes();
    loadAlertStats();
  }, []);

  const loadNodes = async () => {
    setLoading(true);
    try {
      const response = await nodesApi.getNodes();
      setNodes(response.nodes || []);
    } catch (error) {
      console.error('Failed to load nodes:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadAlertStats = async () => {
    try {
      const response = await fetch('/api/alerts/statistics?time_range=24h', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        setActiveAlerts(data.active_alerts || 0);
      }
    } catch (error) {
      console.error('Failed to load alert stats:', error);
    }
  };

  // WebSocket real-time updates
  useWebSocket({
    onMessage: (message) => {
      if (message.type === 'nodes_update') {
        setNodes(message.data);
      } else if (message.type === 'system_stats') {
        setSystemStats(message.data);
      }
    },
  });

  // Calculate stats
  const totalNodes = nodes.length;
  const onlineNodes = nodes.filter(n => n.is_connected).length;
  const offlineNodes = totalNodes - onlineNodes;
  const totalProcesses = systemStats?.running_processes || 0;

  // Table columns
  const columns: ColumnsType<Node> = [
    {
      title: t.nodes.nodeName,
      dataIndex: 'name',
      key: 'name',
      render: (text) => <a onClick={() => navigate(`/nodes/${text}`)}>{text}</a>,
    },
    {
      title: t.nodes.environment,
      dataIndex: 'environment',
      key: 'environment',
      render: (env) => <Tag color="blue">{env}</Tag>,
    },
    {
      title: t.nodes.nodeHost,
      key: 'host',
      render: (_, record) => `${record.host}:${record.port}`,
    },
    {
      title: t.common.status,
      dataIndex: 'is_connected',
      key: 'status',
      render: (connected) =>
        connected ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            {t.nodes.online}
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">
            {t.nodes.offline}
          </Tag>
        ),
    },
    {
      title: t.nodes.processes,
      dataIndex: 'process_count',
      key: 'process_count',
      render: (count) => count || 0,
    },
    {
      title: t.common.actions,
      key: 'action',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          onClick={() => navigate(`/nodes/${record.name}`)}
        >
          {t.nodes.viewDetails}
        </Button>
      ),
    },
  ];

  return (
    <div>
      {/* Statistics Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t.dashboard.totalNodes}
              value={totalNodes}
              prefix={null}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t.dashboard.onlineNodes}
              value={onlineNodes}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t.dashboard.runningProcesses}
              value={totalProcesses}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t.alerts.title}
              value={activeAlerts}
              prefix={<BellOutlined />}
              valueStyle={{ color: activeAlerts > 0 ? '#cf1322' : '#999' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Nodes Table */}
      <Card
        title={t.dashboard.nodeStatus}
        extra={
          <Button type="primary" onClick={loadNodes} loading={loading}>
            {t.common.refresh}
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={nodes}
          rowKey="name"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  );
}
