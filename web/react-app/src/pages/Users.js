import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Badge, Table, Button, Modal, Form } from 'react-bootstrap';
import { toast } from 'sonner';

const Users = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newUser, setNewUser] = useState({ username: '', email: '', isAdmin: false });

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/users', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        setUsers(data);
      } else {
        throw new Error('Failed to fetch users');
      }
    } catch (error) {
      toast.error('Failed to fetch users');
      console.error('Error fetching users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async () => {
    try {
      const response = await fetch('/api/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(newUser)
      });
      if (response.ok) {
        const user = await response.json();
        setUsers([...users, user]);
        setShowCreateModal(false);
        setNewUser({ username: '', email: '', isAdmin: false });
        toast.success('User created successfully');
      } else {
        throw new Error('Failed to create user');
      }
    } catch (error) {
      toast.error('Failed to create user');
      console.error('Error creating user:', error);
    }
  };

  const handleDeleteUser = async (userId) => {
    try {
      setUsers(users.filter(user => user.id !== userId));
      toast.success('User deleted successfully');
    } catch (error) {
      toast.error('Failed to delete user');
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Container>
      <Row>
        <Col>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>User Management</h2>
            <Button variant="primary" onClick={() => setShowCreateModal(true)}>
              Create User
            </Button>
          </div>
          <Card>
            <Card.Header>
              <Card.Title>User List</Card.Title>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Username</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Last Login</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id}>
                      <td>{user.username}</td>
                      <td>{user.email}</td>
                      <td>
                        <Badge bg={user.isAdmin ? 'danger' : 'info'}>
                          {user.isAdmin ? 'Admin' : 'User'}
                        </Badge>
                      </td>
                      <td>
                        <Badge bg={user.status === 'active' ? 'success' : 'secondary'}>
                          {user.status}
                        </Badge>
                      </td>
                      <td>{user.lastLogin}</td>
                      <td>
                        <Button 
                          size="sm" 
                          variant="outline-primary" 
                          className="me-2"
                        >
                          Edit
                        </Button>
                        <Button 
                          size="sm" 
                          variant="outline-danger"
                          onClick={() => handleDeleteUser(user.id)}
                        >
                          Delete
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

      <Modal show={showCreateModal} onHide={() => setShowCreateModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Create New User</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>Username</Form.Label>
              <Form.Control
                type="text"
                value={newUser.username}
                onChange={(e) => setNewUser({...newUser, username: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Email</Form.Label>
              <Form.Control
                type="email"
                value={newUser.email}
                onChange={(e) => setNewUser({...newUser, email: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Check
                type="checkbox"
                label="Admin User"
                checked={newUser.isAdmin}
                onChange={(e) => setNewUser({...newUser, isAdmin: e.target.checked})}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowCreateModal(false)}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleCreateUser}>
            Create User
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
};

export default Users;