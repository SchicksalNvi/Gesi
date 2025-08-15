import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Container, Row, Col, Card, Button, Badge, Table } from 'react-bootstrap';
import { toast } from 'sonner';

const NodeDetail = () => {
  const { nodeName } = useParams();
  const [node, setNode] = useState(null);
  const [processes, setProcesses] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchNodeData();
  }, [nodeName]);

  const fetchNodeData = async () => {
    try {
      setLoading(true);
      
      // 获取节点信息
      const nodeResponse = await fetch(`/api/nodes/${nodeName}`, {
        credentials: 'include'
      });
      if (nodeResponse.ok) {
        const nodeData = await nodeResponse.json();
        setNode(nodeData);
      } else {
        console.warn(`Failed to fetch node info for ${nodeName}`);
        setNode({ name: nodeName, status: 'Unknown', version: 'N/A', uptime: 'N/A' });
      }
      
      // 获取节点上的进程
      try {
        const processesResponse = await fetch(`/api/nodes/${nodeName}/processes`, {
          credentials: 'include'
        });
        if (processesResponse.ok) {
          const processesData = await processesResponse.json();
          setProcesses(processesData.processes || []);
        } else {
          console.warn(`Failed to fetch processes for node ${nodeName}`);
          setProcesses([]);
          // 不显示错误toast，只在控制台记录
        }
      } catch (processError) {
        console.warn('Error fetching processes:', processError);
        setProcesses([]);
      }
    } catch (error) {
      console.error('Error fetching node data:', error);
      setNode({ name: nodeName, status: 'Error', version: 'N/A', uptime: 'N/A' });
      setProcesses([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Container>
      <Row>
        <Col>
          <h2>Node: {node?.name}</h2>
          <Card>
            <Card.Body>
              <p>Status: <Badge bg="success">{node?.status}</Badge></p>
              <p>Version: {node?.version}</p>
              <p>Uptime: {node?.uptime}</p>
            </Card.Body>
          </Card>
        </Col>
      </Row>
      <Row className="mt-4">
        <Col>
          <Card>
            <Card.Header>
              <Card.Title>Processes</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Status</th>
                    <th>PID</th>
                    <th>Uptime</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {processes && processes.length > 0 ? (
                    processes.map((process) => (
                      <tr key={process.name}>
                        <td>{process.name}</td>
                        <td><Badge bg="success">{process.status}</Badge></td>
                        <td>{process.pid}</td>
                        <td>{process.uptime}</td>
                        <td>
                          <Button size="sm" variant="outline-primary" className="me-2">Start</Button>
                          <Button size="sm" variant="outline-warning" className="me-2">Stop</Button>
                          <Button size="sm" variant="outline-info">Restart</Button>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan="5" className="text-center text-muted">
                        No processes found or unable to connect to supervisord
                      </td>
                    </tr>
                  )}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default NodeDetail;