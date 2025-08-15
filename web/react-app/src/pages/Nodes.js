import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Table, Button, Badge, Alert, Modal } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { nodesAPI } from '../services/api';
import { useWebSocket } from '../contexts/WebSocketContext';
import moment from 'moment';

const Nodes = () => {
  const [nodes, setNodes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedNode, setSelectedNode] = useState(null);
  const [showBatchModal, setShowBatchModal] = useState(false);
  const [batchOperation, setBatchOperation] = useState('');
  const [operationLoading, setOperationLoading] = useState(false);
  
  const { realtimeData, subscribeToNode, unsubscribeFromNode } = useWebSocket();

  useEffect(() => {
    fetchNodes();
  }, []);

  useEffect(() => {
    // Update nodes when real-time data is received
    if (realtimeData.nodes.length > 0) {
      setNodes(realtimeData.nodes);
    }
  }, [realtimeData]);

  const fetchNodes = async () => {
    try {
      setLoading(true);
      const response = await nodesAPI.getNodes();
      setNodes(response.data.nodes || []);
    } catch (err) {
      setError('Failed to fetch nodes');
      console.error('Nodes fetch error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleBatchOperation = async (nodeName, operation) => {
    setOperationLoading(true);
    try {
      let response;
      switch (operation) {
        case 'start-all':
          response = await nodesAPI.startAllProcesses(nodeName);
          break;
        case 'stop-all':
          response = await nodesAPI.stopAllProcesses(nodeName);
          break;
        case 'restart-all':
          response = await nodesAPI.restartAllProcesses(nodeName);
          break;
        default:
          throw new Error('Invalid operation');
      }
      
      if (response.data.status === 'success') {
        // Refresh node data
        fetchNodes();
        setShowBatchModal(false);
      }
    } catch (err) {
      setError(`Failed to ${operation.replace('-', ' ')} processes`);
      console.error('Batch operation error:', err);
    } finally {
      setOperationLoading(false);
    }
  };

  const openBatchModal = (node, operation) => {
    setSelectedNode(node);
    setBatchOperation(operation);
    setShowBatchModal(true);
  };

  const getStatusBadge = (isConnected) => {
    return isConnected ? (
      <Badge bg="success">Connected</Badge>
    ) : (
      <Badge bg="danger">Disconnected</Badge>
    );
  };

  const getOperationLabel = (operation) => {
    switch (operation) {
      case 'start-all':
        return 'Start All Processes';
      case 'stop-all':
        return 'Stop All Processes';
      case 'restart-all':
        return 'Restart All Processes';
      default:
        return operation;
    }
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
        <h1>Nodes Management</h1>
        <Button variant="outline-primary" onClick={fetchNodes}>
          <i className="bi bi-arrow-clockwise"></i> Refresh
        </Button>
      </div>

      {error && (
        <Alert variant="danger" dismissible onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Row>
        {nodes.map((node) => (
          <Col md={6} lg={4} key={node.name} className="mb-4">
            <Card>
              <Card.Header className="d-flex justify-content-between align-items-center">
                <h5 className="mb-0">{node.name}</h5>
                {getStatusBadge(node.is_connected)}
              </Card.Header>
              <Card.Body>
                <div className="mb-3">
                  <strong>Environment:</strong> {node.environment}<br />
                  <strong>Host:</strong> {node.host}:{node.port}<br />
                  {node.last_ping && (
                    <>
                      <strong>Last Ping:</strong> {moment(node.last_ping).fromNow()}
                    </>
                  )}
                </div>
                
                <div className="d-grid gap-2">
                  <Link 
                    to={`/nodes/${node.name}`} 
                    className="btn btn-primary"
                  >
                    <i className="bi bi-eye"></i> View Details
                  </Link>
                  
                  {node.is_connected && (
                    <>
                      <Button 
                        variant="success" 
                        size="sm"
                        onClick={() => openBatchModal(node, 'start-all')}
                      >
                        <i className="bi bi-play-fill"></i> Start All
                      </Button>
                      <Button 
                        variant="warning" 
                        size="sm"
                        onClick={() => openBatchModal(node, 'restart-all')}
                      >
                        <i className="bi bi-arrow-clockwise"></i> Restart All
                      </Button>
                      <Button 
                        variant="danger" 
                        size="sm"
                        onClick={() => openBatchModal(node, 'stop-all')}
                      >
                        <i className="bi bi-stop-fill"></i> Stop All
                      </Button>
                    </>
                  )}
                </div>
              </Card.Body>
            </Card>
          </Col>
        ))}
      </Row>

      {nodes.length === 0 && !loading && (
        <Card>
          <Card.Body className="text-center py-5">
            <i className="bi bi-hdd-stack display-1 text-muted"></i>
            <h3 className="mt-3">No Nodes Found</h3>
            <p className="text-muted">No supervisor nodes are configured or available.</p>
          </Card.Body>
        </Card>
      )}

      {/* Batch Operation Confirmation Modal */}
      <Modal show={showBatchModal} onHide={() => setShowBatchModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Confirm Batch Operation</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p>
            Are you sure you want to <strong>{getOperationLabel(batchOperation)}</strong> on node <strong>{selectedNode?.name}</strong>?
          </p>
          <p className="text-muted">
            This operation will affect all processes on this node.
          </p>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowBatchModal(false)}>
            Cancel
          </Button>
          <Button 
            variant="primary" 
            onClick={() => handleBatchOperation(selectedNode?.name, batchOperation)}
            disabled={operationLoading}
          >
            {operationLoading ? (
              <>
                <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                Processing...
              </>
            ) : (
              'Confirm'
            )}
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default Nodes;