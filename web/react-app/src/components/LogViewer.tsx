// LogViewer v2.1 - Fixed timestamp formatting
import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  Modal,
  Button,
  Switch,
  Space,
  Tag,
  Spin,
  message,
  Input,
  Select,
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ClearOutlined,
  DownloadOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { nodesApi } from '@/api/nodes';
import { useWebSocket } from '@/hooks/useWebSocket';
import { LogEntry, LogStream, LogStreamMessage } from '@/types';
import { useStore } from '@/store';

const { Search } = Input;
const { Option } = Select;

interface LogViewerProps {
  visible: boolean;
  onClose: () => void;
  nodeName: string;
  processName: string;
}

const LogViewer: React.FC<LogViewerProps> = ({
  visible,
  onClose,
  nodeName,
  processName,
}) => {
  const { userPreferences } = useStore();
  const [logEntries, setLogEntries] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [realTimeEnabled, setRealTimeEnabled] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [levelFilter, setLevelFilter] = useState<string>('');
  const [autoScroll, setAutoScroll] = useState(true);
  
  const logContainerRef = useRef<HTMLDivElement>(null);
  const isSubscribedRef = useRef(false);

  // LOGVIEWER_TIMESTAMP_V3 - 根据用户时区设置格式化时间戳
  const formatTimestamp = useCallback((timestamp: string) => {
    // 优先使用用户设置的时区，默认使用 Asia/Shanghai
    const timezone = userPreferences?.timezone || 'Asia/Shanghai';
    console.log('[LogViewer] formatTimestamp - timezone:', timezone, 'timestamp:', timestamp);
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) {
        return timestamp || '----/--/-- --:--:--';
      }
      const formatter = new Intl.DateTimeFormat('en-CA', {
        timeZone: timezone,
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
      });
      const parts = formatter.formatToParts(date);
      const get = (type: string) => parts.find(p => p.type === type)?.value || '00';
      return `${get('year')}-${get('month')}-${get('day')} ${get('hour')}:${get('minute')}:${get('second')}`;
    } catch {
      return timestamp || '----/--/-- --:--:--';
    }
  }, [userPreferences?.timezone]);

  // WebSocket for real-time logs
  const { send, isConnected } = useWebSocket({
    onMessage: (message) => {
      if (message.type === 'log_stream' && message.data) {
        const logData = message.data as LogStreamMessage;
        if (logData.node_name === nodeName && logData.process_name === processName) {
          handleNewLogEntries(logData.entries);
        }
      }
    },
  });

  const handleNewLogEntries = useCallback((newEntries: LogEntry[]) => {
    if (newEntries.length > 0) {
      setLogEntries(prev => {
        // 去重：基于 timestamp + message 组合
        const existingKeys = new Set(
          prev.map(e => `${e.timestamp}|${e.message}`)
        );
        const uniqueNewEntries = newEntries.filter(
          e => !existingKeys.has(`${e.timestamp}|${e.message}`)
        );
        
        if (uniqueNewEntries.length === 0) {
          return prev; // 没有新条目，不更新
        }
        
        const combined = [...prev, ...uniqueNewEntries];
        // Keep only last 1000 entries to prevent memory issues
        return combined.slice(-1000);
      });
      
      // Auto scroll to bottom if enabled
      if (autoScroll && logContainerRef.current) {
        setTimeout(() => {
          logContainerRef.current?.scrollTo({
            top: logContainerRef.current.scrollHeight,
            behavior: 'smooth'
          });
        }, 100);
      }
    }
  }, [autoScroll]);

  const loadInitialLogs = async () => {
    setLoading(true);
    try {
      // 不传 offset，让后端从文件末尾读取最新日志
      const response = await nodesApi.getProcessLogStream(nodeName, processName, undefined, 100);
      const logStream = (response as any).data as LogStream;
      
      setLogEntries(logStream.entries || []);
    } catch (error) {
      console.error('Failed to load logs:', error);
      message.error('Failed to load logs');
    } finally {
      setLoading(false);
    }
  };

  const subscribeToLogs = () => {
    if (isConnected && !isSubscribedRef.current) {
      send({
        type: 'subscribe_logs',
        data: {
          node_name: nodeName,
          process_name: processName,
        },
      });
      isSubscribedRef.current = true;
    }
  };

  const unsubscribeFromLogs = () => {
    if (isConnected && isSubscribedRef.current) {
      send({
        type: 'unsubscribe_logs',
        data: {
          node_name: nodeName,
          process_name: processName,
        },
      });
      isSubscribedRef.current = false;
    }
  };

  const toggleRealTime = (enabled: boolean) => {
    setRealTimeEnabled(enabled);
    if (enabled) {
      subscribeToLogs();
    } else {
      unsubscribeFromLogs();
    }
  };

  const clearLogs = () => {
    setLogEntries([]);
  };

  const exportLogs = () => {
    const filteredEntries = getFilteredEntries();
    const logText = filteredEntries
      .map(entry => `[${entry.timestamp}] [${entry.level}] ${entry.message}`)
      .join('\n');
    
    const blob = new Blob([logText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${nodeName}-${processName}-logs.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const getLogLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR':
        return 'red';
      case 'WARN':
      case 'WARNING':
        return 'orange';
      case 'INFO':
        return 'blue';
      case 'DEBUG':
        return 'green';
      case 'TRACE':
        return 'purple';
      default:
        return 'default';
    }
  };

  const getFilteredEntries = () => {
    return logEntries.filter(entry => {
      const matchesSearch = !searchText || 
        entry.message.toLowerCase().includes(searchText.toLowerCase());
      const matchesLevel = !levelFilter || 
        entry.level.toLowerCase() === levelFilter.toLowerCase();
      return matchesSearch && matchesLevel;
    });
  };

  useEffect(() => {
    if (visible) {
      loadInitialLogs();
    } else {
      // Clean up when modal closes
      unsubscribeFromLogs();
      setRealTimeEnabled(false);
      setLogEntries([]);
    }
  }, [visible]);

  useEffect(() => {
    return () => {
      unsubscribeFromLogs();
    };
  }, []);

  const filteredEntries = getFilteredEntries();

  return (
    <Modal
      title={
        <Space>
          <span>Logs: {processName} on {nodeName}</span>
          <Tag color={realTimeEnabled ? 'green' : 'default'}>
            {realTimeEnabled ? 'Live' : 'Static'}
          </Tag>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={1000}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
    >
      <div style={{ marginBottom: 16 }}>
        <Space wrap>
          <Space>
            <span>Real-time:</span>
            <Switch
              checked={realTimeEnabled}
              onChange={toggleRealTime}
              disabled={!isConnected}
              checkedChildren={<PlayCircleOutlined />}
              unCheckedChildren={<PauseCircleOutlined />}
            />
          </Space>
          
          <Space>
            <span>Auto-scroll:</span>
            <Switch
              checked={autoScroll}
              onChange={setAutoScroll}
              size="small"
            />
          </Space>

          <Search
            placeholder="Search logs..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 200 }}
            prefix={<SearchOutlined />}
            allowClear
          />

          <Select
            placeholder="Filter by level"
            value={levelFilter}
            onChange={setLevelFilter}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="error">Error</Option>
            <Option value="warn">Warning</Option>
            <Option value="info">Info</Option>
            <Option value="debug">Debug</Option>
            <Option value="trace">Trace</Option>
          </Select>

          <Button
            icon={<ClearOutlined />}
            onClick={clearLogs}
            size="small"
          >
            Clear
          </Button>

          <Button
            icon={<DownloadOutlined />}
            onClick={exportLogs}
            size="small"
          >
            Export
          </Button>
        </Space>
      </div>

      <div
        ref={logContainerRef}
        style={{
          height: 500,
          overflow: 'auto',
          border: '1px solid #d9d9d9',
          borderRadius: 4,
          padding: 8,
          backgroundColor: '#fafafa',
        }}
      >
        {loading ? (
          <div style={{ textAlign: 'center', padding: 50 }}>
            <Spin size="large" />
          </div>
        ) : filteredEntries.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 50, color: '#999' }}>
            No log entries found
          </div>
        ) : (
          <div style={{ fontFamily: 'monospace', fontSize: 12 }}>
            {filteredEntries.map((entry, index) => (
              <div 
                key={index} 
                style={{ 
                  display: 'flex', 
                  gap: 8, 
                  padding: '4px 0',
                  alignItems: 'flex-start',
                }}
              >
                <span 
                  style={{ 
                    color: 'rgba(0,0,0,0.45)',
                    width: 152,
                    minWidth: 152,
                    flexShrink: 0,
                  }}
                >
                  {formatTimestamp(entry.timestamp)}
                </span>
                <Tag
                  color={getLogLevelColor(entry.level)}
                  style={{ width: 60, minWidth: 60, textAlign: 'center', flexShrink: 0, margin: 0 }}
                >
                  {entry.level.toUpperCase()}
                </Tag>
                <span style={{ flex: 1, wordBreak: 'break-word' }}>
                  {entry.message}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
        Showing {filteredEntries.length} of {logEntries.length} entries
        {realTimeEnabled && isConnected && (
          <span style={{ marginLeft: 16, color: '#52c41a' }}>
            ● Connected - receiving real-time updates
          </span>
        )}
      </div>
    </Modal>
  );
};

export default LogViewer;