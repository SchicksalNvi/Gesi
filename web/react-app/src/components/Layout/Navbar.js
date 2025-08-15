import React from 'react';
import { Navbar as BootstrapNavbar, Nav, NavDropdown, Badge } from 'react-bootstrap';
import { useAuth } from '../../contexts/AuthContext';
import { useWebSocket } from '../../contexts/WebSocketContext';

const Navbar = () => {
  const { user, logout } = useAuth();
  const { connected } = useWebSocket();

  const handleLogout = () => {
    logout();
  };

  return (
    <BootstrapNavbar bg="white" expand="lg" className="border-bottom px-3">
      <BootstrapNavbar.Brand href="#" className="fw-bold text-primary">
        CESI
      </BootstrapNavbar.Brand>
      
      <BootstrapNavbar.Toggle aria-controls="basic-navbar-nav" />
      <BootstrapNavbar.Collapse id="basic-navbar-nav">
        <Nav className="ms-auto">
          <Nav.Item className="d-flex align-items-center me-3">
            {connected ? (
              <Badge bg="success">
                <i className="bi bi-wifi"></i> Connected
              </Badge>
            ) : (
              <Badge bg="danger">
                <i className="bi bi-wifi-off"></i> Disconnected
              </Badge>
            )}
          </Nav.Item>
          
          <NavDropdown 
            title={
              <span>
                <i className="bi bi-person-circle"></i> {user?.username}
              </span>
            } 
            id="user-dropdown"
          >
            <NavDropdown.Item href="/profile">
              <i className="bi bi-person"></i> Profile
            </NavDropdown.Item>
            <NavDropdown.Divider />
            <NavDropdown.Item onClick={handleLogout}>
              <i className="bi bi-box-arrow-right"></i> Logout
            </NavDropdown.Item>
          </NavDropdown>
        </Nav>
      </BootstrapNavbar.Collapse>
    </BootstrapNavbar>
  );
};

export default Navbar;