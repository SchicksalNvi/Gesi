import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Button, Table, Form, Modal, Badge, Alert, Accordion } from 'react-bootstrap';
import { toast } from 'sonner';
import { developerToolsAPI } from '../services/api';

const DeveloperTools = () => {
  const [activeTab, setActiveTab] = useState('api-docs');
  const [apiEndpoints, setApiEndpoints] = useState([]);
  const [debugLogs, setDebugLogs] = useState([]);
  const [performanceMetrics, setPerformanceMetrics] = useState({});
  const [loading, setLoading] = useState(false);
  const [showApiTestModal, setShowApiTestModal] = useState(false);
  const [selectedEndpoint, setSelectedEndpoint] = useState(null);
  const [apiTestResult, setApiTestResult] = useState(null);
  const [apiTestRequest, setApiTestRequest] = useState({
    method: 'GET',
    headers: {},
    body: ''
  });

  useEffect(() => {
    fetchApiEndpoints();
    fetchDebugLogs();
    fetchPerformanceMetrics();
  }, []);

  const fetchApiEndpoints = async () => {
    try {
      setLoading(true);
      const response = await developerToolsAPI.getApiEndpoints();
      setApiEndpoints(response.data || []);
    } catch (error) {
      console.error('Error fetching API endpoints:', error);
      toast.error('Failed to fetch API endpoints');
    } finally {
      setLoading(false);
    }
  };

  const fetchDebugLogs = async () => {
    try {
      const response = await developerToolsAPI.getDebugLogs();
      setDebugLogs(response.data || []);
    } catch (error) {
      console.error('Error fetching debug logs:', error);
      toast.error('Failed to fetch debug logs');
    }
  };

  const fetchPerformanceMetrics = async () => {
    try {
      const response = await developerToolsAPI.getPerformanceMetrics();
      setPerformanceMetrics(response.data || {});
    } catch (error) {
      console.error('Error fetching performance metrics:', error);
      toast.error('Failed to fetch performance metrics');
    }
  };

  const testApiEndpoint = async () => {
    if (!selectedEndpoint) return;

    try {
      setLoading(true);
      const testData = {
        method: apiTestRequest.method,
        path: selectedEndpoint.path,
        headers: apiTestRequest.headers,
        body: apiTestRequest.body
      };
      
      const response = await developerToolsAPI.testApiEndpoint(selectedEndpoint, testData);
      setApiTestResult(response.data);
    } catch (error) {
      console.error('Error testing API endpoint:', error);
      setApiTestResult({
        status: error.response?.status || 500,
        statusText: error.response?.statusText || 'Internal Server Error',
        error: error.message
      });
    } finally {
      setLoading(false);
    }
  };

  const getMethodBadge = (method) => {
    const variants = {
      'GET': 'success',
      'POST': 'primary',
      'PUT': 'warning',
      'DELETE': 'danger',
      'PATCH': 'info'
    };
    return <Badge bg={variants[method] || 'secondary'}>{method}</Badge>;
  };

  const getLevelBadge = (level) => {
    const variants = {
      'DEBUG': 'secondary',
      'INFO': 'info',
      'WARN': 'warning',
      'ERROR': 'danger'
    };
    return <Badge bg={variants[level] || 'secondary'}>{level}</Badge>;
  };

  const renderApiDocs = () => (
    <Card>
      <Card.Header>
        <Card.Title>API Documentation</Card.Title>
      </Card.Header>
      <Card.Body>
        <Accordion>
          {apiEndpoints.map((endpoint) => (
            <Accordion.Item key={endpoint.id} eventKey={endpoint.id.toString()}>
              <Accordion.Header>
                <div className="d-flex align-items-center gap-2">
                  {getMethodBadge(endpoint.method)}
                  <code>{endpoint.path}</code>
                  <span className="text-muted">- {endpoint.description}</span>
                </div>
              </Accordion.Header>
              <Accordion.Body>
                <Row>
                  <Col md={6}>
                    <h6>Parameters</h6>
                    {endpoint.parameters.length > 0 ? (
                      <Table size="sm" striped>
                        <thead>
                          <tr>
                            <th>Name</th>
                            <th>Type</th>
                            <th>Required</th>
                            <th>Description</th>
                          </tr>
                        </thead>
                        <tbody>
                          {endpoint.parameters.map((param, index) => (
                            <tr key={index}>
                              <td><code>{param.name}</code></td>
                              <td>{param.type}</td>
                              <td>{param.required ? 'Yes' : 'No'}</td>
                              <td>{param.description}</td>
                            </tr>
                          ))}
                        </tbody>
                      </Table>
                    ) : (
                      <p className="text-muted">No parameters</p>
                    )}

                    <h6 className="mt-3">Responses</h6>
                    <Table size="sm" striped>
                      <thead>
                        <tr>
                          <th>Status</th>
                          <th>Description</th>
                        </tr>
                      </thead>
                      <tbody>
                        {Object.entries(endpoint.responses).map(([status, description]) => (
                          <tr key={status}>
                            <td><Badge bg={status.startsWith('2') ? 'success' : status.startsWith('4') ? 'warning' : 'danger'}>{status}</Badge></td>
                            <td>{description}</td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  </Col>
                  <Col md={6}>
                    <h6>Example</h6>
                    <div className="mb-2">
                      <strong>Request:</strong>
                      <pre className="bg-light p-2 rounded">{endpoint.example.request}</pre>
                    </div>
                    <div className="mb-2">
                      <strong>Response:</strong>
                      <pre className="bg-light p-2 rounded">{endpoint.example.response}</pre>
                    </div>
                    <Button
                      variant="outline-primary"
                      size="sm"
                      onClick={() => {
                        setSelectedEndpoint(endpoint);
                        setShowApiTestModal(true);
                      }}
                    >
                      Test API
                    </Button>
                  </Col>
                </Row>
              </Accordion.Body>
            </Accordion.Item>
          ))}
        </Accordion>
      </Card.Body>
    </Card>
  );

  const renderDebugTools = () => (
    <Card>
      <Card.Header className="d-flex justify-content-between align-items-center">
        <Card.Title>Debug Logs</Card.Title>
        <Button variant="outline-primary" onClick={fetchDebugLogs}>
          Refresh
        </Button>
      </Card.Header>
      <Card.Body>
        <Table striped bordered hover>
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Level</th>
              <th>Component</th>
              <th>Message</th>
              <th>Details</th>
            </tr>
          </thead>
          <tbody>
            {debugLogs.map((log) => (
              <tr key={log.id}>
                <td>{log.timestamp}</td>
                <td>{getLevelBadge(log.level)}</td>
                <td>{log.component}</td>
                <td>{log.message}</td>
                <td><small className="text-muted">{log.details}</small></td>
              </tr>
            ))}
          </tbody>
        </Table>
      </Card.Body>
    </Card>
  );

  const renderPerformanceMonitoring = () => (
    <div>
      <Row>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <Card.Title>System Metrics</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table borderless>
                <tbody>
                  <tr>
                    <td><strong>Uptime:</strong></td>
                    <td>{performanceMetrics.system?.uptime}</td>
                  </tr>
                  <tr>
                    <td><strong>Memory Usage:</strong></td>
                    <td>{performanceMetrics.system?.memoryUsage}</td>
                  </tr>
                  <tr>
                    <td><strong>CPU Usage:</strong></td>
                    <td>{performanceMetrics.system?.cpuUsage}</td>
                  </tr>
                  <tr>
                    <td><strong>Disk Usage:</strong></td>
                    <td>{performanceMetrics.system?.diskUsage}</td>
                  </tr>
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <Card.Title>API Metrics</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table borderless>
                <tbody>
                  <tr>
                    <td><strong>Total Requests:</strong></td>
                    <td>{performanceMetrics.api?.totalRequests}</td>
                  </tr>
                  <tr>
                    <td><strong>Requests/Min:</strong></td>
                    <td>{performanceMetrics.api?.requestsPerMinute}</td>
                  </tr>
                  <tr>
                    <td><strong>Avg Response Time:</strong></td>
                    <td>{performanceMetrics.api?.averageResponseTime}</td>
                  </tr>
                  <tr>
                    <td><strong>Error Rate:</strong></td>
                    <td>{performanceMetrics.api?.errorRate}</td>
                  </tr>
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
      </Row>
      <Row>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <Card.Title>Database Metrics</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table borderless>
                <tbody>
                  <tr>
                    <td><strong>Active Connections:</strong></td>
                    <td>{performanceMetrics.database?.connections}/{performanceMetrics.database?.maxConnections}</td>
                  </tr>
                  <tr>
                    <td><strong>Avg Query Time:</strong></td>
                    <td>{performanceMetrics.database?.queryTime}</td>
                  </tr>
                  <tr>
                    <td><strong>Slow Queries:</strong></td>
                    <td>{performanceMetrics.database?.slowQueries}</td>
                  </tr>
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
        <Col md={6}>
          <Card className="mb-4">
            <Card.Header>
              <Card.Title>WebSocket Metrics</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table borderless>
                <tbody>
                  <tr>
                    <td><strong>Active Connections:</strong></td>
                    <td>{performanceMetrics.websocket?.activeConnections}</td>
                  </tr>
                  <tr>
                    <td><strong>Messages/Sec:</strong></td>
                    <td>{performanceMetrics.websocket?.messagesPerSecond}</td>
                  </tr>
                  <tr>
                    <td><strong>Total Messages:</strong></td>
                    <td>{performanceMetrics.websocket?.totalMessages}</td>
                  </tr>
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </div>
  );

  const renderTabContent = () => {
    switch (activeTab) {
      case 'api-docs': return renderApiDocs();
      case 'debug': return renderDebugTools();
      case 'performance': return renderPerformanceMonitoring();
      default: return null;
    }
  };

  return (
    <Container>
      <Row>
        <Col>
          <h2>Developer Tools</h2>
          
          {/* Tab Navigation */}
          <Card className="mb-4">
            <Card.Body>
              <div className="d-flex gap-2">
                <Button
                  variant={activeTab === 'api-docs' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('api-docs')}
                >
                  API Documentation
                </Button>
                <Button
                  variant={activeTab === 'debug' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('debug')}
                >
                  Debug Tools
                </Button>
                <Button
                  variant={activeTab === 'performance' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('performance')}
                >
                  Performance Monitoring
                </Button>
              </div>
            </Card.Body>
          </Card>

          {/* Tab Content */}
          {renderTabContent()}

          {/* API Test Modal */}
          <Modal show={showApiTestModal} onHide={() => setShowApiTestModal(false)} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Test API Endpoint</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              {selectedEndpoint && (
                <div>
                  <div className="mb-3">
                    <strong>Endpoint:</strong> {getMethodBadge(selectedEndpoint.method)} <code>{selectedEndpoint.path}</code>
                  </div>
                  
                  <Form>
                    <Form.Group className="mb-3">
                      <Form.Label>Request Headers (JSON)</Form.Label>
                      <Form.Control
                        as="textarea"
                        rows={3}
                        placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
                        value={JSON.stringify(apiTestRequest.headers, null, 2)}
                        onChange={(e) => {
                          try {
                            const headers = JSON.parse(e.target.value);
                            setApiTestRequest({...apiTestRequest, headers});
                          } catch (err) {
                            // Invalid JSON, keep the text for user to fix
                          }
                        }}
                      />
                    </Form.Group>
                    
                    {selectedEndpoint.method !== 'GET' && (
                      <Form.Group className="mb-3">
                        <Form.Label>Request Body</Form.Label>
                        <Form.Control
                          as="textarea"
                          rows={4}
                          placeholder='Request body (JSON)'
                          value={apiTestRequest.body}
                          onChange={(e) => setApiTestRequest({...apiTestRequest, body: e.target.value})}
                        />
                      </Form.Group>
                    )}
                  </Form>
                  
                  {apiTestResult && (
                    <div className="mt-3">
                      <h6>Response</h6>
                      <Alert variant={apiTestResult.status < 400 ? 'success' : 'danger'}>
                        <strong>Status:</strong> {apiTestResult.status} {apiTestResult.statusText}
                        {apiTestResult.responseTime && (
                          <span className="ms-3"><strong>Response Time:</strong> {apiTestResult.responseTime}ms</span>
                        )}
                      </Alert>
                      
                      {apiTestResult.headers && (
                        <div className="mb-2">
                          <strong>Headers:</strong>
                          <pre className="bg-light p-2 rounded">{JSON.stringify(apiTestResult.headers, null, 2)}</pre>
                        </div>
                      )}
                      
                      {apiTestResult.data && (
                        <div>
                          <strong>Response Body:</strong>
                          <pre className="bg-light p-2 rounded">{apiTestResult.data}</pre>
                        </div>
                      )}
                      
                      {apiTestResult.error && (
                        <div>
                          <strong>Error:</strong>
                          <pre className="bg-danger text-white p-2 rounded">{apiTestResult.error}</pre>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={() => setShowApiTestModal(false)}>
                Close
              </Button>
              <Button variant="primary" onClick={testApiEndpoint} disabled={loading}>
                {loading ? 'Testing...' : 'Send Request'}
              </Button>
            </Modal.Footer>
          </Modal>
        </Col>
      </Row>
    </Container>
  );
};

export default DeveloperTools;