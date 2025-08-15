import React, { createContext, useContext, useEffect, useState, useRef, useCallback } from 'react';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const [socket, setSocket] = useState(null);
  const [connected, setConnected] = useState(false);
  const [realtimeData, setRealtimeData] = useState({
    nodes: [],
    processes: {},
    systemStats: {},
    events: [],
  });
  const socketRef = useRef(null);

  const initializeSocket = useCallback(() => {
    const token = localStorage.getItem('token');
    if (!token) return;

    // 动态获取WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//localhost:8081/ws?token=${encodeURIComponent(token)}`;
    
    console.log('Connecting to WebSocket:', wsUrl);
    const newSocket = new WebSocket(wsUrl);

    newSocket.onopen = () => {
      console.log('WebSocket connected successfully');
      setConnected(true);
    };

    newSocket.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      setConnected(false);
      
      // 重连机制
      if (isAuthenticated) {
        console.log('Attempting to reconnect in 3 seconds...');
        setTimeout(() => {
          if (isAuthenticated && !socketRef.current) {
            initializeSocket();
          }
        }, 3000);
      }
    };

    newSocket.onerror = (error) => {
      console.error('WebSocket connection error:', error);
      setConnected(false);
    };

    // Listen for real-time updates
    newSocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        switch (data.Type) {
          case 'nodes_update':
            setRealtimeData(prev => ({
              ...prev,
              nodes: data.Data,
            }));
            break;
            
          case 'process_update':
            setRealtimeData(prev => ({
              ...prev,
              processes: {
                ...prev.processes,
                [data.Data.nodeName]: data.Data.processes,
              },
            }));
            break;
            
          case 'process_status_change':
            const { NodeName, ProcessName, Status, Timestamp } = data.Data;
            setRealtimeData(prev => ({
              ...prev,
              events: [
                {
                  id: Date.now(),
                  type: 'process_status_change',
                  nodeName: NodeName,
                  processName: ProcessName,
                  status: Status,
                  timestamp: Timestamp,
                },
                ...prev.events.slice(0, 99), // Keep last 100 events
              ],
            }));
            break;
            
          case 'system_stats':
            setRealtimeData(prev => ({
              ...prev,
              systemStats: data.Data,
            }));
            break;
            
          case 'activity_log':
            setRealtimeData(prev => ({
              ...prev,
              events: [
                {
                  id: Date.now(),
                  type: 'activity_log',
                  ...data.Data,
                },
                ...prev.events.slice(0, 99),
              ],
            }));
            break;
            
          default:
            console.log('Unknown message type:', data.Type);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    socketRef.current = newSocket;
    setSocket(newSocket);
  }, [isAuthenticated]);

  const disconnectSocket = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
      setSocket(null);
      setConnected(false);
    }
  }, []);

  useEffect(() => {
    if (isAuthenticated && !socketRef.current) {
      initializeSocket();
    } else if (!isAuthenticated && socketRef.current) {
      disconnectSocket();
    }

    return () => {
      disconnectSocket();
    };
  }, [isAuthenticated, initializeSocket, disconnectSocket]);

  const subscribeToNode = (nodeName) => {
    if (socket && connected && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        Type: 'subscribe_node',
        Data: nodeName
      }));
    }
  };

  const unsubscribeFromNode = (nodeName) => {
    if (socket && connected && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        Type: 'unsubscribe_node',
        Data: nodeName
      }));
    }
  };

  const requestNodeUpdate = (nodeName) => {
    if (socket && connected && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        Type: 'request_node_update',
        Data: nodeName
      }));
    }
  };

  const value = {
    socket,
    connected,
    realtimeData,
    subscribeToNode,
    unsubscribeFromNode,
    requestNodeUpdate,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
};