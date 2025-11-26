import { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Table, Tag, Space, Button } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  PlayCircleOutlined,
  StopOutlined,
  BellOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { nodesApi } from '@/api/nodes';
import { useStore } from '@/store';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Node, SystemStats } from '@/types';
import ReactECharts from 'echarts-for-react';

export default function Dashboard() {
  const navigate = useNavigate();
  const { nodes, setNodes, systemStats, setSystemStats } = useStore();
  const [loading, setLoading] = useState(false);
  const [cpuData, setCpuData] = useState<number[]>([]);
  const [memoryData, setMemoryData] = useState<number[]>([]);
  const [timestamps, setTimestamps] = useState<string[]>([]);

  // Load initial data
  useEffect(() => {
    loadNodes();
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

  // WebSocket real-time updates
  useWebSocket({
    onMessage: (message) => {
      if (message.type === 'nodes_update') {
        setNodes(message.data);
      } else if (message.type === 'system_stats') {
        setSystemStats(message.data);
        
        // Update chart data
        const now = new Date().toLocaleTimeString();
        setTimestamps(prev => [...prev.slice(-19), now]);
        setCpuData(prev => [...prev.slice(-19), Math.random() * 100]);
        setMemoryData(prev => [...prev.slice(-19), Math.random() * 100]);
      }
    },
  });

  // Calculate stats
  const totalNodes = nodes.length;
  const onlineNodes = nodes.filter(n => n.is_connected).length;
  const offlineNodes = totalNodes - onlineNodes;
  const totalProcesses = systemStats?.running_processes || 0;
  const activeAlerts = systemStats?.active_alerts || 0;

  // Table columns
  const columns: ColumnsType<Node> = [
    {
      title: 'Node Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <a onClick={() => navigate(`/nodes/${text}`)}>{text}</a>,
    },
    {
      title: 'Environment',
      dataIndex: 'environment',
      key: 'environment',
      render: (env) => <Tag color="blue">{env}</Tag>,
    },
    {
      title: 'Host',
      key: 'host',
      render: (_, record) => `${record.host}:${record.port}`,
    },
    {
      title: 'Status',
      dataIndex: 'is_connected',
      key: 'status',
      render: (connected) =>
        connected ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            Online
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">
            Offline
          </Tag>
        ),
    },
    {
      title: 'Processes',
      dataIndex: 'process_count',
      key: 'process_count',
      render: (count) => count || 0,
    },
    {
      title: 'Action',
      key: 'action',
      render: (_, record) => (
        <Button
          type="link"
          size="small"
          onClick={() => navigate(`/nodes/${record.name}`)}
        >
          View Details
        </Button>
      ),
    },
  ];

  // CPU Chart Option
  const cpuChartOption = {
    title: { text: 'CPU Usage', left: 'center' },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: timestamps,
      boundaryGap: false,
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLabel: { formatter: '{value}%' },
    },
    series: [
      {
        name: 'CPU',
        type: 'line',
        data: cpuData,
        smooth: true,
        areaStyle: { opacity: 0.3 },
        itemStyle: { color: '#1890ff' },
      },
    ],
    grid: { left: 50, right: 20, bottom: 30, top: 50 },
  };

  // Memory Chart Option
  const memoryChartOption = {
    title: { text: 'Memory Usage', left: 'center' },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: timestamps,
      boundaryGap: false,
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLabel: { formatter: '{value}%' },
    },
    series: [
      {
        name: 'Memory',
        type: 'line',
        data: memoryData,
        smooth: true,
        areaStyle: { opacity: 0.3 },
        itemStyle: { color: '#52c41a' },
      },
    ],
    grid: { left: 50, right: 20, bottom: 30, top: 50 },
  };

  return (
    <div>
      {/* Statistics Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Nodes"
              value={totalNodes}
              prefix={null}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Online Nodes"
              value={onlineNodes}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Running Processes"
              value={totalProcesses}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Active Alerts"
              value={activeAlerts}
              prefix={<BellOutlined />}
              valueStyle={{ color: activeAlerts > 0 ? '#cf1322' : '#999' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Charts */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card>
            <ReactECharts option={cpuChartOption} style={{ height: 300 }} />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card>
            <ReactECharts option={memoryChartOption} style={{ height: 300 }} />
          </Card>
        </Col>
      </Row>

      {/* Nodes Table */}
      <Card
        title="Nodes Overview"
        extra={
          <Button type="primary" onClick={loadNodes} loading={loading}>
            Refresh
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
