import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Card, Button, Table, Form, InputGroup, Modal, ProgressBar, Badge } from 'react-bootstrap';
import { toast } from 'sonner';
import { dataManagementAPI } from '../services/api';

const DataManagement = () => {
  const [activeTab, setActiveTab] = useState('export');
  const [exportRecords, setExportRecords] = useState([]);
  const [backupRecords, setBackupRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showExportModal, setShowExportModal] = useState(false);
  const [showBackupModal, setShowBackupModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const [backupProgress, setBackupProgress] = useState(0);
  const [importProgress, setImportProgress] = useState(0);
  const [exportConfig, setExportConfig] = useState({
    format: 'json',
    tables: [],
    dateRange: 'all'
  });
  const [backupConfig, setBackupConfig] = useState({
    includeFiles: true,
    compression: true,
    description: ''
  });
  const [importFile, setImportFile] = useState(null);

  useEffect(() => {
    fetchExportRecords();
    fetchBackupRecords();
  }, []);

  const fetchExportRecords = async () => {
    try {
      setLoading(true);
      const response = await dataManagementAPI.getExportRecords();
      setExportRecords(response.data);
    } catch (error) {
      toast.error('Failed to fetch export records');
    } finally {
      setLoading(false);
    }
  };

  const fetchBackupRecords = async () => {
    try {
      setLoading(true);
      const response = await dataManagementAPI.getBackupRecords();
      setBackupRecords(response.data);
    } catch (error) {
      toast.error('Failed to fetch backup records');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = async () => {
    try {
      setExportProgress(0);
      const response = await dataManagementAPI.exportData(exportConfig);
       
       // Simulate progress for now (until backend implements progress tracking)
       const interval = setInterval(() => {
         setExportProgress(prev => {
           if (prev >= 100) {
             clearInterval(interval);
             setShowExportModal(false);
             toast.success('Data export completed successfully');
             fetchExportRecords();
             return 100;
           }
           return prev + 10;
         });
       }, 500);
    } catch (error) {
      toast.error('Failed to start data export');
    }
  };

  const handleBackup = async () => {
    try {
      setBackupProgress(0);
      const response = await dataManagementAPI.createBackup(backupConfig);
       
       // Simulate progress for now (until backend implements progress tracking)
       const interval = setInterval(() => {
         setBackupProgress(prev => {
           if (prev >= 100) {
             clearInterval(interval);
             setShowBackupModal(false);
             toast.success('System backup completed successfully');
             fetchBackupRecords();
             return 100;
           }
           return prev + 8;
         });
       }, 600);
    } catch (error) {
      toast.error('Failed to start system backup');
    }
  };

  const handleImport = async () => {
    if (!importFile) {
      toast.error('Please select a file to import');
      return;
    }

    try {
      setImportProgress(0);
      const formData = new FormData();
      formData.append('file', importFile);
      
      const response = await dataManagementAPI.importData(formData);
       
       // Simulate progress for now (until backend implements progress tracking)
       const interval = setInterval(() => {
         setImportProgress(prev => {
           if (prev >= 100) {
             clearInterval(interval);
             setShowImportModal(false);
             setImportFile(null);
             toast.success('Data import completed successfully');
             return 100;
           }
           return prev + 12;
         });
       }, 400);
    } catch (error) {
      toast.error('Failed to import data');
    }
  };

  const handleDownload = (url, filename) => {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    toast.success('Download started');
  };

  const handleDelete = async (type, id) => {
    try {
       if (type === 'export') {
         await dataManagementAPI.deleteExportRecord(id);
         setExportRecords(prev => prev.filter(record => record.id !== id));
       } else {
         await dataManagementAPI.deleteBackupRecord(id);
         setBackupRecords(prev => prev.filter(record => record.id !== id));
       }
      toast.success(`${type} record deleted successfully`);
    } catch (error) {
      toast.error(`Failed to delete ${type} record`);
    }
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case 'completed': return <Badge bg="success">Completed</Badge>;
      case 'processing': return <Badge bg="warning">Processing</Badge>;
      case 'failed': return <Badge bg="danger">Failed</Badge>;
      default: return <Badge bg="secondary">Unknown</Badge>;
    }
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'export':
        return (
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <Card.Title>Data Export Records</Card.Title>
              <Button variant="primary" onClick={() => setShowExportModal(true)}>
                New Export
              </Button>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Filename</th>
                    <th>Format</th>
                    <th>Size</th>
                    <th>Status</th>
                    <th>Created At</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {exportRecords.map((record) => (
                    <tr key={record.id}>
                      <td>{record.filename}</td>
                      <td>{record.format}</td>
                      <td>{record.size}</td>
                      <td>{getStatusBadge(record.status)}</td>
                      <td>{record.createdAt}</td>
                      <td>
                        {record.downloadUrl && (
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                            onClick={() => handleDownload(record.downloadUrl, record.filename)}
                          >
                            Download
                          </Button>
                        )}
                        <Button
                          variant="outline-danger"
                          size="sm"
                          onClick={() => handleDelete('export', record.id)}
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
        );

      case 'backup':
        return (
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <Card.Title>System Backup Records</Card.Title>
              <Button variant="primary" onClick={() => setShowBackupModal(true)}>
                Create Backup
              </Button>
            </Card.Header>
            <Card.Body>
              <Table striped bordered hover>
                <thead>
                  <tr>
                    <th>Filename</th>
                    <th>Size</th>
                    <th>Status</th>
                    <th>Description</th>
                    <th>Created At</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {backupRecords.map((record) => (
                    <tr key={record.id}>
                      <td>{record.filename}</td>
                      <td>{record.size}</td>
                      <td>{getStatusBadge(record.status)}</td>
                      <td>{record.description}</td>
                      <td>{record.createdAt}</td>
                      <td>
                        {record.downloadUrl && (
                          <Button
                            variant="outline-primary"
                            size="sm"
                            className="me-2"
                            onClick={() => handleDownload(record.downloadUrl, record.filename)}
                          >
                            Download
                          </Button>
                        )}
                        <Button
                          variant="outline-danger"
                          size="sm"
                          onClick={() => handleDelete('backup', record.id)}
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
        );

      case 'import':
        return (
          <Card>
            <Card.Header className="d-flex justify-content-between align-items-center">
              <Card.Title>Data Import</Card.Title>
              <Button variant="primary" onClick={() => setShowImportModal(true)}>
                Import Data
              </Button>
            </Card.Header>
            <Card.Body>
              <div className="text-center py-5">
                <h5>Import Data from File</h5>
                <p className="text-muted">
                  Import data from previously exported files or backup archives.
                  Supported formats: JSON, CSV, SQL
                </p>
                <Button variant="primary" onClick={() => setShowImportModal(true)}>
                  Select File to Import
                </Button>
              </div>
            </Card.Body>
          </Card>
        );

      default:
        return null;
    }
  };

  return (
    <Container>
      <Row>
        <Col>
          <h2>Data Management</h2>
          
          {/* Tab Navigation */}
          <Card className="mb-4">
            <Card.Body>
              <div className="d-flex gap-2">
                <Button
                  variant={activeTab === 'export' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('export')}
                >
                  Data Export
                </Button>
                <Button
                  variant={activeTab === 'backup' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('backup')}
                >
                  System Backup
                </Button>
                <Button
                  variant={activeTab === 'import' ? 'primary' : 'outline-primary'}
                  onClick={() => setActiveTab('import')}
                >
                  Data Import
                </Button>
              </div>
            </Card.Body>
          </Card>

          {/* Tab Content */}
          {renderTabContent()}

          {/* Export Modal */}
          <Modal show={showExportModal} onHide={() => setShowExportModal(false)} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Export Data</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              {exportProgress > 0 ? (
                <div>
                  <h6>Export Progress</h6>
                  <ProgressBar now={exportProgress} label={`${exportProgress}%`} />
                </div>
              ) : (
                <Form>
                  <Form.Group className="mb-3">
                    <Form.Label>Export Format</Form.Label>
                    <Form.Select
                      value={exportConfig.format}
                      onChange={(e) => setExportConfig({...exportConfig, format: e.target.value})}
                    >
                      <option value="json">JSON</option>
                      <option value="csv">CSV</option>
                      <option value="sql">SQL</option>
                    </Form.Select>
                  </Form.Group>
                  <Form.Group className="mb-3">
                    <Form.Label>Date Range</Form.Label>
                    <Form.Select
                      value={exportConfig.dateRange}
                      onChange={(e) => setExportConfig({...exportConfig, dateRange: e.target.value})}
                    >
                      <option value="all">All Data</option>
                      <option value="last_month">Last Month</option>
                      <option value="last_week">Last Week</option>
                      <option value="today">Today</option>
                    </Form.Select>
                  </Form.Group>
                </Form>
              )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={() => setShowExportModal(false)}>
                Cancel
              </Button>
              {exportProgress === 0 && (
                <Button variant="primary" onClick={handleExport}>
                  Start Export
                </Button>
              )}
            </Modal.Footer>
          </Modal>

          {/* Backup Modal */}
          <Modal show={showBackupModal} onHide={() => setShowBackupModal(false)} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Create System Backup</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              {backupProgress > 0 ? (
                <div>
                  <h6>Backup Progress</h6>
                  <ProgressBar now={backupProgress} label={`${backupProgress}%`} />
                </div>
              ) : (
                <Form>
                  <Form.Group className="mb-3">
                    <Form.Label>Description</Form.Label>
                    <Form.Control
                      type="text"
                      placeholder="Enter backup description"
                      value={backupConfig.description}
                      onChange={(e) => setBackupConfig({...backupConfig, description: e.target.value})}
                    />
                  </Form.Group>
                  <Form.Group className="mb-3">
                    <Form.Check
                      type="checkbox"
                      label="Include configuration files"
                      checked={backupConfig.includeFiles}
                      onChange={(e) => setBackupConfig({...backupConfig, includeFiles: e.target.checked})}
                    />
                  </Form.Group>
                  <Form.Group className="mb-3">
                    <Form.Check
                      type="checkbox"
                      label="Enable compression"
                      checked={backupConfig.compression}
                      onChange={(e) => setBackupConfig({...backupConfig, compression: e.target.checked})}
                    />
                  </Form.Group>
                </Form>
              )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={() => setShowBackupModal(false)}>
                Cancel
              </Button>
              {backupProgress === 0 && (
                <Button variant="primary" onClick={handleBackup}>
                  Create Backup
                </Button>
              )}
            </Modal.Footer>
          </Modal>

          {/* Import Modal */}
          <Modal show={showImportModal} onHide={() => setShowImportModal(false)} size="lg">
            <Modal.Header closeButton>
              <Modal.Title>Import Data</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              {importProgress > 0 ? (
                <div>
                  <h6>Import Progress</h6>
                  <ProgressBar now={importProgress} label={`${importProgress}%`} />
                </div>
              ) : (
                <Form>
                  <Form.Group className="mb-3">
                    <Form.Label>Select File</Form.Label>
                    <Form.Control
                      type="file"
                      accept=".json,.csv,.sql,.tar.gz,.zip"
                      onChange={(e) => setImportFile(e.target.files[0])}
                    />
                    <Form.Text className="text-muted">
                      Supported formats: JSON, CSV, SQL, TAR.GZ, ZIP
                    </Form.Text>
                  </Form.Group>
                  {importFile && (
                    <div className="alert alert-info">
                      <strong>Selected file:</strong> {importFile.name} ({(importFile.size / 1024 / 1024).toFixed(2)} MB)
                    </div>
                  )}
                </Form>
              )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={() => setShowImportModal(false)}>
                Cancel
              </Button>
              {importProgress === 0 && (
                <Button variant="primary" onClick={handleImport} disabled={!importFile}>
                  Start Import
                </Button>
              )}
            </Modal.Footer>
          </Modal>
        </Col>
      </Row>
    </Container>
  );
};

export default DataManagement;