import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Badge, Table, Button } from 'react-bootstrap';
import { toast } from 'sonner';

const Groups = () => {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchGroups();
  }, []);

  const fetchGroups = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/groups', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        setGroups(data.groups || []);
      } else {
        throw new Error('Failed to fetch groups');
      }
    } catch (error) {
      toast.error('Failed to fetch groups');
      console.error('Error fetching groups:', error);
      setGroups([]);
    } finally {
      setLoading(false);
    }
  };

  const handleGroupAction = async (groupName, action) => {
    try {
      toast.success(`${action} action triggered for group: ${groupName}`);
      // Refresh data
      fetchGroups();
    } catch (error) {
      toast.error(`Failed to ${action} group: ${groupName}`);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Container>
      <Row>
        <Col>
          <h2>Process Groups</h2>
          <Card>
            <Card.Header>
              <Card.Title>Group List</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Processes</th>
                    <th>Node</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {groups.map((group) => (
                    <tr key={group.name}>
                      <td>{group.name}</td>
                      <td>{group.description}</td>
                      <td>{group.processes.join(', ')}</td>
                      <td>{group.node}</td>
                      <td>
                        <Badge bg={group.status === 'running' ? 'success' : 'danger'}>
                          {group.status}
                        </Badge>
                      </td>
                      <td>
                        <Button 
                          size="sm" 
                          variant="outline-success" 
                          className="me-2"
                          onClick={() => handleGroupAction(group.name, 'start')}
                        >
                          Start
                        </Button>
                        <Button 
                          size="sm" 
                          variant="outline-warning" 
                          className="me-2"
                          onClick={() => handleGroupAction(group.name, 'stop')}
                        >
                          Stop
                        </Button>
                        <Button 
                          size="sm" 
                          variant="outline-info"
                          onClick={() => handleGroupAction(group.name, 'restart')}
                        >
                          Restart
                        </Button>
                      </td>
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

export default Groups;