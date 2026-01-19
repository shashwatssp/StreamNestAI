import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Row, 
  Col, 
  Card, 
  Button, 
  Form, 
  Table,
  Badge,
  Alert,
  Spinner,
  Nav,
  Tab,
  ProgressBar
} from 'react-bootstrap';
import { analyticsApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

const AnalyticsDashboard = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dateRange, setDateRange] = useState('7d');

  // State for different analytics data
  const [overviewStats, setOverviewStats] = useState(null);
  const [userAnalytics, setUserAnalytics] = useState(null);
  const [contentAnalytics, setContentAnalytics] = useState(null);
  const [engagementMetrics, setEngagementMetrics] = useState(null);
  const [realTimeMetrics, setRealTimeMetrics] = useState(null);
  const [topContent, setTopContent] = useState([]);
  const [userActivity, setUserActivity] = useState([]);
  const [revenueData, setRevenueData] = useState(null);

  useEffect(() => {
    if (auth?.user_id) {
      loadOverviewStats();
      loadRealTimeMetrics();
    }
  }, [auth, dateRange]);

  useEffect(() => {
    if (activeTab === 'users') loadUserAnalytics();
    if (activeTab === 'content') loadContentAnalytics();
    if (activeTab === 'engagement') loadEngagementMetrics();
    if (activeTab === 'top-content') loadTopContent();
    if (activeTab === 'activity') loadUserActivity();
    if (activeTab === 'revenue') loadRevenueData();
  }, [activeTab, dateRange]);

  const loadOverviewStats = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getOverviewStats(dateRange);
      setOverviewStats(response.data.stats);
    } catch (err) {
      setError('Failed to load overview statistics');
    } finally {
      setLoading(false);
    }
  };

  const loadUserAnalytics = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getUserAnalytics(dateRange);
      setUserAnalytics(response.data.analytics);
    } catch (err) {
      setError('Failed to load user analytics');
    } finally {
      setLoading(false);
    }
  };

  const loadContentAnalytics = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getContentAnalytics(dateRange);
      setContentAnalytics(response.data.analytics);
    } catch (err) {
      setError('Failed to load content analytics');
    } finally {
      setLoading(false);
    }
  };

  const loadEngagementMetrics = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getEngagementMetrics(dateRange);
      setEngagementMetrics(response.data.metrics);
    } catch (err) {
      setError('Failed to load engagement metrics');
    } finally {
      setLoading(false);
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

  const loadTopContent = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getTopContent(dateRange);
      setTopContent(response.data.content || []);
    } catch (err) {
      setError('Failed to load top content');
    } finally {
      setLoading(false);
    }
  };

  const loadUserActivity = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getUserActivity(dateRange);
      setUserActivity(response.data.activities || []);
    } catch (err) {
      setError('Failed to load user activity');
    } finally {
      setLoading(false);
    }
  };

  const loadRevenueData = async () => {
    try {
      setLoading(true);
      const response = await analyticsApi.getRevenueAnalytics(dateRange);
      setRevenueData(response.data.revenue);
    } catch (err) {
      setError('Failed to load revenue data');
    } finally {
      setLoading(false);
    }
  };

  const clearMessages = () => {
    setError('');
  };

  const formatNumber = (num) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  const formatPercentage = (value) => {
    return (value * 100).toFixed(1) + '%';
  };

  const getGrowthBadge = (growth) => {
    if (growth > 0) return <Badge bg="success">↑ {formatPercentage(growth)}</Badge>;
    if (growth < 0) return <Badge bg="danger">↓ {formatPercentage(Math.abs(growth))}</Badge>;
    return <Badge bg="secondary">→ 0%</Badge>;
  };

  return (
    <Container fluid className="py-4">
      <Row>
        <Col>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>📊 Analytics Dashboard</h2>
            <div className="d-flex align-items-center">
              <Form.Select 
                value={dateRange} 
                onChange={(e) => setDateRange(e.target.value)}
                className="me-2"
                style={{ width: '150px' }}
              >
                <option value="1d">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
                <option value="90d">Last 90 Days</option>
              </Form.Select>
              <Button variant="outline-info" onClick={loadRealTimeMetrics}>
                🔄 Refresh
              </Button>
            </div>
          </div>

          {/* Alerts */}
          {error && (
            <Alert variant="danger" dismissible onClose={clearMessages}>
              {error}
            </Alert>
          )}

          {/* Real-time Metrics */}
          {realTimeMetrics && (
            <Row className="mb-4">
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-primary">{realTimeMetrics.active_users}</h4>
                    <p className="mb-0">Active Users</p>
                    <small className="text-muted">Now</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-success">{realTimeMetrics.concurrent_streams}</h4>
                    <p className="mb-0">Live Streams</p>
                    <small className="text-muted">Now</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-warning">{realTimeMetrics.api_requests_per_minute}</h4>
                    <p className="mb-0">API Requests/min</p>
                    <small className="text-muted">Now</small>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h4 className="text-info">{realTimeMetrics.server_load}%</h4>
                    <p className="mb-0">Server Load</p>
                    <small className="text-muted">Now</small>
                  </Card.Body>
                </Card>
              </Col>
            </Row>
          )}

          {/* Main Content Tabs */}
          <Card>
            <Card.Header>
              <Nav variant="tabs" activeKey={activeTab} onSelect={(k) => setActiveTab(k)}>
                <Nav.Item>
                  <Nav.Link eventKey="overview">📈 Overview</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="users">👥 Users</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="content">🎬 Content</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="engagement">💡 Engagement</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="top-content">🏆 Top Content</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="activity">⚡ Activity</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="revenue">💰 Revenue</Nav.Link>
                </Nav.Item>
              </Nav>
            </Card.Header>

            <Card.Body>
              <Tab.Content>
                {/* Overview Tab */}
                <Tab.Pane active={activeTab === 'overview'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : overviewStats ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">User Statistics</Card.Header>
                          <Card.Body>
                            <Row>
                              <Col md={6}>
                                <p><strong>Total Users:</strong> {formatNumber(overviewStats.total_users)}</p>
                                <p><strong>New Users:</strong> {formatNumber(overviewStats.new_users)}</p>
                                <p><strong>Active Users:</strong> {formatNumber(overviewStats.active_users)}</p>
                              </Col>
                              <Col md={6}>
                                <p><strong>User Growth:</strong> {getGrowthBadge(overviewStats.user_growth)}</p>
                                <p><strong>Retention Rate:</strong> {formatPercentage(overviewStats.retention_rate)}</p>
                                <p><strong>Churn Rate:</strong> {formatPercentage(overviewStats.churn_rate)}</p>
                              </Col>
                            </Row>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Content Statistics</Card.Header>
                          <Card.Body>
                            <Row>
                              <Col md={6}>
                                <p><strong>Total Movies:</strong> {formatNumber(overviewStats.total_movies)}</p>
                                <p><strong>Total Reviews:</strong> {formatNumber(overviewStats.total_reviews)}</p>
                                <p><strong>Total Watchlists:</strong> {formatNumber(overviewStats.total_watchlists)}</p>
                              </Col>
                              <Col md={6}>
                                <p><strong>Content Growth:</strong> {getGrowthBadge(overviewStats.content_growth)}</p>
                                <p><strong>Avg Rating:</strong> {overviewStats.average_rating?.toFixed(1) || 'N/A'}</p>
                                <p><strong>Review Rate:</strong> {formatPercentage(overviewStats.review_rate)}</p>
                              </Col>
                            </Row>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">No overview data available</Alert>
                  )}
                </Tab.Pane>

                {/* Users Tab */}
                <Tab.Pane active={activeTab === 'users'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : userAnalytics ? (
                    <Row>
                      <Col md={8}>
                        <Card>
                          <Card.Header as="h6">User Growth Trend</Card.Header>
                          <Card.Body>
                            <p><strong>Daily Active Users:</strong> {formatNumber(userAnalytics.daily_active_users)}</p>
                            <p><strong>Weekly Active Users:</strong> {formatNumber(userAnalytics.weekly_active_users)}</p>
                            <p><strong>Monthly Active Users:</strong> {formatNumber(userAnalytics.monthly_active_users)}</p>
                            <p><strong>User Acquisition Cost:</strong> ${userAnalytics.user_acquisition_cost}</p>
                            <p><strong>Lifetime Value:</strong> ${userAnalytics.lifetime_value}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={4}>
                        <Card>
                          <Card.Header as="h6">User Demographics</Card.Header>
                          <Card.Body>
                            <p><strong>Avg Session Duration:</strong> {userAnalytics.avg_session_duration} min</p>
                            <p><strong>Pages per Session:</strong> {userAnalytics.pages_per_session}</p>
                            <p><strong>Bounce Rate:</strong> {formatPercentage(userAnalytics.bounce_rate)}</p>
                            <p><strong>Return User Rate:</strong> {formatPercentage(userAnalytics.return_user_rate)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">No user analytics data available</Alert>
                  )}
                </Tab.Pane>

                {/* Content Tab */}
                <Tab.Pane active={activeTab === 'content'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : contentAnalytics ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Content Performance</Card.Header>
                          <Card.Body>
                            <p><strong>Total Views:</strong> {formatNumber(contentAnalytics.total_views)}</p>
                            <p><strong>Unique Viewers:</strong> {formatNumber(contentAnalytics.unique_viewers)}</p>
                            <p><strong>Avg Watch Time:</strong> {contentAnalytics.avg_watch_time} min</p>
                            <p><strong>Completion Rate:</strong> {formatPercentage(contentAnalytics.completion_rate)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Content Quality</Card.Header>
                          <Card.Body>
                            <p><strong>Avg Rating:</strong> {contentAnalytics.avg_rating?.toFixed(2) || 'N/A'}</p>
                            <p><strong>Total Reviews:</strong> {formatNumber(contentAnalytics.total_reviews)}</p>
                            <p><strong>Review Engagement:</strong> {formatPercentage(contentAnalytics.review_engagement)}</p>
                            <p><strong>Content Score:</strong> {contentAnalytics.content_score?.toFixed(1) || 'N/A'}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">No content analytics data available</Alert>
                  )}
                </Tab.Pane>

                {/* Engagement Tab */}
                <Tab.Pane active={activeTab === 'engagement'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : engagementMetrics ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">User Engagement</Card.Header>
                          <Card.Body>
                            <p><strong>Daily Engagement Rate:</strong> {formatPercentage(engagementMetrics.daily_engagement_rate)}</p>
                            <p><strong>Weekly Engagement Rate:</strong> {formatPercentage(engagementMetrics.weekly_engagement_rate)}</p>
                            <p><strong>Monthly Engagement Rate:</strong> {formatPercentage(engagementMetrics.monthly_engagement_rate)}</p>
                            <p><strong>Feature Adoption:</strong> {formatPercentage(engagementMetrics.feature_adoption_rate)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Interaction Metrics</Card.Header>
                          <Card.Body>
                            <p><strong>Likes per User:</strong> {engagementMetrics.avg_likes_per_user}</p>
                            <p><strong>Comments per User:</strong> {engagementMetrics.avg_comments_per_user}</p>
                            <p><strong>Shares per User:</strong> {engagementMetrics.avg_shares_per_user}</p>
                            <p><strong>Watchlist Additions:</strong> {formatNumber(engagementMetrics.watchlist_additions)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">No engagement metrics available</Alert>
                  )}
                </Tab.Pane>

                {/* Top Content Tab */}
                <Tab.Pane active={activeTab === 'top-content'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : topContent.length > 0 ? (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>Rank</th>
                          <th>Title</th>
                          <th>Genre</th>
                          <th>Views</th>
                          <th>Rating</th>
                          <th>Reviews</th>
                          <th>Engagement</th>
                        </tr>
                      </thead>
                      <tbody>
                        {topContent.map((content, index) => (
                          <tr key={index}>
                            <td>
                              <Badge bg="warning">#{index + 1}</Badge>
                            </td>
                            <td>
                              <strong>{content.title}</strong>
                              <br />
                              <small className="text-muted">{content.year}</small>
                            </td>
                            <td>{content.genre}</td>
                            <td>{formatNumber(content.views)}</td>
                            <td>
                              <Badge bg="success">⭐ {content.rating?.toFixed(1)}</Badge>
                            </td>
                            <td>{formatNumber(content.review_count)}</td>
                            <td>
                              <ProgressBar 
                                now={content.engagement_score * 100} 
                                label={formatPercentage(content.engagement_score)}
                                variant="info"
                              />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  ) : (
                    <Alert variant="info">No top content data available</Alert>
                  )}
                </Tab.Pane>

                {/* Activity Tab */}
                <Tab.Pane active={activeTab === 'activity'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : userActivity.length > 0 ? (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>User</th>
                          <th>Activity Type</th>
                          <th>Content</th>
                          <th>Duration</th>
                          <th>Timestamp</th>
                          <th>Device</th>
                        </tr>
                      </thead>
                      <tbody>
                        {userActivity.map((activity, index) => (
                          <tr key={index}>
                            <td>{activity.user_display_name}</td>
                            <td>
                              <Badge bg="primary">{activity.activity_type}</Badge>
                            </td>
                            <td>{activity.content_title}</td>
                            <td>{activity.duration || '-'}</td>
                            <td>{new Date(activity.timestamp).toLocaleString()}</td>
                            <td>{activity.device_type}</td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  ) : (
                    <Alert variant="info">No user activity data available</Alert>
                  )}
                </Tab.Pane>

                {/* Revenue Tab */}
                <Tab.Pane active={activeTab === 'revenue'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : revenueData ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Revenue Overview</Card.Header>
                          <Card.Body>
                            <p><strong>Total Revenue:</strong> ${formatNumber(revenueData.total_revenue)}</p>
                            <p><strong>Monthly Revenue:</strong> ${formatNumber(revenueData.monthly_revenue)}</p>
                            <p><strong>Daily Revenue:</strong> ${formatNumber(revenueData.daily_revenue)}</p>
                            <p><strong>Revenue Growth:</strong> {getGrowthBadge(revenueData.revenue_growth)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Revenue Metrics</Card.Header>
                          <Card.Body>
                            <p><strong>ARPU:</strong> ${revenueData.average_revenue_per_user}</p>
                            <p><strong>ARPPU:</strong> ${revenueData.average_revenue_per_paying_user}</p>
                            <p><strong>Conversion Rate:</strong> {formatPercentage(revenueData.conversion_rate)}</p>
                            <p><strong>Churn Revenue:</strong> ${formatNumber(revenueData.churn_revenue)}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">No revenue data available</Alert>
                  )}
                </Tab.Pane>
              </Tab.Content>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default AnalyticsDashboard;