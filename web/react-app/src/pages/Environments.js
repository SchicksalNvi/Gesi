import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Badge, Table } from 'react-bootstrap';
import { toast } from 'sonner';

const Environments = () => {
  const [environments, setEnvironments] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchEnvironments();
  }, []);

  const fetchEnvironments = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/environments', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        setEnvironments(data.environments || []);
      } else {
        throw new Error('Failed to fetch environments');
      }
    } catch (error) {
      toast.error('Failed to fetch environments');
      console.error('Error fetching environments:', error);
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
          <h2>Environments</h2>
          <Card>
            <Card.Header>
              <Card.Title>Environment List</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Nodes</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {environments.map((env) => (
                    <tr key={env.name}>
                      <td>{env.name}</td>
                      <td>{env.description || 'No description'}</td>
                      <td>{env.members ? env.members.map(member => member.name).join(', ') : 'No nodes'}</td>
                      <td><Badge bg="success">{env.status || 'Active'}</Badge></td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default Environments;