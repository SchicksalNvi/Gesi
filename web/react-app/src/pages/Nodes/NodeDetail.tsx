import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Table,
  Tag,
  Button,
  Space,
  Descriptions,
  Tabs,
  message,
  Popconfirm,
  Spin,
  Row,
  Col,
  Statistic,
} from 'antd';
import {
  ArrowLeftOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { nodesApi } from '@/api/nodes';
import { Node, Process } from '@/types';
import LogViewer from '@/components/LogViewer';

import { useStore } from '@/store';

const NodeDetail: React.FC = () => {
  const { nodeName } = useParams<{ nodeName: string }>();
  const navigate = useNavigate();
  const { t } = useStore();
  const [node, setNode] = useState<Node | null>(null);
  const [processes, setProcesses] = useState<Process[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({});
  const [logViewerVisible, setLogViewerVisible] = useState(false);
  const [selectedProcess, setSelectedProcess] = useState<Process | null>(null);

  useEffect(() => {
    if (nodeName) {
      loadNodeDetail();
    }
  }, [nodeName]);

  const loadNodeDetail = async () => {
    if (!nodeName) return;
    
    setLoading(true);
    try {
      // 加载节点信息
      const nodesResponse = await nodesApi.getNodes();
      // 后端直接返回 { status, nodes }，不是嵌套在 data 里
      const foundNode = (nodesResponse as any).nodes?.find((n: Node) => n.name === nodeName);
      
      if (foundNode) {
        setNode(foundNode);
        
        // 加载进程列表
        if (foundNode.is_connected) {
          const processResponse = await nodesApi.getNodeProcesses(nodeName);
          // 后端直接返回 { status, processes }
          setProcesses((processResponse as any).processes || []);
        }
      } else {
        message.error('Node not found');
        navigate('/nodes');
      }
    } catch (error) {
      console.error('Failed to load node detail:', error);
      message.error('Failed to load node detail');
    } finally {
      setLoading(false);
    }
  };

  const handleProcessAction = async (
    processName: string,
    action: 'start' | 'stop' | 'restart'
  ) => {
    if (!nodeName) return;
    
    const actionKey = `${processName}-${action}`;
    setActionLoading(prev => ({ ...prev, [actionKey]: true }));
    
    try {
      switch (action) {
        case 'start':
          await nodesApi.startProcess(nodeName, processName);
          message.success(`Started ${processName}`);
          break;
        case 'stop':
          await nodesApi.stopProcess(nodeName, processName);
          message.success(`Stopped ${processName}`);
          break;
        case 'restart':
          await nodesApi.restartProcess(nodeName, processName);
          message.success(`Restarted ${processName}`);
          break;
      }
      // 重新加载进程列表
      await loadNodeDetail();
    } catch (error) {
      console.error(`Failed to ${action} process:`, error);
      message.error(`Failed to ${action} process`);
    } finally {
      setActionLoading(prev => ({ ...prev, [actionKey]: false }));
    }
  };

  const handleViewLogs = (process: Process) => {
    setSelectedProcess(process);
    setLogViewerVisible(true);
  };

  const getProcessStateColor = (state: number) => {
    switch (state) {
      case 20: return 'success'; // RUNNING
      case 0: return 'default'; // STOPPED
      case 10: return 'processing'; // STARTING
      case 30: return 'warning'; // BACKOFF
      case 100: return 'error'; // FATAL
      default: return 'default';
    }
  };

  const processColumns = [
    {
      title: t.common.name,
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <strong>{name}</strong>,
    },
    {
      title: t.processes.processGroup,
      dataIndex: 'group',
      key: 'group',
    },
    {
      title: t.processes.processState,
      dataIndex: 'state_string',
      key: 'state_string',
      render: (state: string, record: Process) => (
        <Tag color={getProcessStateColor(record.state)}>
          {state}
        </Tag>
      ),
    },
    {
      title: t.processes.pid,
      dataIndex: 'pid',
      key: 'pid',
      render: (pid: number) => pid || '-',
    },
    {
      title: t.processes.uptime,
      dataIndex: 'uptime_human',
      key: 'uptime_human',
      render: (uptime: string) => uptime || '-',
    },
    {
      title: t.common.actions,
      key: 'actions',
      render: (_: any, record: Process) => (
        <Space>
          {record.state !== 20 && (
            <Button
              type="primary"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleProcessAction(record.name, 'start')}
              loading={actionLoading[`${record.name}-start`]}
            >
              {t.processes.start}
            </Button>
          )}
          {record.state === 20 && (
            <Popconfirm
              title={t.nodeDetail.stopProcess}
              onConfirm={() => handleProcessAction(record.name, 'stop')}
              okText={t.common.yes}
              cancelText={t.common.no}
            >
              <Button
                size="small"
                icon={<StopOutlined />}
                loading={actionLoading[`${record.name}-stop`]}
                danger
              >
                {t.processes.stop}
              </Button>
            </Popconfirm>
          )}
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => handleProcessAction(record.name, 'restart')}
            loading={actionLoading[`${record.name}-restart`]}
          >
            {t.processes.restart}
          </Button>
          <Button
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => handleViewLogs(record)}
          >
            {t.logs.title}
          </Button>
        </Space>
      ),
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!node) {
    return <div>Node not found</div>;
  }

  const runningProcesses = processes.filter(p => p.state === 20).length;
  const stoppedProcesses = processes.filter(p => p.state === 0).length;
  const totalProcesses = processes.length;

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/nodes')}>
            {t.common.back}
          </Button>
          <h1 style={{ margin: 0 }}>{t.nav.nodes}: {node.name}</h1>
          {node.is_connected ? (
            <Tag icon={<CheckCircleOutlined />} color="success">
              {t.nodeDetail.connected}
            </Tag>
          ) : (
            <Tag icon={<CloseCircleOutlined />} color="error">
              {t.nodeDetail.disconnected}
            </Tag>
          )}
        </Space>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={loadNodeDetail}
          loading={loading}
        >
          {t.common.refresh}
        </Button>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title={t.nodeDetail.totalProcesses}
              value={totalProcesses}
              prefix={<InfoCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title={t.nodeDetail.running}
              value={runningProcesses}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title={t.nodeDetail.stopped}
              value={stoppedProcesses}
              prefix={<StopOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      <Tabs
        defaultActiveKey="processes"
        items={[
          {
            key: 'processes',
            label: t.processes.title,
            children: (
              <Card>
                <Table
                  columns={processColumns}
                  dataSource={processes}
                  rowKey="name"
                  pagination={{ pageSize: 10 }}
                  scroll={{ x: 800 }}
                />
              </Card>
            ),
          },
          {
            key: 'info',
            label: t.nodeDetail.nodeInfo,
            children: (
              <Card>
                <Descriptions bordered column={2}>
                  <Descriptions.Item label={t.common.name}>{node.name}</Descriptions.Item>
                  <Descriptions.Item label={t.nodeDetail.host}>{node.host}</Descriptions.Item>
                  <Descriptions.Item label={t.nodeDetail.port}>{node.port}</Descriptions.Item>
                  <Descriptions.Item label={t.nodes.environment}>
                    <Tag color={node.environment === 'production' ? 'red' : 'blue'}>
                      {node.environment || 'development'}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={t.nodeDetail.username}>{node.username}</Descriptions.Item>
                  <Descriptions.Item label={t.common.status}>
                    <Tag color={node.is_connected ? 'success' : 'error'}>
                      {node.is_connected ? t.nodeDetail.connected : t.nodeDetail.disconnected}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label={t.nodeDetail.lastPing} span={2}>
                    {node.last_ping ? new Date(node.last_ping).toLocaleString() : '-'}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            ),
          },
        ]}
      />

      {/* Enhanced Log Viewer */}
      {selectedProcess && (
        <LogViewer
          visible={logViewerVisible}
          onClose={() => {
            setLogViewerVisible(false);
            setSelectedProcess(null);
          }}
          nodeName={nodeName || ''}
          processName={selectedProcess.name}
        />
      )}
    </div>
  );
};

export default NodeDetail;
