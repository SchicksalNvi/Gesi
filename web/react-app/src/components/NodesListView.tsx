import React, { useMemo, useState } from 'react';
import { Table, Tag, Button, Space, Tooltip, Typography, Alert } from 'antd';
import type { ColumnsType, TableProps } from 'antd/es/table';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { Node } from '@/types';
import { VirtualizedNodesTable } from './VirtualizedNodesTable';
import { useStore } from '@/store';

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
  const { t } = useStore();
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
              message={t.nodesListView.largeDataset.replace('{count}', String(nodes.length))}
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
              action={
                <Button
                  size="small"
                  type="text"
                  onClick={() => setUseVirtualization(false)}
                >
                  {t.nodesListView.useStandardTable}
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
      title: t.common.status,
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
      title: t.common.name,
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
      title: t.nodes.environment,
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
      title: t.nodesListView.hostPort,
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
      title: t.nodesListView.user,
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
      title: t.nodes.processes,
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
      title: t.nodesListView.connectionStatus,
      key: 'connectionStatus',
      width: 100,
      responsive: ['sm'],
      render: (_, record) => (
        <Tag
          icon={record.is_connected ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
          color={record.is_connected ? 'success' : 'error'}
        >
          {record.is_connected ? t.nodes.online : t.nodes.offline}
        </Tag>
      ),
    },
    {
      title: t.common.actions,
      key: 'actions',
      width: 120,
      align: 'center',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title={t.nodesListView.viewDetails}>
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
            <Tooltip title={t.nodesListView.refreshNode}>
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
          t.nodesListView.ofNodes.replace('{start}', String(range[0])).replace('{end}', String(range[1])).replace('{total}', String(total)),
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
          ? t.nodesListView.noNodesMatching.replace('{query}', searchQuery)
          : t.nodesListView.noNodesAvailable
      }}
    />
  );
};

export default NodesListView;