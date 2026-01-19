import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Row, 
  Col, 
  Card, 
  Button, 
  Badge,
  Alert,
  Spinner,
  ProgressBar,
  Table,
  Form,
  InputGroup
} from 'react-bootstrap';
import { monitoringApi, analyticsApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

const SystemMonitoring = () => {
  const { auth } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval, setRefreshInterval] = useState(5000);

  // State for monitoring data
  const [systemHealth, setSystemHealth] = useState(null);
  const [realTimeMetrics, setRealTimeMetrics] = useState(null);
  const [performanceMetrics, setPerformanceMetrics] = useState(null);
  const [errorLogs, setErrorLogs] = useState([]);
  const [apiMetrics, setApiMetrics] = useState(null);
  const [databaseStats, setDatabaseStats] = useState(null);
  const [cacheStats, setCacheStats] = useState(null);
  const [serverMetrics, setServerMetrics] = useState(null);

  useEffect(() => {
    if (auth?.user_id) {
      loadSystemHealth();
      loadRealTimeMetrics();
      loadPerformanceMetrics();
      loadErrorLogs();
      loadApiMetrics();
      loadDatabaseStats();
      loadCacheStats();
      loadServerMetrics();
    }
  }, [auth]);

  useEffect(() => {
    let interval = null;
    if (autoRefresh) {
      interval = setInterval(() => {
        loadSystemHealth();
        loadRealTimeMetrics();
        loadPerformanceMetrics();
      }, refreshInterval);
    }
    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval]);

  const loadSystemHealth = async () => {
    try {
      const response = await monitoringApi.getSystemHealth();
      setSystemHealth(response.data);
    } catch (err) {
      console.error('Failed to load system health:', err);
    }
  };

  const loadRealTimeMetrics = async () => {
    try {
      const response = await analyticsApi.getRealTimeMetrics();
      setRealTimeMetrics(response.data.metrics);
    } catch (err) {
      console.error('Failed to load real-time metrics:', err);
    }
  };

  const loadPerformanceMetrics = async () => {
    try {
      const response = await monitoringApi.getPerformanceMetrics();
      setPerformanceMetrics(response.data.metrics);
    } catch (err) {
      console.error('Failed to load performance metrics:', err);
    }
  };

  const loadErrorLogs = async () => {
    try {
      const response = await monitoringApi.getErrorLogs();
      setErrorLogs(response.data.logs || []);
    } catch (err) {
      console.error('Failed to load error logs:', err);
    }
  };

  const loadApiMetrics = async () => {
    try {
      const response = await monitoringApi.getApiMetrics();
      setApiMetrics(response.data.metrics);
    } catch (err) {
      console.error('Failed to load API metrics:', err);
    }
  };

  const loadDatabaseStats = async () => {
    try {
      const response = await monitoringApi.getDatabaseStats();
      setDatabaseStats(response.data.stats);
    } catch (err) {
      console.error('Failed to load database stats:', err);
    }
  };

  const loadCacheStats = async () => {
    try {
      const response = await monitoringApi.getCacheStats();
      setCacheStats(response.data.stats);
    } catch (err) {
      console.error('Failed to load cache stats:', err);
    }
  };

  const loadServerMetrics = async () => {
    try {
      const response = await monitoringApi.getServerMetrics();
      setServerMetrics(response.data.metrics);
    } catch (err) {
      console.error('Failed to load server metrics:', err);
    }
  };

  const clearMessages = () => {
    setError('');
  };

  const getStatusBadge = (status) => {
    const variants = {
      healthy: 'success',
      warning: 'warning',
      critical: 'danger',
      unknown: 'secondary'
    };
    return <Badge bg={variants[status] || 'unknown'}>{status}</Badge>;
  };

  const getProgressBarVariant = (percentage) => {
    if (percentage >= 90) return 'danger';
    if (percentage >= 75) return 'warning';
    if (percentage >= 50) return 'info';
    return 'success';
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatUptime = (seconds) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${minutes}m`;
  };

  return (
    <Container fluid className="py-4">
      <Row>
        <Col>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>🖥️ System Monitoring</h2>
            <div className="d-flex align-items-center">
              <Form.Check
                type="switch"
                label="Auto Refresh"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="me-3"
              />
              <InputGroup style={{ width: '150px' }} className="me-2">
                <Form.Select
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                >
                  <option value={5000}>5s</option>
                  <option value={10000}>10s</option>
                  <option value={30000}>30s</option>
                  <option value={60000}>60s</option>
                </Form.Select>
              </InputGroup>
              <Button variant="outline-info" onClick={() => {
                loadSystemHealth();
                loadRealTimeMetrics();
                loadPerformanceMetrics();
                loadErrorLogs();
                loadApiMetrics();
                loadDatabaseStats();
                loadCacheStats();
                loadServerMetrics();
              }}>
                🔄 Refresh All
              </Button>
            </div>
          </div>

          {/* Alerts */}
          {error && (
            <Alert variant="danger" dismissible onClose={clearMessages}>
              {error}
            </Alert>
          )}

          {/* System Health Overview */}
          {systemHealth && (
            <Row className="mb-4">
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-primary">{getStatusBadge(systemHealth.database.status)}</h4>
                    <p className="mb-0">Database</p>
                    <small className="text-muted">{systemHealth.database.response_time}ms</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-success">{getStatusBadge(systemHealth.redis.status)}</h4>
                    <p className="mb-0">Redis Cache</p>
                    <small className="text-muted">{systemHealth.redis.response_time}ms</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-warning">{getStatusBadge(systemHealth.api.status)}</h4>
                    <p className="mb-0">API Server</p>
                    <small className="text-muted">{systemHealth.api.response_time}ms</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-info">{getStatusBadge(systemHealth.websocket.status)}</h4>
                    <p className="mb-0">WebSocket</p>
                    <small className="text-muted">{systemHealth.websocket.connections} connections</small>
                  </Card.Body>
                </Card>
              </Col>
            </Row>
          )}

          {/* Real-time Metrics */}
          {realTimeMetrics && (
            <Row className="mb-4">
              <Col md={3}>
                <Card>
                  <Card.Header as="h6">👥 Active Users</Card.Header>
                  <Card.Body className="text-center">
                    <h3 className="text-primary">{realTimeMetrics.active_users}</h3>
                    <small className="text-muted">Currently online</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card>
                  <Card.Header as="h6">🎬 Live Streams</Card.Header>
                  <Card.Body className="text-center">
                    <h3 className="text-success">{realTimeMetrics.concurrent_streams}</h3>
                    <small className="text-muted">Streaming now</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card>
                  <Card.Header as="h6">📡 API Requests</Card.Header>
                  <Card.Body className="text-center">
                    <h3 className="text-warning">{realTimeMetrics.api_requests_per_minute}</h3>
                    <small className="text-muted">Per minute</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card>
                  <Card.Header as="h6">💻 Server Load</Card.Header>
                  <Card.Body className="text-center">
                    <h3 className="text-info">{realTimeMetrics.server_load}%</h3>
                    <ProgressBar 
                      now={realTimeMetrics.server_load} 
                      variant={getProgressBarVariant(realTimeMetrics.server_load)}
                    />
                  </Card.Body>
                </Card>
              </Col>
            </Row>
          )}

          {/* Detailed Metrics */}
          <Row>
            {/* Server Metrics */}
            <Col lg={6}>
              <Card className="mb-4">
                <Card.Header as="h5">🖥️ Server Metrics</Card.Header>
                <Card.Body>
                  {serverMetrics ? (
                    <div>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>CPU Usage:</strong></p>
                          <ProgressBar 
                            now={serverMetrics.cpu_usage} 
                            variant={getProgressBarVariant(serverMetrics.cpu_usage)}
                            label={`${serverMetrics.cpu_usage}%`}
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Memory Usage:</strong></p>
                          <ProgressBar 
                            now={serverMetrics.memory_usage} 
                            variant={getProgressBarVariant(serverMetrics.memory_usage)}
                            label={`${serverMetrics.memory_usage}%`}
                          />
                        </Col>
                      </Row>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Disk Usage:</strong></p>
                          <ProgressBar 
                            now={serverMetrics.disk_usage} 
                            variant={getProgressBarVariant(serverMetrics.disk_usage)}
                            label={`${serverMetrics.disk_usage}%`}
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Network I/O:</strong></p>
                          <ProgressBar 
                            now={serverMetrics.network_usage} 
                            variant="info"
                            label={`${serverMetrics.network_usage}%`}
                          />
                        </Col>
                      </Row>
                      <Row>
                        <Col md={6}>
                          <p><strong>Uptime:</strong> {formatUptime(serverMetrics.uptime)}</p>
                          <p><strong>Processes:</strong> {serverMetrics.process_count}</p>
                        </Col>
                        <Col md={6}>
                          <p><strong>Load Average:</strong> {serverMetrics.load_average}</p>
                          <p><strong>Temperature:</strong> {serverMetrics.temperature}°C</p>
                        </Col>
                      </Row>
                    </div>
                  ) : (
                    <Spinner animation="border" size="sm" />
                  )}
                </Card.Body>
              </Card>
            </Col>

            {/* Database Stats */}
            <Col lg={6}>
              <Card className="mb-4">
                <Card.Header as="h5">🗄️ Database Statistics</Card.Header>
                <Card.Body>
                  {databaseStats ? (
                    <div>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Connections:</strong> {databaseStats.active_connections}/{databaseStats.max_connections}</p>
                          <ProgressBar 
                            now={(databaseStats.active_connections / databaseStats.max_connections) * 100} 
                            variant="info"
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Query Rate:</strong> {databaseStats.queries_per_second}/sec</p>
                          <ProgressBar 
                            now={(databaseStats.queries_per_second / 1000) * 100} 
                            variant="warning"
                          />
                        </Col>
                      </Row>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Database Size:</strong> {formatBytes(databaseStats.database_size)}</p>
                          <p><strong>Collections:</strong> {databaseStats.collection_count}</p>
                        </Col>
                        <Col md={6}>
                          <p><strong>Avg Query Time:</strong> {databaseStats.avg_query_time}ms</p>
                          <p><strong>Slow Queries:</strong> {databaseStats.slow_queries}</p>
                        </Col>
                      </Row>
                      <Row>
                        <Col md={6}>
                          <p><strong>Index Hit Rate:</strong> {databaseStats.index_hit_rate}%</p>
                          <ProgressBar 
                            now={databaseStats.index_hit_rate} 
                            variant="success"
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Cache Hit Rate:</strong> {databaseStats.cache_hit_rate}%</p>
                          <ProgressBar 
                            now={databaseStats.cache_hit_rate} 
                            variant="success"
                          />
                        </Col>
                      </Row>
                    </div>
                  ) : (
                    <Spinner animation="border" size="sm" />
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>

          <Row>
            {/* Cache Stats */}
            <Col lg={6}>
              <Card className="mb-4">
                <Card.Header as="h5">⚡ Cache Statistics</Card.Header>
                <Card.Body>
                  {cacheStats ? (
                    <div>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Memory Used:</strong> {formatBytes(cacheStats.memory_used)}/{formatBytes(cacheStats.memory_total)}</p>
                          <ProgressBar 
                            now={(cacheStats.memory_used / cacheStats.memory_total) * 100} 
                            variant={getProgressBarVariant((cacheStats.memory_used / cacheStats.memory_total) * 100)}
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Hit Rate:</strong> {cacheStats.hit_rate}%</p>
                          <ProgressBar 
                            now={cacheStats.hit_rate} 
                            variant="success"
                          />
                        </Col>
                      </Row>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Keys:</strong> {cacheStats.total_keys}</p>
                          <p><strong>Expires:</strong> {cacheStats.expired_keys}</p>
                        </Col>
                        <Col md={6}>
                          <p><strong>Operations/sec:</strong> {cacheStats.ops_per_second}</p>
                          <p><strong>Evicted Keys:</strong> {cacheStats.evicted_keys}</p>
                        </Col>
                      </Row>
                      <Row>
                        <Col md={6}>
                          <p><strong>Connected Clients:</strong> {cacheStats.connected_clients}</p>
                        </Col>
                        <Col md={6}>
                          <p><strong>Uptime:</strong> {formatUptime(cacheStats.uptime)}</p>
                        </Col>
                      </Row>
                    </div>
                  ) : (
                    <Spinner animation="border" size="sm" />
                  )}
                </Card.Body>
              </Card>
            </Col>

            {/* API Metrics */}
            <Col lg={6}>
              <Card className="mb-4">
                <Card.Header as="h5">🌐 API Metrics</Card.Header>
                <Card.Body>
                  {apiMetrics ? (
                    <div>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Requests/min:</strong> {apiMetrics.requests_per_minute}</p>
                          <ProgressBar 
                            now={(apiMetrics.requests_per_minute / 1000) * 100} 
                            variant="info"
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Response Time:</strong> {apiMetrics.avg_response_time}ms</p>
                          <ProgressBar 
                            now={Math.min((apiMetrics.avg_response_time / 1000) * 100, 100)} 
                            variant={apiMetrics.avg_response_time < 200 ? 'success' : 'warning'}
                          />
                        </Col>
                      </Row>
                      <Row className="mb-3">
                        <Col md={6}>
                          <p><strong>Success Rate:</strong> {apiMetrics.success_rate}%</p>
                          <ProgressBar 
                            now={apiMetrics.success_rate} 
                            variant="success"
                          />
                        </Col>
                        <Col md={6}>
                          <p><strong>Error Rate:</strong> {apiMetrics.error_rate}%</p>
                          <ProgressBar 
                            now={apiMetrics.error_rate} 
                            variant="danger"
                          />
                        </Col>
                      </Row>
                      <Row>
                        <Col md={6}>
                          <p><strong>2xx Responses:</strong> {apiMetrics.status_codes_2xx}</p>
                          <p><strong>4xx Responses:</strong> {apiMetrics.status_codes_4xx}</p>
                        </Col>
                        <Col md={6}>
                          <p><strong>5xx Responses:</strong> {apiMetrics.status_codes_5xx}</p>
                          <p><strong>Total Requests:</strong> {apiMetrics.total_requests}</p>
                        </Col>
                      </Row>
                    </div>
                  ) : (
                    <Spinner animation="border" size="sm" />
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>

          {/* Error Logs */}
          <Row>
            <Col lg={12}>
              <Card className="mb-4">
                <Card.Header as="h5">🚨 Recent Error Logs</Card.Header>
                <Card.Body>
                  {errorLogs.length > 0 ? (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>Timestamp</th>
                          <th>Level</th>
                          <th>Service</th>
                          <th>Message</th>
                          <th>Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {errorLogs.slice(0, 10).map((log, index) => (
                          <tr key={index}>
                            <td>{new Date(log.timestamp).toLocaleString()}</td>
                            <td>
                              <Badge bg={
                                log.level === 'error' ? 'danger' : 
                                log.level === 'warning' ? 'warning' : 
                                log.level === 'info' ? 'info' : 'secondary'
                              }>
                                {log.level}
                              </Badge>
                            </td>
                            <td>{log.service}</td>
                            <td className="text-truncate" style={{ maxWidth: '300px' }}>
                              {log.message}
                            </td>
                            <td>
                              <Button variant="outline-primary" size="sm">
                                Details
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  ) : (
                    <Alert variant="success">No errors in the last 24 hours</Alert>
                  )}
                </Card.Body>
              </Card>
            </Col>
          </Row>

          {/* Performance Metrics */}
          {performanceMetrics && (
            <Row>
              <Col lg={12}>
                <Card>
                  <Card.Header as="h5">📊 Performance Metrics</Card.Header>
                  <Card.Body>
                    <Row>
                      <Col md={3}>
                        <h6>Response Times</h6>
                        <p><strong>Average:</strong> {performanceMetrics.avg_response_time}ms</p>
                        <p><strong>P95:</strong> {performanceMetrics.p95_response_time}ms</p>
                        <p><strong>P99:</strong> {performanceMetrics.p99_response_time}ms</p>
                      </Col>
                      <Col md={3}>
                        <h6>Throughput</h6>
                        <p><strong>Requests/sec:</strong> {performanceMetrics.requests_per_second}</p>
                        <p><strong>Peak:</strong> {performanceMetrics.peak_requests_per_second}</p>
                        <p><strong>Today:</strong> {performanceMetrics.total_requests_today}</p>
                      </Col>
                      <Col md={3}>
                        <h6>Errors</h6>
                        <p><strong>Error Rate:</strong> {performanceMetrics.error_rate}%</p>
                        <p><strong>Timeouts:</strong> {performanceMetrics.timeouts}</p>
                        <p><strong>5xx Errors:</strong> {performanceMetrics.server_errors}</p>
                      </Col>
                      <Col md={3}>
                        <h6>Availability</h6>
                        <p><strong>Uptime:</strong> {performanceMetrics.uptime_percentage}%</p>
                        <p><strong>Downtime:</strong> {performanceMetrics.downtime_minutes}min</p>
                        <p><strong>SLA:</strong> {performanceMetrics.sla_met ? '✅ Met' : '❌ Not Met'}</p>
                      </Col>
                    </Row>
                  </Card.Body>
                </Card>
              </Col>
            </Row>
          )}
        </Col>
      </Row>
    </Container>
  );
};

export default SystemMonitoring;