import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Table, Button, Badge, Alert } from 'react-bootstrap';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import moment from 'moment';
import { nodesAPI, activityLogsAPI } from '../services/api';
import { useWebSocket } from '../contexts/WebSocketContext';

const Dashboard = () => {
  const [nodes, setNodes] = useState([]);
  const [processes, setProcesses] = useState([]);
  const [stats, setStats] = useState({
    totalNodes: 0,
    runningProcesses: 0,
    stoppedProcesses: 0,
    totalProcesses: 0,
  });
  const [recentLogs, setRecentLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  const { realtimeData, connected } = useWebSocket();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  useEffect(() => {
    // Update data when real-time updates are received
    if (realtimeData.nodes.length > 0) {
      setNodes(realtimeData.nodes);
      calculateStats(realtimeData.nodes);
    }
  }, [realtimeData]);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const [nodesResponse, logsResponse] = await Promise.all([
        nodesAPI.getNodes(),
        activityLogsAPI.getRecentLogs(),
      ]);

      const nodesData = nodesResponse.data.nodes || [];
      setNodes(nodesData);
      setRecentLogs(logsResponse.data.logs || []);
      
      // Fetch processes for all nodes
      const allProcesses = [];
      for (const node of nodesData) {
        try {
          const processResponse = await nodesAPI.getNodeProcesses(node.name);
          const nodeProcesses = processResponse.data.processes || [];
          allProcesses.push(...nodeProcesses.map(p => ({ ...p, nodeName: node.name })));
        } catch (err) {
          console.error(`Failed to fetch processes for node ${node.name}:`, err);
        }
      }
      
      setProcesses(allProcesses);
      calculateStats(nodesData, allProcesses);
    } catch (err) {
      setError('Failed to fetch dashboard data');
      console.error('Dashboard fetch error:', err);
    } finally {
      setLoading(false);
    }
  };

  const calculateStats = (nodesData, processesData = processes) => {
    const totalNodes = nodesData.length;
    const connectedNodes = nodesData.filter(node => node.is_connected).length;
    const runningProcesses = processesData.filter(p => p.state === 'RUNNING').length;
    const stoppedProcesses = processesData.filter(p => p.state === 'STOPPED').length;
    const totalProcesses = processesData.length;

    setStats({
      totalNodes,
      connectedNodes,
      runningProcesses,
      stoppedProcesses,
      totalProcesses,
    });
  };

  const getStatusBadge = (state) => {
    switch (state) {
      case 'RUNNING':
        return <Badge bg="success">Running</Badge>;
      case 'STOPPED':
        return <Badge bg="danger">Stopped</Badge>;
      case 'STARTING':
        return <Badge bg="warning">Starting</Badge>;
      case 'STOPPING':
        return <Badge bg="warning">Stopping</Badge>;
      default:
        return <Badge bg="secondary">{state}</Badge>;
    }
  };

  const refreshData = () => {
    fetchDashboardData();
  };

  if (loading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: '400px' }}>
        <div className="spinner-border" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Dashboard</h1>
        <div>
          {connected && (
            <Badge bg="success" className="me-2">
              <i className="bi bi-wifi"></i> Real-time Connected
            </Badge>
          )}
          <Button variant="outline-primary" onClick={refreshData}>
            <i className="bi bi-arrow-clockwise"></i> Refresh
          </Button>
        </div>
      </div>

      {error && (
        <Alert variant="danger" dismissible onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Statistics Cards */}
      <Row className="mb-4">
        <Col md={3}>
          <Card className="text-center">
            <Card.Body>
              <h3 className="text-primary">{stats.totalNodes}</h3>
              <p className="mb-0">Total Nodes</p>
              <small className="text-muted">{stats.connectedNodes} connected</small>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="text-center">
            <Card.Body>
              <h3 className="text-success">{stats.runningProcesses}</h3>
              <p className="mb-0">Running Processes</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="text-center">
            <Card.Body>
              <h3 className="text-danger">{stats.stoppedProcesses}</h3>
              <p className="mb-0">Stopped Processes</p>
            </Card.Body>
          </Card>
        </Col>
        <Col md={3}>
          <Card className="text-center">
            <Card.Body>
              <h3 className="text-info">{stats.totalProcesses}</h3>
              <p className="mb-0">Total Processes</p>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row>
        {/* Nodes Overview */}
        <Col md={6}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">Nodes Overview</h5>
            </Card.Header>
            <Card.Body>
              <Table responsive>
                <thead>
                  <tr>
                    <th>Node</th>
                    <th>Environment</th>
                    <th>Status</th>
                    <th>Processes</th>
                  </tr>
                </thead>
                <tbody>
                  {nodes.map((node) => {
                    const nodeProcesses = processes.filter(p => p.nodeName === node.name);
                    const runningCount = nodeProcesses.filter(p => p.state === 'RUNNING').length;
                    return (
                      <tr key={node.name}>
                        <td>{node.name}</td>
                        <td>{node.environment}</td>
                        <td>
                          {node.is_connected ? (
                            <Badge bg="success">Connected</Badge>
                          ) : (
                            <Badge bg="danger">Disconnected</Badge>
                          )}
                        </td>
                        <td>
                          <span className="text-success">{runningCount}</span>/
                          <span className="text-muted">{nodeProcesses.length}</span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>

        {/* Recent Activity */}
        <Col md={6}>
          <Card>
            <Card.Header>
              <h5 className="mb-0">Recent Activity</h5>
            </Card.Header>
            <Card.Body style={{ maxHeight: '400px', overflowY: 'auto' }}>
              {recentLogs.length === 0 ? (
                <p className="text-muted">No recent activity</p>
              ) : (
                <div>
                  {recentLogs.map((log, index) => (
                    <div key={index} className="border-bottom pb-2 mb-2">
                      <div className="d-flex justify-content-between">
                        <small className="text-muted">
                          {moment(log.timestamp).format('HH:mm:ss')}
                        </small>
                        <Badge bg="info" size="sm">{log.action}</Badge>
                      </div>
                      <div>{log.message}</div>
                    </div>
                  ))}
                </div>
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Real-time Events */}
      {realtimeData.events.length > 0 && (
        <Row className="mt-4">
          <Col>
            <Card>
              <Card.Header>
                <h5 className="mb-0">
                  <i className="bi bi-broadcast"></i> Real-time Events
                </h5>
              </Card.Header>
              <Card.Body style={{ maxHeight: '300px', overflowY: 'auto' }}>
                {realtimeData.events.slice(0, 10).map((event) => (
                  <div key={event.id} className="border-bottom pb-2 mb-2">
                    <div className="d-flex justify-content-between">
                      <small className="text-muted">
                        {moment(event.timestamp).format('HH:mm:ss')}
                      </small>
                      <Badge bg="warning" size="sm">{event.type}</Badge>
                    </div>
                    <div>
                      {event.nodeName && <strong>{event.nodeName}</strong>}
                      {event.processName && <span> - {event.processName}</span>}
                      {event.status && <span> ({event.status})</span>}
                    </div>
                  </div>
                ))}
              </Card.Body>
            </Card>
          </Col>
        </Row>
      )}
    </div>
  );
};

export default Dashboard;