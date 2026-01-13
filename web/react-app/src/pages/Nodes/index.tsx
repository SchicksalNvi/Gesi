import { useEffect, useState } from 'react';
import { Card, Row, Col, Tag, Button, Space, Spin, Empty, message } from 'antd';
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
import { useLocalStorage } from '@/hooks/useLocalStorage';
import { useNodeFiltering } from '@/hooks/useNodeFiltering';
import { useIsMobile } from '@/hooks/useResponsive';
import { 
  NodesToolbar, 
  NodesListView, 
  PaginatedCardView,
  ErrorBoundary,
  ViewMode,
  NodeFilters,
  BulkAction
} from '@/components';
import { usePerformanceMonitor, useListPerformance } from '@/hooks/usePerformanceMonitor';
import { Node } from '@/types';

function NodeList() {
  const navigate = useNavigate();
  const { nodes, setNodes } = useStore();
  const [loading, setLoading] = useState(false);
  const isMobile = useIsMobile();

  // Performance monitoring
  const performanceMetrics = usePerformanceMonitor('NodeList', process.env.NODE_ENV === 'development');
  const { shouldOptimize, recommendations } = useListPerformance(nodes.length);

  // View mode and search/filter state
  const [viewMode, setViewMode] = useLocalStorage<ViewMode>('nodes-view-mode', 'card');
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<NodeFilters>({});
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);

  // Force list view on mobile
  const effectiveViewMode = isMobile ? 'list' : viewMode;

  // Filter nodes
  const { filteredNodes, availableEnvironments, stats } = useNodeFiltering(
    nodes,
    searchQuery,
    filters
  );

  useEffect(() => {
    loadNodes();
  }, []);

  const loadNodes = async () => {
    setLoading(true);
    try {
      const response = await nodesApi.getNodes();
      setNodes(response.data?.nodes || []);
    } catch (error) {
      console.error('Failed to load nodes:', error);
      message.error('Failed to load nodes');
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

  const handleBulkAction = async (action: BulkAction) => {
    try {
      switch (action.type) {
        case 'refresh_all':
          message.info(`Refreshing ${action.nodeIds.length} nodes...`);
          // TODO: Implement bulk refresh
          break;
        case 'restart_all':
          message.info(`Restarting all processes on ${action.nodeIds.length} nodes...`);
          // TODO: Implement bulk restart
          break;
        case 'start_all':
          message.info(`Starting all processes on ${action.nodeIds.length} nodes...`);
          // TODO: Implement bulk start
          break;
        case 'stop_all':
          message.info(`Stopping all processes on ${action.nodeIds.length} nodes...`);
          // TODO: Implement bulk stop
          break;
        case 'delete_selected':
          message.warning('Node deletion not implemented yet');
          break;
      }
      setSelectedNodes([]);
    } catch (error) {
      console.error('Bulk action failed:', error);
      message.error('Bulk action failed');
    }
  };

  if (loading && nodes.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <div>
        <NodesToolbar
          viewMode={effectiveViewMode}
          onViewModeChange={setViewMode}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          filters={filters}
          onFiltersChange={setFilters}
          environments={availableEnvironments}
          totalNodes={stats.total}
          filteredNodes={stats.filtered}
          selectedNodes={selectedNodes}
          onBulkAction={handleBulkAction}
          onRefreshAll={loadNodes}
          loading={loading}
          isMobile={isMobile}
        />

        {/* Performance recommendations */}
        {shouldOptimize && recommendations.length > 0 && process.env.NODE_ENV === 'development' && (
          <div style={{ marginBottom: 16 }}>
            {recommendations.map((rec, index) => (
              <div key={index} style={{ 
                padding: 8, 
                backgroundColor: '#fff7e6', 
                border: '1px solid #ffd591',
                borderRadius: 4,
                marginBottom: 4,
                fontSize: 12,
                color: '#d46b08'
              }}>
                💡 {rec}
              </div>
            ))}
          </div>
        )}

        {filteredNodes.length === 0 ? (
          <Card>
            <Empty
              description={
                stats.total === 0 
                  ? "No nodes configured"
                  : "No nodes match your search criteria"
              }
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            >
              {stats.total === 0 ? (
                <p style={{ color: '#999' }}>
                  Add nodes in your config.toml file
                </p>
              ) : (
                <Button onClick={() => {
                  setSearchQuery('');
                  setFilters({});
                }}>
                  Clear filters
                </Button>
              )}
            </Empty>
          </Card>
        ) : effectiveViewMode === 'list' ? (
          <ErrorBoundary fallback={
            <div style={{ padding: 20, textAlign: 'center' }}>
              <p>Failed to load list view. Please try refreshing the page.</p>
            </div>
          }>
            <NodesListView
              nodes={filteredNodes}
              loading={loading}
              selectedNodes={selectedNodes}
              onSelectionChange={setSelectedNodes}
              onNodeClick={(nodeName) => navigate(`/nodes/${nodeName}`)}
              onRefreshNode={(nodeName) => {
                message.info(`Refreshing node: ${nodeName}`);
                // TODO: Implement single node refresh
              }}
              searchQuery={searchQuery}
            />
          </ErrorBoundary>
        ) : (
          <ErrorBoundary fallback={
            <div style={{ padding: 20, textAlign: 'center' }}>
              <p>Failed to load card view. Please try refreshing the page.</p>
            </div>
          }>
            <PaginatedCardView
              nodes={filteredNodes}
              onNodeClick={(nodeName) => navigate(`/nodes/${nodeName}`)}
              searchQuery={searchQuery}
            />
          </ErrorBoundary>
        )}
      </div>
    </ErrorBoundary>
  );
}

export default NodeList;
