import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Button, Form, Table, Modal, Badge, Alert } from 'react-bootstrap';
import { toast } from 'sonner';
import { systemSettingsAPI } from '../services/api';

const SystemSettings = () => {
  const [activeTab, setActiveTab] = useState('theme');
  const [settings, setSettings] = useState({
    theme: {
      primaryColor: '#007bff',
      secondaryColor: '#6c757d',
      darkMode: false,
      compactMode: false,
      sidebarCollapsed: false
    },
    language: {
      current: 'en',
      available: ['en', 'zh', 'es', 'fr', 'de']
    },
    system: {
      sessionTimeout: 30,
      autoRefresh: true,
      refreshInterval: 5,
      maxLogEntries: 1000,
      enableNotifications: true,
      enableAuditLog: true
    },
    email: {
      smtpHost: '',
      smtpPort: 587,
      smtpUser: '',
      smtpPassword: '',
      enableTLS: true,
      fromAddress: '',
      testEmail: ''
    }
  });
  const [loading, setLoading] = useState(false);
  const [showEmailTestModal, setShowEmailTestModal] = useState(false);
  const [emailTestResult, setEmailTestResult] = useState(null);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const response = await systemSettingsAPI.getSystemSettings();
      setSettings(response.data);
    } catch (error) {
      toast.error('Failed to fetch system settings');
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async (category) => {
    try {
      setLoading(true);
      await systemSettingsAPI.updateMultipleSettings({ [category]: settings[category] });
      toast.success(`${category} settings saved successfully`);
    } catch (error) {
      toast.error(`Failed to save ${category} settings`);
    } finally {
      setLoading(false);
    }
  };

  const testEmailConfiguration = async () => {
    try {
      setLoading(true);
      setEmailTestResult(null);
      
      const response = await systemSettingsAPI.testEmailConfiguration({
        testEmail: settings.email.testEmail,
        smtpConfig: settings.email
      });
      
      setEmailTestResult({
        success: true,
        message: response.data.message || 'Test email sent successfully!'
      });
    } catch (error) {
      setEmailTestResult({
        success: false,
        message: error.response?.data?.message || 'Failed to send test email. Please check your SMTP configuration.'
      });
    } finally {
      setLoading(false);
    }
  };

  const updateSetting = (category, key, value) => {
    setSettings(prev => ({
      ...prev,
      [category]: {
        ...prev[category],
        [key]: value
      }
    }));
  };

  const renderThemeSettings = () => (
    <Card>
      <Card.Header>
        <Card.Title>Theme & Appearance</Card.Title>
      </Card.Header>
      <Card.Body>
        <Form>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Primary Color</Form.Label>
                <Form.Control
                  type="color"
                  value={settings.theme.primaryColor}
                  onChange={(e) => updateSetting('theme', 'primaryColor', e.target.value)}
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Secondary Color</Form.Label>
                <Form.Control
                  type="color"
                  value={settings.theme.secondaryColor}
                  onChange={(e) => updateSetting('theme', 'secondaryColor', e.target.value)}
                />
              </Form.Group>
            </Col>
          </Row>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Enable Dark Mode"
              checked={settings.theme.darkMode}
              onChange={(e) => updateSetting('theme', 'darkMode', e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Compact Mode"
              checked={settings.theme.compactMode}
              onChange={(e) => updateSetting('theme', 'compactMode', e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Collapse Sidebar by Default"
              checked={settings.theme.sidebarCollapsed}
              onChange={(e) => updateSetting('theme', 'sidebarCollapsed', e.target.checked)}
            />
          </Form.Group>
          
          <Button variant="primary" onClick={() => saveSettings('theme')} disabled={loading}>
            Save Theme Settings
          </Button>
        </Form>
      </Card.Body>
    </Card>
  );

  const renderLanguageSettings = () => (
    <Card>
      <Card.Header>
        <Card.Title>Language & Localization</Card.Title>
      </Card.Header>
      <Card.Body>
        <Form>
          <Form.Group className="mb-3">
            <Form.Label>Current Language</Form.Label>
            <Form.Select
              value={settings.language.current}
              onChange={(e) => updateSetting('language', 'current', e.target.value)}
            >
              <option value="en">English</option>
              <option value="zh">中文 (Chinese)</option>
              <option value="es">Español (Spanish)</option>
              <option value="fr">Français (French)</option>
              <option value="de">Deutsch (German)</option>
            </Form.Select>
          </Form.Group>
          
          <Alert variant="info">
            <strong>Note:</strong> Language changes will take effect after page refresh.
          </Alert>
          
          <Button variant="primary" onClick={() => saveSettings('language')} disabled={loading}>
            Save Language Settings
          </Button>
        </Form>
      </Card.Body>
    </Card>
  );

  const renderSystemSettings = () => (
    <Card>
      <Card.Header>
        <Card.Title>System Parameters</Card.Title>
      </Card.Header>
      <Card.Body>
        <Form>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Session Timeout (minutes)</Form.Label>
                <Form.Control
                  type="number"
                  min="5"
                  max="480"
                  value={settings.system.sessionTimeout}
                  onChange={(e) => updateSetting('system', 'sessionTimeout', parseInt(e.target.value))}
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Auto Refresh Interval (seconds)</Form.Label>
                <Form.Control
                  type="number"
                  min="1"
                  max="300"
                  value={settings.system.refreshInterval}
                  onChange={(e) => updateSetting('system', 'refreshInterval', parseInt(e.target.value))}
                  disabled={!settings.system.autoRefresh}
                />
              </Form.Group>
            </Col>
          </Row>
          
          <Form.Group className="mb-3">
            <Form.Label>Maximum Log Entries</Form.Label>
            <Form.Control
              type="number"
              min="100"
              max="10000"
              value={settings.system.maxLogEntries}
              onChange={(e) => updateSetting('system', 'maxLogEntries', parseInt(e.target.value))}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Enable Auto Refresh"
              checked={settings.system.autoRefresh}
              onChange={(e) => updateSetting('system', 'autoRefresh', e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Enable Notifications"
              checked={settings.system.enableNotifications}
              onChange={(e) => updateSetting('system', 'enableNotifications', e.target.checked)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Enable Audit Logging"
              checked={settings.system.enableAuditLog}
              onChange={(e) => updateSetting('system', 'enableAuditLog', e.target.checked)}
            />
          </Form.Group>
          
          <Button variant="primary" onClick={() => saveSettings('system')} disabled={loading}>
            Save System Settings
          </Button>
        </Form>
      </Card.Body>
    </Card>
  );

  const renderEmailSettings = () => (
    <Card>
      <Card.Header>
        <Card.Title>Email Configuration</Card.Title>
      </Card.Header>
      <Card.Body>
        <Form>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>SMTP Host</Form.Label>
                <Form.Control
                  type="text"
                  placeholder="smtp.example.com"
                  value={settings.email.smtpHost}
                  onChange={(e) => updateSetting('email', 'smtpHost', e.target.value)}
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>SMTP Port</Form.Label>
                <Form.Control
                  type="number"
                  value={settings.email.smtpPort}
                  onChange={(e) => updateSetting('email', 'smtpPort', parseInt(e.target.value))}
                />
              </Form.Group>
            </Col>
          </Row>
          
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>SMTP Username</Form.Label>
                <Form.Control
                  type="text"
                  value={settings.email.smtpUser}
                  onChange={(e) => updateSetting('email', 'smtpUser', e.target.value)}
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>SMTP Password</Form.Label>
                <Form.Control
                  type="password"
                  value={settings.email.smtpPassword}
                  onChange={(e) => updateSetting('email', 'smtpPassword', e.target.value)}
                />
              </Form.Group>
            </Col>
          </Row>
          
          <Form.Group className="mb-3">
            <Form.Label>From Address</Form.Label>
            <Form.Control
              type="email"
              placeholder="noreply@example.com"
              value={settings.email.fromAddress}
              onChange={(e) => updateSetting('email', 'fromAddress', e.target.value)}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Check
              type="checkbox"
              label="Enable TLS/SSL"
              checked={settings.email.enableTLS}
              onChange={(e) => updateSetting('email', 'enableTLS', e.target.checked)}
            />
          </Form.Group>
          
          <div className="d-flex gap-2">
            <Button variant="primary" onClick={() => saveSettings('email')} disabled={loading}>
              Save Email Settings
            </Button>
            <Button variant="outline-primary" onClick={() => setShowEmailTestModal(true)}>
              Test Configuration
            </Button>
          </div>
        </Form>
      </Card.Body>
    </Card>
  );

  const renderTabContent = () => {
    switch (activeTab) {
      case 'theme': return renderThemeSettings();
      case 'language': return renderLanguageSettings();
      case 'system': return renderSystemSettings();
      case 'email': return renderEmailSettings();
      default: return null;
    }
  };

  return (
    <Container>
      <Row>
        <Col>
          <h2>System Settings</h2>
          
          {/* Tab Navigation */}
          <Card className="mb-4">
            <Card.Body>
              <div className="d-flex gap-2">
                <Button
                  variant={activeTab === 'theme' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('theme')}
                >
                  Theme & Appearance
                </Button>
                <Button
                  variant={activeTab === 'language' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('language')}
                >
                  Language
                </Button>
                <Button
                  variant={activeTab === 'system' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('system')}
                >
                  System Parameters
                </Button>
                <Button
                  variant={activeTab === 'email' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('email')}
                >
                  Email Configuration
                </Button>
              </div>
            </Card.Body>
          </Card>

          {/* Tab Content */}
          {renderTabContent()}

          {/* Email Test Modal */}
          <Modal show={showEmailTestModal} onHide={() => setShowEmailTestModal(false)}>
            <Modal.Header closeButton>
              <Modal.Title>Test Email Configuration</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              <Form>
                <Form.Group className="mb-3">
                  <Form.Label>Test Email Address</Form.Label>
                  <Form.Control
                    type="email"
                    placeholder="test@example.com"
                    value={settings.email.testEmail}
                    onChange={(e) => updateSetting('email', 'testEmail', e.target.value)}
                  />
                </Form.Group>
                
                {emailTestResult && (
                  <Alert variant={emailTestResult.success ? 'success' : 'danger'}>
                    {emailTestResult.message}
                  </Alert>
                )}
              </Form>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={() => setShowEmailTestModal(false)}>
                Cancel
              </Button>
              <Button 
                variant="primary" 
                onClick={testEmailConfiguration}
                disabled={loading || !settings.email.testEmail}
              >
                {loading ? 'Sending...' : 'Send Test Email'}
              </Button>
            </Modal.Footer>
          </Modal>
        </Col>
      </Row>
    </Container>
  );
};

export default SystemSettings;