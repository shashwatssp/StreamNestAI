import React, { useState, useEffect } from 'react';
import {
  Container,
  Row,
  Col,
  Card,
  Button,
  Form,
  Badge,
  Alert,
  Spinner,
  Modal,
  ListGroup,
  ProgressBar,
  Tab,
  Nav
} from 'react-bootstrap';
import { socialApi, recommendationsApi, analyticsApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

// Comprehensive logging utility
const logProfileFeature = (action, data = {}, level = 'info') => {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    feature: 'ENHANCED_PROFILE',
    action,
    data,
    level,
    userId: localStorage.getItem('user_id') || 'anonymous'
  };
  
  const logMessage = `[${timestamp}] 👤 [ENHANCED_PROFILE] ${action}: ${JSON.stringify(data)}`;
  
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

const EnhancedProfile = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // State for user profile data
  const [userProfile, setUserProfile] = useState(null);
  const [socialStats, setSocialStats] = useState(null);
  const [userPreferences, setUserPreferences] = useState(null);
  const [userActivity, setUserActivity] = useState([]);
  const [userStats, setUserStats] = useState(null);
  const [achievements, setAchievements] = useState([]);
  const [watchHistory, setWatchHistory] = useState([]);
  const [userReviews, setUserReviews] = useState([]);

  // Modal states
  const [showEditProfileModal, setShowEditProfileModal] = useState(false);
  const [showPreferencesModal, setShowPreferencesModal] = useState(false);
  const [showAchievementModal, setShowAchievementModal] = useState(false);

  // Form states
  const [profileForm, setProfileForm] = useState({
    display_name: '',
    bio: '',
    avatar: '',
    location: '',
    website: '',
    birth_date: '',
    is_private: false
  });
  const [preferencesForm, setPreferencesForm] = useState({
    preferred_genres: [],
    disliked_genres: [],
    preferred_decades: [],
    language: 'en',
    timezone: 'UTC',
    email_notifications: true,
    push_notifications: false,
    public_profile: true,
    show_watch_history: true
  });

  useEffect(() => {
    logProfileFeature('COMPONENT_MOUNT', { userId: auth?.user_id, hasAuth: !!auth?.user_id });
    if (auth?.user_id) {
      logProfileFeature('INITIAL_LOAD_START', { userId: auth.user_id });
      loadUserProfile();
      loadSocialStats();
      loadUserPreferences();
      loadUserActivity();
      loadUserStats();
      loadAchievements();
      loadWatchHistory();
      loadUserReviews();
    } else {
      logProfileFeature('NO_AUTH_USER', {}, 'warn');
    }
  }, [auth]);

  const loadUserProfile = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logProfileFeature('API_CALL_START', {
        endpoint: 'getUserProfile',
        userId: auth.user_id
      });
      
      const response = await socialApi.getUserProfile(auth.user_id);
      const duration = Date.now() - startTime;
      
      logProfileFeature('API_CALL_SUCCESS', {
        endpoint: 'getUserProfile',
        userId: auth.user_id,
        duration: `${duration}ms`,
        hasAvatar: !!response.data.avatar,
        hasBio: !!response.data.bio,
        isPrivate: response.data.is_private
      });
      
      setUserProfile(response.data);
      setProfileForm({
        display_name: response.data.display_name || '',
        bio: response.data.bio || '',
        avatar: response.data.avatar || '',
        location: response.data.location || '',
        website: response.data.website || '',
        birth_date: response.data.birth_date || '',
        is_private: response.data.is_private || false
      });
    } catch (err) {
      const duration = Date.now() - startTime;
      logProfileFeature('API_CALL_ERROR', {
        endpoint: 'getUserProfile',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load user profile');
    } finally {
      setLoading(false);
    }
  };

  const loadSocialStats = async () => {
    const startTime = Date.now();
    try {
      logProfileFeature('API_CALL_START', {
        endpoint: 'getSocialStats',
        userId: auth?.user_id
      });
      
      const response = await socialApi.getSocialStats();
      const duration = Date.now() - startTime;
      
      logProfileFeature('API_CALL_SUCCESS', {
        endpoint: 'getSocialStats',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        followersCount: response.data.stats?.followers_count,
        followingCount: response.data.stats?.following_count,
        postsCount: response.data.stats?.posts_count
      });
      
      setSocialStats(response.data.stats);
    } catch (err) {
      const duration = Date.now() - startTime;
      logProfileFeature('API_CALL_ERROR', {
        endpoint: 'getSocialStats',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
    }
  };

  const loadUserPreferences = async () => {
    const startTime = Date.now();
    try {
      logProfileFeature('API_CALL_START', {
        endpoint: 'getRecommendationPreferences',
        userId: auth?.user_id
      });
      
      const response = await recommendationsApi.getRecommendationPreferences();
      const duration = Date.now() - startTime;
      
      logProfileFeature('API_CALL_SUCCESS', {
        endpoint: 'getRecommendationPreferences',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        hasPreferences: !!response.data.preferences,
        preferredGenresCount: response.data.preferences?.preferred_genres?.length || 0,
        emailNotifications: response.data.preferences?.email_notifications
      });
      
      setUserPreferences(response.data.preferences);
      setPreferencesForm(response.data.preferences || preferencesForm);
    } catch (err) {
      const duration = Date.now() - startTime;
      logProfileFeature('API_CALL_ERROR', {
        endpoint: 'getRecommendationPreferences',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
    }
  };

  const loadUserActivity = async () => {
    try {
      const response = await analyticsApi.getUserActivity('30d');
      setUserActivity(response.data.activities || []);
    } catch (err) {
      console.error('Failed to load user activity:', err);
    }
  };

  const loadUserStats = async () => {
    try {
      const response = await analyticsApi.getUserAnalytics('30d');
      setUserStats(response.data.analytics);
    } catch (err) {
      console.error('Failed to load user stats:', err);
    }
  };

  const loadAchievements = async () => {
    try {
      // Mock achievements data - would come from API
      setAchievements([
        { id: 1, name: 'Movie Buff', description: 'Watched 100 movies', icon: '🎬', unlocked: true, progress: 100 },
        { id: 2, name: 'Reviewer', description: 'Written 50 reviews', icon: '⭐', unlocked: true, progress: 100 },
        { id: 3, name: 'Social Butterfly', description: 'Followed 50 users', icon: '🦋', unlocked: false, progress: 60 },
        { id: 4, name: 'Explorer', description: 'Watched 10 different genres', icon: '🗺️', unlocked: true, progress: 100 },
        { id: 5, name: 'Trendsetter', description: 'Created 5 popular watchlists', icon: '🔥', unlocked: false, progress: 40 }
      ]);
    } catch (err) {
      console.error('Failed to load achievements:', err);
    }
  };

  const loadWatchHistory = async () => {
    try {
      // Mock watch history - would come from API
      setWatchHistory([
        { movie_id: '1', title: 'The Shawshank Redemption', watched_at: '2024-01-15', rating: 5 },
        { movie_id: '2', title: 'The Godfather', watched_at: '2024-01-14', rating: 5 },
        { movie_id: '3', title: 'The Dark Knight', watched_at: '2024-01-13', rating: 4 }
      ]);
    } catch (err) {
      console.error('Failed to load watch history:', err);
    }
  };

  const loadUserReviews = async () => {
    try {
      // Mock user reviews - would come from API
      setUserReviews([
        { review_id: '1', movie_title: 'Inception', rating: 5, review_text: 'Amazing movie!', created_at: '2024-01-10', likes: 12 },
        { review_id: '2', movie_title: 'Interstellar', rating: 4, review_text: 'Great visuals', created_at: '2024-01-08', likes: 8 }
      ]);
    } catch (err) {
      console.error('Failed to load user reviews:', err);
    }
  };

  const handleUpdateProfile = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logProfileFeature('UPDATE_PROFILE_START', {
        userId: auth.user_id,
        profileData: {
          hasDisplayName: !!profileForm.display_name,
          hasBio: !!profileForm.bio,
          hasAvatar: !!profileForm.avatar,
          hasLocation: !!profileForm.location,
          hasWebsite: !!profileForm.website,
          isPrivate: profileForm.is_private
        }
      });
      
      await socialApi.updateProfile(profileForm);
      const duration = Date.now() - startTime;
      
      logProfileFeature('UPDATE_PROFILE_SUCCESS', {
        userId: auth.user_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Profile updated successfully!');
      setShowEditProfileModal(false);
      loadUserProfile();
    } catch (err) {
      const duration = Date.now() - startTime;
      logProfileFeature('UPDATE_PROFILE_ERROR', {
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to update profile');
    }
  };

  const handleUpdatePreferences = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logProfileFeature('UPDATE_PREFERENCES_START', {
        userId: auth.user_id,
        preferencesData: {
          preferredGenresCount: preferencesForm.preferred_genres?.length || 0,
          dislikedGenresCount: preferencesForm.disliked_genres?.length || 0,
          emailNotifications: preferencesForm.email_notifications,
          pushNotifications: preferencesForm.push_notifications,
          publicProfile: preferencesForm.public_profile,
          showWatchHistory: preferencesForm.show_watch_history
        }
      });
      
      await recommendationsApi.updateRecommendationPreferences(preferencesForm);
      const duration = Date.now() - startTime;
      
      logProfileFeature('UPDATE_PREFERENCES_SUCCESS', {
        userId: auth.user_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Preferences updated successfully!');
      setShowPreferencesModal(false);
      loadUserPreferences();
    } catch (err) {
      const duration = Date.now() - startTime;
      logProfileFeature('UPDATE_PREFERENCES_ERROR', {
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to update preferences');
    }
  };

  const clearMessages = () => {
    logProfileFeature('CLEAR_MESSAGES', { userId: auth?.user_id });
    setError('');
    setSuccess('');
  };

  const getLevelBadge = (level) => {
    const colors = ['success', 'info', 'warning', 'danger'];
    return <Badge bg={colors[Math.min(level - 1, 3)]}>Level {level}</Badge>;
  };

  const getAchievementBadge = (unlocked) => {
    return <Badge bg={unlocked ? 'success' : 'secondary'}>
      {unlocked ? '✓ Unlocked' : '🔒 Locked'}
    </Badge>;
  };

  return (
    <Container fluid className="py-4">
      <Row>
        <Col lg={4}>
          {/* Profile Card */}
          <Card className="mb-4">
            <Card.Body className="text-center">
              {userProfile ? (
                <>
                  <div className="mb-3">
                    {userProfile.avatar ? (
                      <img 
                        src={userProfile.avatar} 
                        alt="Avatar" 
                        className="rounded-circle"
                        style={{ width: '120px', height: '120px', objectFit: 'cover' }}
                      />
                    ) : (
                      <div 
                        className="rounded-circle bg-primary d-flex align-items-center justify-content-center text-white mx-auto"
                        style={{ width: '120px', height: '120px', fontSize: '48px' }}
                      >
                        {userProfile.display_name?.charAt(0) || 'U'}
                      </div>
                    )}
                  </div>
                  <h4>{userProfile.display_name || 'Anonymous'}</h4>
                  <p className="text-muted">@{userProfile.username}</p>
                  <p className="small">{userProfile.bio || 'No bio available'}</p>
                  
                  {userProfile.location && (
                    <p className="small text-muted">
                      📍 {userProfile.location}
                    </p>
                  )}
                  
                  {userProfile.website && (
                    <p className="small">
                      <a href={userProfile.website} target="_blank" rel="noopener noreferrer">
                        🌐 {userProfile.website}
                      </a>
                    </p>
                  )}

                  <div className="d-flex justify-content-center gap-2 mb-3">
                    {userStats && getLevelBadge(userStats.level || 1)}
                    {userProfile.is_private && <Badge bg="warning">Private</Badge>}
                  </div>

                  <div className="d-grid gap-2">
                    <Button 
                      variant="primary" 
                      onClick={() => setShowEditProfileModal(true)}
                    >
                      Edit Profile
                    </Button>
                    <Button 
                      variant="outline-info" 
                      onClick={() => setShowPreferencesModal(true)}
                    >
                      Preferences
                    </Button>
                  </div>
                </>
              ) : (
                <Spinner animation="border" />
              )}
            </Card.Body>
          </Card>

          {/* Social Stats */}
          {socialStats && (
            <Card className="mb-4">
              <Card.Header as="h6">📊 Social Stats</Card.Header>
              <Card.Body>
                <div className="d-flex justify-content-around text-center">
                  <div>
                    <div className="fw-bold">{socialStats.followers_count}</div>
                    <div className="small text-muted">Followers</div>
                  </div>
                  <div>
                    <div className="fw-bold">{socialStats.following_count}</div>
                    <div className="small text-muted">Following</div>
                  </div>
                  <div>
                    <div className="fw-bold">{socialStats.posts_count}</div>
                    <div className="small text-muted">Posts</div>
                  </div>
                </div>
              </Card.Body>
            </Card>
          )}

          {/* Achievements */}
          <Card>
            <Card.Header as="h6">🏆 Achievements</Card.Header>
            <Card.Body>
              <div className="d-grid gap-2">
                {achievements.slice(0, 3).map((achievement) => (
                  <div key={achievement.id} className="d-flex justify-content-between align-items-center">
                    <div>
                      <span className="me-2">{achievement.icon}</span>
                      <small>{achievement.name}</small>
                    </div>
                    {getAchievementBadge(achievement.unlocked)}
                  </div>
                ))}
              </div>
              <Button 
                variant="outline-primary" 
                size="sm" 
                className="w-100 mt-2"
                onClick={() => setShowAchievementModal(true)}
              >
                View All
              </Button>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={8}>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>👤 My Profile</h2>
            <Button variant="outline-info" onClick={loadUserProfile}>
              🔄 Refresh
            </Button>
          </div>

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
                  <Nav.Link eventKey="overview">📋 Overview</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="activity">⚡ Activity</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="history">📺 Watch History</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="reviews">⭐ Reviews</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="stats">📈 Statistics</Nav.Link>
                </Nav.Item>
              </Nav>
            </Card.Header>

            <Card.Body>
              <Tab.Content>
                {/* Overview Tab */}
                <Tab.Pane active={activeTab === 'overview'}>
                  {userStats ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Viewing Statistics</Card.Header>
                          <Card.Body>
                            <p><strong>Total Movies Watched:</strong> {userStats.total_movies_watched || 0}</p>
                            <p><strong>Total Watch Time:</strong> {userStats.total_watch_time || 0} hours</p>
                            <p><strong>Average Rating:</strong> ⭐ {userStats.average_rating?.toFixed(1) || 'N/A'}</p>
                            <p><strong>Favorite Genre:</strong> {userStats.favorite_genre || 'N/A'}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Contribution Statistics</Card.Header>
                          <Card.Body>
                            <p><strong>Reviews Written:</strong> {userStats.total_reviews || 0}</p>
                            <p><strong>Watchlists Created:</strong> {userStats.watchlists_created || 0}</p>
                            <p><strong>Followers:</strong> {socialStats?.followers_count || 0}</p>
                            <p><strong>Profile Views:</strong> {userStats.profile_views || 0}</p>
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">Loading statistics...</Alert>
                  )}
                </Tab.Pane>

                {/* Activity Tab */}
                <Tab.Pane active={activeTab === 'activity'}>
                  {userActivity.length > 0 ? (
                    <ListGroup>
                      {userActivity.map((activity, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-start">
                          <div>
                            <Badge bg="primary" className="me-2">
                              {activity.activity_type}
                            </Badge>
                            <strong>{activity.content_title}</strong>
                            <p className="mb-1 mt-2 small">{activity.description}</p>
                          </div>
                          <small className="text-muted">
                            {new Date(activity.timestamp).toLocaleDateString()}
                          </small>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">No recent activity</Alert>
                  )}
                </Tab.Pane>

                {/* Watch History Tab */}
                <Tab.Pane active={activeTab === 'history'}>
                  {watchHistory.length > 0 ? (
                    <ListGroup>
                      {watchHistory.map((movie, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-center">
                          <div>
                            <strong>{movie.title}</strong>
                            <br />
                            <small className="text-muted">
                              Watched on {new Date(movie.watched_at).toLocaleDateString()}
                            </small>
                          </div>
                          <Badge bg="warning">⭐ {movie.rating}</Badge>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">No watch history available</Alert>
                  )}
                </Tab.Pane>

                {/* Reviews Tab */}
                <Tab.Pane active={activeTab === 'reviews'}>
                  {userReviews.length > 0 ? (
                    <ListGroup>
                      {userReviews.map((review, index) => (
                        <ListGroup.Item key={index}>
                          <div className="d-flex justify-content-between align-items-start mb-2">
                            <h6>{review.movie_title}</h6>
                            <Badge bg="warning">⭐ {review.rating}</Badge>
                          </div>
                          <p className="mb-2">{review.review_text}</p>
                          <div className="d-flex justify-content-between align-items-center">
                            <small className="text-muted">
                              {new Date(review.created_at).toLocaleDateString()}
                            </small>
                            <small className="text-muted">
                              ❤️ {review.likes} likes
                            </small>
                          </div>
                        </ListGroup.Item>
                      ))}
                    </ListGroup>
                  ) : (
                    <Alert variant="info">No reviews written yet</Alert>
                  )}
                </Tab.Pane>

                {/* Statistics Tab */}
                <Tab.Pane active={activeTab === 'stats'}>
                  {userStats ? (
                    <Row>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Genre Distribution</Card.Header>
                          <Card.Body>
                            {userStats.genre_distribution ? (
                              Object.entries(userStats.genre_distribution).map(([genre, count]) => (
                                <div key={genre} className="mb-2">
                                  <div className="d-flex justify-content-between">
                                    <small>{genre}</small>
                                    <small>{count}</small>
                                  </div>
                                  <ProgressBar 
                                    now={(count / userStats.total_movies_watched) * 100} 
                                    variant="info"
                                    style={{ height: '5px' }}
                                  />
                                </div>
                              ))
                            ) : (
                              <p className="text-muted">No genre data available</p>
                            )}
                          </Card.Body>
                        </Card>
                      </Col>
                      <Col md={6}>
                        <Card className="mb-3">
                          <Card.Header as="h6">Rating Distribution</Card.Header>
                          <Card.Body>
                            {userStats.rating_distribution ? (
                              Object.entries(userStats.rating_distribution).map(([rating, count]) => (
                                <div key={rating} className="mb-2">
                                  <div className="d-flex justify-content-between">
                                    <small>⭐ {rating} stars</small>
                                    <small>{count}</small>
                                  </div>
                                  <ProgressBar 
                                    now={(count / userStats.total_reviews) * 100} 
                                    variant="warning"
                                    style={{ height: '5px' }}
                                  />
                                </div>
                              ))
                            ) : (
                              <p className="text-muted">No rating data available</p>
                            )}
                          </Card.Body>
                        </Card>
                      </Col>
                    </Row>
                  ) : (
                    <Alert variant="info">Loading statistics...</Alert>
                  )}
                </Tab.Pane>
              </Tab.Content>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Edit Profile Modal */}
      <Modal show={showEditProfileModal} onHide={() => setShowEditProfileModal(false)}>
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
            <Form.Group className="mb-3">
              <Form.Label>Location</Form.Label>
              <Form.Control
                type="text"
                value={profileForm.location}
                onChange={(e) => setProfileForm({...profileForm, location: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Website</Form.Label>
              <Form.Control
                type="url"
                value={profileForm.website}
                onChange={(e) => setProfileForm({...profileForm, website: e.target.value})}
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
            <Button variant="secondary" onClick={() => setShowEditProfileModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Save Changes
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Preferences Modal */}
      <Modal show={showPreferencesModal} onHide={() => setShowPreferencesModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Preferences</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleUpdatePreferences}>
          <Modal.Body>
            <Row>
              <Col md={6}>
                <h6>Content Preferences</h6>
                <Form.Group className="mb-3">
                  <Form.Label>Preferred Genres</Form.Label>
                  <Form.Control
                    type="text"
                    value={preferencesForm.preferred_genres.join(', ')}
                    onChange={(e) => setPreferencesForm({
                      ...preferencesForm, 
                      preferred_genres: e.target.value.split(',').map(g => g.trim()).filter(g => g)
                    })}
                    placeholder="action, comedy, drama"
                  />
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>Disliked Genres</Form.Label>
                  <Form.Control
                    type="text"
                    value={preferencesForm.disliked_genres.join(', ')}
                    onChange={(e) => setPreferencesForm({
                      ...preferencesForm, 
                      disliked_genres: e.target.value.split(',').map(g => g.trim()).filter(g => g)
                    })}
                    placeholder="horror, thriller"
                  />
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>Minimum Rating</Form.Label>
                  <Form.Range
                    min="1"
                    max="10"
                    value={preferencesForm.preferred_ratings || 7}
                    onChange={(e) => setPreferencesForm({
                      ...preferencesForm, 
                      preferred_ratings: parseInt(e.target.value)
                    })}
                  />
                  <small className="text-muted">{preferencesForm.preferred_ratings || 7}+ stars</small>
                </Form.Group>
              </Col>
              <Col md={6}>
                <h6>Notification Settings</h6>
                <Form.Check
                  type="checkbox"
                  label="Email Notifications"
                  checked={preferencesForm.email_notifications}
                  onChange={(e) => setPreferencesForm({
                    ...preferencesForm, 
                    email_notifications: e.target.checked
                  })}
                  className="mb-2"
                />
                <Form.Check
                  type="checkbox"
                  label="Push Notifications"
                  checked={preferencesForm.push_notifications}
                  onChange={(e) => setPreferencesForm({
                    ...preferencesForm, 
                    push_notifications: e.target.checked
                  })}
                  className="mb-2"
                />
                <Form.Check
                  type="checkbox"
                  label="Public Profile"
                  checked={preferencesForm.public_profile}
                  onChange={(e) => setPreferencesForm({
                    ...preferencesForm, 
                    public_profile: e.target.checked
                  })}
                  className="mb-2"
                />
                <Form.Check
                  type="checkbox"
                  label="Show Watch History"
                  checked={preferencesForm.show_watch_history}
                  onChange={(e) => setPreferencesForm({
                    ...preferencesForm, 
                    show_watch_history: e.target.checked
                  })}
                />
              </Col>
            </Row>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowPreferencesModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Save Preferences
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Achievements Modal */}
      <Modal show={showAchievementModal} onHide={() => setShowAchievementModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>🏆 Achievements</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="d-grid gap-3">
            {achievements.map((achievement) => (
              <Card key={achievement.id}>
                <Card.Body className="d-flex justify-content-between align-items-center">
                  <div>
                    <div className="d-flex align-items-center mb-2">
                      <span className="me-2" style={{ fontSize: '24px' }}>{achievement.icon}</span>
                      <h6 className="mb-0">{achievement.name}</h6>
                    </div>
                    <p className="small text-muted mb-2">{achievement.description}</p>
                    {!achievement.unlocked && (
                      <ProgressBar now={achievement.progress} variant="info" style={{ height: '5px' }} />
                    )}
                  </div>
                  {getAchievementBadge(achievement.unlocked)}
                </Card.Body>
              </Card>
            ))}
          </div>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowAchievementModal(false)}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
};

export default EnhancedProfile;