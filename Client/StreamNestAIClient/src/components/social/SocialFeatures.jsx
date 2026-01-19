import React, { useState, useEffect } from 'react';
import {
  Container,
  Row,
  Col,
  Card,
  Button,
  Form,
  InputGroup,
  Badge,
  ListGroup,
  Nav,
  Tab,
  Alert,
  Spinner,
  Modal
} from 'react-bootstrap';
import { socialApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

// Comprehensive logging utility
const logSocialFeature = (action, data = {}, level = 'info') => {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    feature: 'SOCIAL_FEATURES',
    action,
    data,
    level,
    userId: localStorage.getItem('user_id') || 'anonymous'
  };
  
  const logMessage = `[${timestamp}] 🟢 [SOCIAL_FEATURES] ${action}: ${JSON.stringify(data)}`;
  
  switch(level) {
    case 'error':
      console.error(logMessage, logEntry);
      break;
    case 'warn':
      console.warn(logMessage, logEntry);
      break;
    case 'debug':
      console.debug(logMessage, logEntry);
      break;
    default:
      console.log(logMessage, logEntry);
  }
};

const SocialFeatures = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('feed');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // State for different sections
  const [activityFeed, setActivityFeed] = useState([]);
  const [userProfile, setUserProfile] = useState(null);
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [socialStats, setSocialStats] = useState(null);
  const [searchResults, setSearchResults] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');

  // Modal states
  const [showFollowModal, setShowFollowModal] = useState(false);
  const [showActivityModal, setShowActivityModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);

  // Form states
  const [profileForm, setProfileForm] = useState({
    display_name: '',
    bio: '',
    avatar: '',
    is_private: false
  });
  const [activityForm, setActivityForm] = useState({
    type: 'watched',
    content_id: '',
    content_type: 'movie',
    title: '',
    description: ''
  });

  useEffect(() => {
    logSocialFeature('COMPONENT_MOUNT', { userId: auth?.user_id, hasAuth: !!auth?.user_id });
    if (auth?.user_id) {
      logSocialFeature('USER_DATA_LOAD_START', { userId: auth.user_id });
      loadUserData();
      logSocialFeature('ACTIVITY_FEED_LOAD_START', { userId: auth.user_id });
      loadActivityFeed();
      logSocialFeature('SOCIAL_STATS_LOAD_START', { userId: auth.user_id });
      loadSocialStats();
    } else {
      logSocialFeature('NO_AUTH_USER', {}, 'warn');
    }
  }, [auth]);

  const loadUserData = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logSocialFeature('API_CALL_START', { endpoint: 'getUserProfile', userId: auth.user_id });
      const response = await socialApi.getUserProfile(auth.user_id);
      const duration = Date.now() - startTime;
      
      logSocialFeature('API_CALL_SUCCESS', {
        endpoint: 'getUserProfile',
        userId: auth.user_id,
        duration: `${duration}ms`,
        dataReceived: !!response.data,
        profileFields: Object.keys(response.data || {})
      });
      
      setUserProfile(response.data);
      setProfileForm({
        display_name: response.data.display_name || '',
        bio: response.data.bio || '',
        avatar: response.data.avatar || '',
        is_private: response.data.is_private || false
      });
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('API_CALL_ERROR', {
        endpoint: 'getUserProfile',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load user profile');
    } finally {
      setLoading(false);
      logSocialFeature('USER_DATA_LOAD_END', { userId: auth.user_id, loading: false });
    }
  };

  const loadActivityFeed = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logSocialFeature('API_CALL_START', { endpoint: 'getActivityFeed', userId: auth.user_id });
      const response = await socialApi.getActivityFeed();
      const duration = Date.now() - startTime;
      
      logSocialFeature('API_CALL_SUCCESS', {
        endpoint: 'getActivityFeed',
        userId: auth.user_id,
        duration: `${duration}ms`,
        activitiesCount: response.data?.activities?.length || 0
      });
      
      setActivityFeed(response.data.activities || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('API_CALL_ERROR', {
        endpoint: 'getActivityFeed',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load activity feed');
    } finally {
      setLoading(false);
      logSocialFeature('ACTIVITY_FEED_LOAD_END', { userId: auth.user_id, activitiesCount: activityFeed.length });
    }
  };

  const loadSocialStats = async () => {
    const startTime = Date.now();
    try {
      logSocialFeature('API_CALL_START', { endpoint: 'getSocialStats', userId: auth.user_id });
      const response = await socialApi.getSocialStats();
      const duration = Date.now() - startTime;
      
      logSocialFeature('API_CALL_SUCCESS', {
        endpoint: 'getSocialStats',
        userId: auth.user_id,
        duration: `${duration}ms`,
        statsReceived: !!response.data.stats,
        statsFields: Object.keys(response.data.stats || {})
      });
      
      setSocialStats(response.data.stats);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('API_CALL_ERROR', {
        endpoint: 'getSocialStats',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
    }
  };

  const loadFollowers = async () => {
    try {
      setLoading(true);
      const response = await socialApi.getFollowers(auth.user_id);
      setFollowers(response.data.followers || []);
    } catch (err) {
      setError('Failed to load followers');
    } finally {
      setLoading(false);
    }
  };

  const loadFollowing = async () => {
    try {
      setLoading(true);
      const response = await socialApi.getFollowing(auth.user_id);
      setFollowing(response.data.following || []);
    } catch (err) {
      setError('Failed to load following');
    } finally {
      setLoading(false);
    }
  };

  const handleSearchUsers = async () => {
    if (!searchQuery.trim()) {
      logSocialFeature('SEARCH_USERS_EMPTY_QUERY', { query: searchQuery }, 'warn');
      return;
    }
    
    const startTime = Date.now();
    try {
      setLoading(true);
      logSocialFeature('SEARCH_USERS_START', { query: searchQuery, userId: auth.user_id });
      const response = await socialApi.searchUsers(searchQuery);
      const duration = Date.now() - startTime;
      
      logSocialFeature('SEARCH_USERS_SUCCESS', {
        query: searchQuery,
        userId: auth.user_id,
        duration: `${duration}ms`,
        resultsCount: response.data?.users?.length || 0
      });
      
      setSearchResults(response.data.users || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('SEARCH_USERS_ERROR', {
        query: searchQuery,
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to search users');
    } finally {
      setLoading(false);
    }
  };

  const handleFollowUser = async (userId) => {
    const startTime = Date.now();
    try {
      logSocialFeature('FOLLOW_USER_START', { targetUserId: userId, userId: auth.user_id });
      await socialApi.followUser(userId);
      const duration = Date.now() - startTime;
      
      logSocialFeature('FOLLOW_USER_SUCCESS', {
        targetUserId: userId,
        userId: auth.user_id,
        duration: `${duration}ms`
      });
      
      setSuccess('User followed successfully!');
      loadFollowing();
      loadSocialStats();
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('FOLLOW_USER_ERROR', {
        targetUserId: userId,
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to follow user');
    }
  };

  const handleUnfollowUser = async (userId) => {
    try {
      await socialApi.unfollowUser(userId);
      setSuccess('User unfollowed successfully!');
      loadFollowing();
      loadSocialStats();
    } catch (err) {
      setError('Failed to unfollow user');
    }
  };

  const handleUpdateProfile = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logSocialFeature('UPDATE_PROFILE_START', {
        userId: auth.user_id,
        formFields: Object.keys(profileForm),
        hasAvatar: !!profileForm.avatar,
        isPrivate: profileForm.is_private
      });
      
      await socialApi.updateProfile(profileForm);
      const duration = Date.now() - startTime;
      
      logSocialFeature('UPDATE_PROFILE_SUCCESS', {
        userId: auth.user_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Profile updated successfully!');
      loadUserData();
    } catch (err) {
      const duration = Date.now() - startTime;
      logSocialFeature('UPDATE_PROFILE_ERROR', {
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to update profile');
    }
  };

  const handleCreateActivity = async (e) => {
    e.preventDefault();
    try {
      await socialApi.createActivity(activityForm);
      setSuccess('Activity created successfully!');
      setShowActivityModal(false);
      loadActivityFeed();
      setActivityForm({
        type: 'watched',
        content_id: '',
        content_type: 'movie',
        title: '',
        description: ''
      });
    } catch (err) {
      setError('Failed to create activity');
    }
  };

  const clearMessages = () => {
    logSocialFeature('CLEAR_MESSAGES', { userId: auth.user_id });
    setError('');
    setSuccess('');
  };

  return (
    <Container fluid className="py-4">
      <Row>
        <Col lg={3}>
          {/* User Profile Card */}
          <Card className="mb-4">
            <Card.Header as="h5">👤 My Profile</Card.Header>
            <Card.Body>
              {userProfile ? (
                <div className="text-center">
                  <div className="mb-3">
                    {userProfile.avatar ? (
                      <img 
                        src={userProfile.avatar} 
                        alt="Avatar" 
                        className="rounded-circle"
                        style={{ width: '80px', height: '80px', objectFit: 'cover' }}
                      />
                    ) : (
                      <div 
                        className="rounded-circle bg-primary d-flex align-items-center justify-content-center text-white"
                        style={{ width: '80px', height: '80px', margin: '0 auto' }}
                      >
                        {userProfile.display_name?.charAt(0) || 'U'}
                      </div>
                    )}
                  </div>
                  <h6>{userProfile.display_name || 'Anonymous'}</h6>
                  <p className="text-muted small">{userProfile.bio || 'No bio available'}</p>
                  
                  {socialStats && (
                    <div className="d-flex justify-content-around mt-3">
                      <div className="text-center">
                        <div className="fw-bold">{socialStats.followers_count}</div>
                        <div className="small text-muted">Followers</div>
                      </div>
                      <div className="text-center">
                        <div className="fw-bold">{socialStats.following_count}</div>
                        <div className="small text-muted">Following</div>
                      </div>
                      <div className="text-center">
                        <div className="fw-bold">{socialStats.posts_count}</div>
                        <div className="small text-muted">Posts</div>
                      </div>
                    </div>
                  )}
                  
                  <Button 
                    variant="outline-primary" 
                    size="sm" 
                    className="mt-3 w-100"
                    onClick={() => setShowFollowModal(true)}
                  >
                    Edit Profile
                  </Button>
                </div>
              ) : (
                <Spinner animation="border" size="sm" />
              )}
            </Card.Body>
          </Card>

          {/* Quick Actions */}
          <Card>
            <Card.Header as="h5">🚀 Quick Actions</Card.Header>
            <Card.Body>
              <div className="d-grid gap-2">
                <Button 
                  variant="primary" 
                  size="sm"
                  onClick={() => setShowActivityModal(true)}
                >
                  📝 Create Activity
                </Button>
                <Button 
                  variant="outline-primary" 
                  size="sm"
                  onClick={() => setActiveTab('search')}
                >
                  🔍 Find Users
                </Button>
                <Button 
                  variant="outline-info" 
                  size="sm"
                  onClick={loadSocialStats}
                >
                  📊 Refresh Stats
                </Button>
              </div>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={9}>
          {/* Alerts */}
          {error && (
            <Alert variant="danger" dismissible onClose={clearMessages}>
              {error}
            </Alert>
          )}
          {success && (
            <Alert variant="success" dismissible onClose={clearMessages}>
              {success}
            </Alert>
          )}

          {/* Main Content Tabs */}
          <Card>
            <Card.Header>
              <Nav variant="tabs" activeKey={activeTab} onSelect={(k) => setActiveTab(k)}>
                <Nav.Item>
                  <Nav.Link eventKey="feed">📋 Activity Feed</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="followers">👥 Followers</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="following">🫶 Following</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="search">🔍 Search Users</Nav.Link>
                </Nav.Item>
              </Nav>
            </Card.Header>

            <Card.Body>
              <Tab.Content>
                {/* Activity Feed Tab */}
                <Tab.Pane active={activeTab === 'feed'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : activityFeed.length > 0 ? (
                    <ListGroup>
                      {activityFeed.map((activity, index) => (
                        <ListGroup.Item key={index} className="mb-2">
                          <div className="d-flex justify-content-between align-items-start">
                            <div>
                              <Badge bg="primary" className="me-2">
                                {activity.type}
                              </Badge>
                              <strong>{activity.user_display_name}</strong>
                              <p className="mb-1 mt-2">{activity.title}</p>
                              {activity.description && (
                                <small className="text-muted">{activity.description}</small>
                              )}
                            </div>
                            <small className="text-muted">
                              {new Date(activity.created_at).toLocaleDateString()}
                            </small>
                          </div>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">
                      No activities found. Start by creating your first activity!
                    </Alert>
                  )}
                </Tab.Pane>

                {/* Followers Tab */}
                <Tab.Pane active={activeTab === 'followers'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : followers.length > 0 ? (
                    <ListGroup>
                      {followers.map((follower, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-center">
                          <div>
                            <strong>{follower.display_name}</strong>
                            <br />
                            <small className="text-muted">{follower.username}</small>
                          </div>
                          <Button variant="outline-primary" size="sm">
                            View Profile
                          </Button>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">
                      No followers yet. Start connecting with other users!
                    </Alert>
                  )}
                </Tab.Pane>

                {/* Following Tab */}
                <Tab.Pane active={activeTab === 'following'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : following.length > 0 ? (
                    <ListGroup>
                      {following.map((user, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-center">
                          <div>
                            <strong>{user.display_name}</strong>
                            <br />
                            <small className="text-muted">{user.username}</small>
                          </div>
                          <Button 
                            variant="outline-danger" 
                            size="sm"
                            onClick={() => handleUnfollowUser(user.user_id)}
                          >
                            Unfollow
                          </Button>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">
                      You're not following anyone yet. Search for users to follow!
                    </Alert>
                  )}
                </Tab.Pane>

                {/* Search Users Tab */}
                <Tab.Pane active={activeTab === 'search'}>
                  <InputGroup className="mb-3">
                    <Form.Control
                      placeholder="Search for users..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && handleSearchUsers()}
                    />
                    <Button variant="primary" onClick={handleSearchUsers}>
                      Search
                    </Button>
                  </InputGroup>

                  {searchResults.length > 0 ? (
                    <ListGroup>
                      {searchResults.map((user, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-center">
                          <div>
                            <strong>{user.display_name}</strong>
                            <br />
                            <small className="text-muted">{user.username}</small>
                          </div>
                          <Button 
                            variant="primary" 
                            size="sm"
                            onClick={() => handleFollowUser(user.user_id)}
                          >
                            Follow
                          </Button>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : searchQuery ? (
                    <Alert variant="info">
                      No users found for "{searchQuery}"
                    </Alert>
                  ) : (
                    <Alert variant="info">
                      Enter a search query to find users
                    </Alert>
                  )}
                </Tab.Pane>
              </Tab.Content>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Edit Profile Modal */}
      <Modal show={showFollowModal} onHide={() => setShowFollowModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Edit Profile</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleUpdateProfile}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Display Name</Form.Label>
              <Form.Control
                type="text"
                value={profileForm.display_name}
                onChange={(e) => setProfileForm({...profileForm, display_name: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Bio</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={profileForm.bio}
                onChange={(e) => setProfileForm({...profileForm, bio: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Avatar URL</Form.Label>
              <Form.Control
                type="url"
                value={profileForm.avatar}
                onChange={(e) => setProfileForm({...profileForm, avatar: e.target.value})}
              />
            </Form.Group>
            <Form.Check
              type="checkbox"
              label="Private Profile"
              checked={profileForm.is_private}
              onChange={(e) => setProfileForm({...profileForm, is_private: e.target.checked})}
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowFollowModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Save Changes
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Create Activity Modal */}
      <Modal show={showActivityModal} onHide={() => setShowActivityModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Create Activity</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleCreateActivity}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Activity Type</Form.Label>
              <Form.Select
                value={activityForm.type}
                onChange={(e) => setActivityForm({...activityForm, type: e.target.value})}
              >
                <option value="watched">Watched</option>
                <option value="reviewed">Reviewed</option>
                <option value="rated">Rated</option>
                <option value="added_to_watchlist">Added to Watchlist</option>
              </Form.Select>
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Content ID</Form.Label>
              <Form.Control
                type="text"
                value={activityForm.content_id}
                onChange={(e) => setActivityForm({...activityForm, content_id: e.target.value})}
                placeholder="Movie ID or Content ID"
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Title</Form.Label>
              <Form.Control
                type="text"
                value={activityForm.title}
                onChange={(e) => setActivityForm({...activityForm, title: e.target.value})}
                placeholder="Activity title"
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={activityForm.description}
                onChange={(e) => setActivityForm({...activityForm, description: e.target.value})}
                placeholder="Activity description (optional)"
              />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowActivityModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Create Activity
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </Container>
  );
};

export default SocialFeatures;