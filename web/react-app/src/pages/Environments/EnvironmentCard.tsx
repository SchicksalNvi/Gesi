import { Card, Space, Tag } from 'antd';
import {
  AppstoreOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { Environment } from '@/types';

interface EnvironmentCardProps {
  environment: Environment;
  onClick: () => void;
}

export default function EnvironmentCard({ environment, onClick }: EnvironmentCardProps) {
  const totalNodes = environment.members.length;
  const onlineNodes = environment.members.filter((node) => node.is_connected).length;
  const offlineNodes = totalNodes - onlineNodes;

  return (
    <Card
      hoverable
      style={{
        height: '100%',
        cursor: 'pointer',
      }}
      onClick={onClick}
    >
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: 8,
              background: '#52c41a20',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <AppstoreOutlined style={{ fontSize: 24, color: '#52c41a' }} />
          </div>
          <div>
            <h3 style={{ margin: 0, fontSize: 18 }}>{environment.name}</h3>
            <p style={{ margin: 0, color: '#999', fontSize: 14 }}>
              {totalNodes} {totalNodes === 1 ? 'node' : 'nodes'}
            </p>
          </div>
        </div>

        {/* Stats */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-around',
            paddingTop: 16,
            borderTop: '1px solid #f0f0f0',
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 24, fontWeight: 'bold', color: '#52c41a' }}>
              {onlineNodes}
            </div>
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              <CheckCircleOutlined style={{ marginRight: 4 }} />
              Online
            </div>
          </div>
          <div
            style={{
              width: 1,
              background: '#f0f0f0',
            }}
          />
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 24, fontWeight: 'bold', color: '#ff4d4f' }}>
              {offlineNodes}
            </div>
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              <CloseCircleOutlined style={{ marginRight: 4 }} />
              Offline
            </div>
          </div>
        </div>

        {/* Status Badge */}
        <div style={{ textAlign: 'center' }}>
          {onlineNodes === totalNodes ? (
            <Tag color="success" style={{ margin: 0 }}>
              All nodes online
            </Tag>
          ) : onlineNodes === 0 ? (
            <Tag color="error" style={{ margin: 0 }}>
              All nodes offline
            </Tag>
          ) : (
            <Tag color="warning" style={{ margin: 0 }}>
              Partial connectivity
            </Tag>
          )}
        </div>
      </Space>
    </Card>
  );
}
