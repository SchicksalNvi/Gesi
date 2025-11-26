import { useEffect, useState } from 'react';
import { Card, Row, Col, Tag, Button, Space, Spin, Empty } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { nodesApi } from '@/api/nodes';
import { useStore } from '@/store';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Node } from '@/types';

export default function NodeList() {
  const navigate = useNavigate();
  const { nodes, setNodes } = useStore();
  const [loading, setLoading] = useState(false);

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
      }
    },
  });

  if (loading && nodes.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 24 }}>Nodes</h2>
          <p style={{ color: '#666', marginTop: 8 }}>
            Manage your supervisor nodes
          </p>
        </div>
        <Button
          type="primary"
          icon={<ReloadOutlined />}
          onClick={loadNodes}
          loading={loading}
        >
          Refresh
        </Button>
      </div>

      {nodes.length === 0 ? (
        <Card>
          <Empty
            description="No nodes configured"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <p style={{ color: '#999' }}>
              Add nodes in your config.toml file
            </p>
          </Empty>
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          {nodes.map((node) => (
            <Col xs={24} sm={12} lg={8} key={node.name}>
              <NodeCard node={node} onView={() => navigate(`/nodes/${node.name}`)} />
            </Col>
          ))}
        </Row>
      )}
    </div>
  );
}

// Node Card Component
function NodeCard({ node, onView }: { node: Node; onView: () => void }) {
  const isOnline = node.is_connected;

  return (
    <Card
      hoverable
      style={{
        height: '100%',
        cursor: 'pointer',
      }}
      onClick={onView}
    >
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                width: 48,
                height: 48,
                borderRadius: 8,
                background: '#1890ff20',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <CheckCircleOutlined style={{ fontSize: 24, color: '#1890ff' }} />
            </div>
            <div>
              <h3 style={{ margin: 0, fontSize: 18 }}>{node.name}</h3>
              <Tag color="blue" style={{ marginTop: 4 }}>
                {node.environment || 'default'}
              </Tag>
            </div>
          </div>
          <div
            style={{
              width: 12,
              height: 12,
              borderRadius: '50%',
              background: isOnline ? '#52c41a' : '#ff4d4f',
            }}
          />
        </div>

        {/* Info */}
        <div>
          <div style={{ color: '#666', fontSize: 14, marginBottom: 4 }}>
            {node.host}:{node.port}
          </div>
          {node.username && (
            <div style={{ color: '#666', fontSize: 14 }}>
              User: {node.username}
            </div>
          )}
        </div>

        {/* Stats */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            paddingTop: 16,
            borderTop: '1px solid #f0f0f0',
          }}
        >
          <div style={{ textAlign: 'center', flex: 1 }}>
            <div style={{ fontSize: 24, fontWeight: 'bold', color: '#1890ff' }}>
              {node.process_count || 0}
            </div>
            <div style={{ fontSize: 12, color: '#999' }}>Processes</div>
          </div>
          <div
            style={{
              textAlign: 'center',
              flex: 1,
              borderLeft: '1px solid #f0f0f0',
            }}
          >
            <Tag
              icon={isOnline ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
              color={isOnline ? 'success' : 'error'}
            >
              {isOnline ? 'Online' : 'Offline'}
            </Tag>
          </div>
        </div>

        {/* Action Button */}
        <Button
          type="primary"
          block
          icon={<EyeOutlined />}
          onClick={(e) => {
            e.stopPropagation();
            onView();
          }}
        >
          View Details
        </Button>
      </Space>
    </Card>
  );
}
