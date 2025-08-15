import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Badge, Table, Form, InputGroup } from 'react-bootstrap';
import { toast } from 'sonner';
import { activityLogsAPI } from '../services/api';

const ActivityLogs = () => {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterLevel, setFilterLevel] = useState('all');

  useEffect(() => {
    fetchLogs();
  }, []);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      const response = await activityLogsAPI.getActivityLogs();
      // 后端返回的数据结构是 {status: 'success', data: {logs: [...], pagination: {...}}}
      const logsData = response.data?.data?.logs || [];
      setLogs(Array.isArray(logsData) ? logsData : []);
    } catch (error) {
      console.error('Failed to fetch activity logs:', error);
      toast.error('Failed to fetch activity logs');
      setLogs([]);
    } finally {
      setLoading(false);
    }
  };

  const filteredLogs = Array.isArray(logs) ? logs.filter(log => {
    // 确保log对象存在且包含必要的字段
    if (!log || typeof log !== 'object') return false;
    
    const message = log.message || '';
    const action = log.action || '';
    const user = log.user || '';
    const target = log.target || '';
    const level = log.level || '';
    
    const matchesSearch = message.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         action.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         user.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         target.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesLevel = filterLevel === 'all' || level === filterLevel;
    return matchesSearch && matchesLevel;
  }) : [];

  const getLevelBadgeVariant = (level) => {
    switch (level) {
      case 'error': return 'danger';
      case 'warning': return 'warning';
      case 'info': return 'info';
      case 'debug': return 'secondary';
      default: return 'primary';
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Container>
      <Row>
        <Col>
          <h2>Activity Logs</h2>
          
          <Card className="mb-4">
            <Card.Body>
              <Row>
                <Col md={8}>
                  <InputGroup>
                    <Form.Control
                      type="text"
                      placeholder="Search logs..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                    />
                  </InputGroup>
                </Col>
                <Col md={4}>
                  <Form.Select
                    value={filterLevel}
                    onChange={(e) => setFilterLevel(e.target.value)}
                  >
                    <option value="all">All Levels</option>
                    <option value="error">Error</option>
                    <option value="warning">Warning</option>
                    <option value="info">Info</option>
                    <option value="debug">Debug</option>
                  </Form.Select>
                </Col>
              </Row>
            </Card.Body>
          </Card>

          <Card>
            <Card.Header>
              <Card.Title>Activity Log Entries ({filteredLogs.length})</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Level</th>
                    <th>Action</th>
                    <th>User</th>
                    <th>Target</th>
                    <th>Node</th>
                    <th>Message</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredLogs.map((log) => (
                    <tr key={log.id}>
                      <td>{log.timestamp}</td>
                      <td>
                        <Badge bg={getLevelBadgeVariant(log.level)}>
                          {log.level.toUpperCase()}
                        </Badge>
                      </td>
                      <td>{log.action}</td>
                      <td>{log.user}</td>
                      <td>{log.target}</td>
                      <td>{log.node}</td>
                      <td>{log.message}</td>
                    </tr>
                  ))}
                </tbody>
              </Table>
              {filteredLogs.length === 0 && (
                <div className="text-center py-4">
                  <p className="text-muted">No logs found matching your criteria.</p>
                </div>
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default ActivityLogs;