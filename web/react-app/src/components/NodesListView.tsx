import React, { useMemo, useState } from 'react';
import { Table, Tag, Button, Space, Tooltip, Typography, Alert } from 'antd';
import type { ColumnsType, TableProps } from 'antd/es/table';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  ReloadOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { Node } from '@/types';
import { VirtualizedNodesTable } from './VirtualizedNodesTable';

const { Text } = Typography;

export interface NodesListViewProps {
  nodes: Node[];
  loading?: boolean;
  selectedNodes: string[];
  onSelectionChange: (nodeIds: string[]) => void;
  onNodeClick: (nodeName: string) => void;
  onRefreshNode?: (nodeName: string) => void;
  searchQuery?: string;
}

interface NodeListItem extends Node {
  key: string;
  statusColor: string;
  displayName: string;
}

const VIRTUALIZATION_THRESHOLD = 100;

export const NodesListView: React.FC<NodesListViewProps> = ({
  nodes,
  loading = false,
  selectedNodes,
  onSelectionChange,
  onNodeClick,
  onRefreshNode,
  searchQuery = '',
}) => {
  const [sortedInfo, setSortedInfo] = useState<any>({});
  const [useVirtualization, setUseVirtualization] = useState(true);

  // Determine if we should use virtualization
  const shouldVirtualize = nodes.length >= VIRTUALIZATION_THRESHOLD && useVirtualization;

  // If virtualization is enabled and we have enough nodes, use the virtualized table
  if (shouldVirtualize) {
    try {
      return (
        <div>
          {nodes.length >= VIRTUALIZATION_THRESHOLD && (
            <Alert
              message={`Large dataset detected (${nodes.length} nodes). Using virtualized table for better performance.`}
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
              action={
                <Button
                  size="small"
                  type="text"
                  onClick={() => setUseVirtualization(false)}
                >
                  Use Standard Table
                </Button>
              }
            />
          )}
          <VirtualizedNodesTable
            nodes={nodes}
            loading={loading}
            selectedNodes={selectedNodes}
            onSelectionChange={onSelectionChange}
            onNodeClick={onNodeClick}
            onRefreshNode={onRefreshNode}
            searchQuery={searchQuery}
            height={600}
          />
        </div>
      );
    } catch (error) {
      console.warn('Virtual scrolling failed, falling back to standard table:', error);
      setUseVirtualization(false);
    }
  }

  // Transform nodes for table display
  const tableData: NodeListItem[] = useMemo(() => {
    return nodes.map(node => ({
      ...node,
      key: node.name,
      statusColor: node.is_connected ? '#52c41a' : '#ff4d4f',
      displayName: node.name,
    }));
  }, [nodes]);

  // Highlight search matches in text
  const highlightText = (text: string, query: string) => {
    if (!query) return text;
    
    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);
    
    return parts.map((part, index) => 
      regex.test(part) ? (
        <mark key={index} style={{ backgroundColor: '#fff566', padding: 0 }}>
          {part}
        </mark>
      ) : part
    );
  };

  // Table columns configuration
  const columns: ColumnsType<NodeListItem> = [
    {
      title: 'Status',
      dataIndex: 'is_connected',
      key: 'status',
      width: 80,
      align: 'center',
      sorter: (a, b) => Number(a.is_connected) - Number(b.is_connected),
      sortOrder: sortedInfo.columnKey === 'status' ? sortedInfo.order : null,
      render: (isConnected: boolean) => (
        <div
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            backgroundColor: isConnected ? '#52c41a' : '#ff4d4f',
            margin: '0 auto',
          }}
        />
      ),
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      sorter: (a, b) => a.name.localeCompare(b.name),
      sortOrder: sortedInfo.columnKey === 'name' ? sortedInfo.order : null,
      render: (name: string) => (
        <Button
          type="link"
          onClick={() => onNodeClick(name)}
          style={{ padding: 0, height: 'auto', fontWeight: 500 }}
        >
          {highlightText(name, searchQuery)}
        </Button>
      ),
    },
    {
      title: 'Environment',
      dataIndex: 'environment',
      key: 'environment',
      width: 120,
      responsive: ['md'],
      sorter: (a, b) => a.environment.localeCompare(b.environment),
      sortOrder: sortedInfo.columnKey === 'environment' ? sortedInfo.order : null,
      render: (environment: string) => (
        <Tag color="blue">
          {highlightText(environment || 'default', searchQuery)}
        </Tag>
      ),
    },
    {
      title: 'Host:Port',
      key: 'hostPort',
      width: 180,
      responsive: ['lg'],
      render: (_, record) => (
        <Text code style={{ fontSize: 12 }}>
          {highlightText(`${record.host}:${record.port}`, searchQuery)}
        </Text>
      ),
    },
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      width: 100,
      responsive: ['xl'],
      render: (username: string) => (
        username ? (
          <Text type="secondary" style={{ fontSize: 12 }}>
            {highlightText(username, searchQuery)}
          </Text>
        ) : (
          <Text type="secondary" style={{ fontSize: 12 }}>-</Text>
        )
      ),
    },
    {
      title: 'Processes',
      dataIndex: 'process_count',
      key: 'process_count',
      width: 100,
      align: 'center',
      responsive: ['md'],
      sorter: (a, b) => (a.process_count || 0) - (b.process_count || 0),
      sortOrder: sortedInfo.columnKey === 'process_count' ? sortedInfo.order : null,
      render: (count: number) => (
        <Tag color={count > 0 ? 'blue' : 'default'}>
          {count || 0}
        </Tag>
      ),
    },
    {
      title: 'Status',
      key: 'connectionStatus',
      width: 100,
      responsive: ['sm'],
      render: (_, record) => (
        <Tag
          icon={record.is_connected ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
          color={record.is_connected ? 'success' : 'error'}
        >
          {record.is_connected ? 'Online' : 'Offline'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      align: 'center',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="View Details">
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={(e) => {
                e.stopPropagation();
                onNodeClick(record.name);
              }}
            />
          </Tooltip>
          {onRefreshNode && (
            <Tooltip title="Refresh Node">
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  onRefreshNode(record.name);
                }}
              />
            </Tooltip>
          )}
          <Tooltip title="Node Settings">
            <Button
              type="text"
              size="small"
              icon={<SettingOutlined />}
              onClick={(e) => {
                e.stopPropagation();
                // TODO: Implement node settings
                console.log('Node settings for:', record.name);
              }}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  // Row selection configuration
  const rowSelection: TableProps<NodeListItem>['rowSelection'] = {
    selectedRowKeys: selectedNodes,
    onChange: (selectedRowKeys) => {
      onSelectionChange(selectedRowKeys as string[]);
    },
    getCheckboxProps: (record) => ({
      name: record.name,
    }),
  };

  // Handle table change (sorting, pagination, etc.)
  const handleTableChange: TableProps<NodeListItem>['onChange'] = (pagination, filters, sorter) => {
    setSortedInfo(sorter);
  };

  return (
    <Table<NodeListItem>
      columns={columns}
      dataSource={tableData}
      rowSelection={rowSelection}
      loading={loading}
      onChange={handleTableChange}
      pagination={{
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total, range) => 
          `${range[0]}-${range[1]} of ${total} nodes`,
        pageSizeOptions: ['10', '20', '50', '100'],
        defaultPageSize: 20,
      }}
      scroll={{ x: 800 }}
      size="middle"
      rowClassName={(record) => 
        selectedNodes.includes(record.key) ? 'ant-table-row-selected' : ''
      }
      onRow={(record) => ({
        onClick: () => onNodeClick(record.name),
        style: { cursor: 'pointer' },
      })}
      locale={{
        emptyText: searchQuery 
          ? `No nodes found matching "${searchQuery}"`
          : 'No nodes available'
      }}
    />
  );
};

export default NodesListView;