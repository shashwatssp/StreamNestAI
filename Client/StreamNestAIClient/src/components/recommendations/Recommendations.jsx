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
  InputGroup,
  ListGroup,
  ProgressBar,
  Nav,
  Tab
} from 'react-bootstrap';
import { recommendationsApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

// Comprehensive logging utility
const logRecommendationFeature = (action, data = {}, level = 'info') => {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    feature: 'RECOMMENDATIONS',
    action,
    data,
    level,
    userId: localStorage.getItem('user_id') || 'anonymous'
  };
  
  const logMessage = `[${timestamp}] 🎯 [RECOMMENDATIONS] ${action}: ${JSON.stringify(data)}`;
  
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

const Recommendations = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('personalized');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // State for different recommendation types
  const [personalizedRecs, setPersonalizedRecs] = useState([]);
  const [trendingContent, setTrendingContent] = useState([]);
  const [similarMovies, setSimilarMovies] = useState([]);
  const [collaborativeRecs, setCollaborativeRecs] = useState([]);
  const [contentBasedRecs, setContentBasedRecs] = useState([]);
  const [genreBasedRecs, setGenreBasedRecs] = useState([]);
  const [moodBasedRecs, setMoodBasedRecs] = useState([]);
  const [seasonalRecs, setSeasonalRecs] = useState([]);

  // Filter and search states
  const [selectedGenre, setSelectedGenre] = useState('all');
  const [selectedMood, setSelectedMood] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedMovie, setSelectedMovie] = useState(null);

  // Modal states
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [showFeedbackModal, setShowFeedbackModal] = useState(false);
  const [selectedRecommendation, setSelectedRecommendation] = useState(null);

  // Feedback form
  const [feedbackForm, setFeedbackForm] = useState({
    recommendation_id: '',
    rating: 5,
    feedback_text: '',
    helpful: true
  });

  // Recommendation preferences
  const [preferences, setPreferences] = useState({
    preferred_genres: [],
    disliked_genres: [],
    preferred_decades: [],
    preferred_ratings: 7,
    enable_adult_content: false,
    recommendation_algorithm: 'hybrid'
  });

  useEffect(() => {
    logRecommendationFeature('COMPONENT_MOUNT', { userId: auth?.user_id, hasAuth: !!auth?.user_id });
    if (auth?.user_id) {
      logRecommendationFeature('INITIAL_LOAD_START', { userId: auth.user_id });
      loadPersonalizedRecommendations();
      loadTrendingContent();
      loadRecommendationPreferences();
    } else {
      logRecommendationFeature('NO_AUTH_USER', {}, 'warn');
    }
  }, [auth]);

  useEffect(() => {
    logRecommendationFeature('TAB_CHANGE', {
      activeTab,
      userId: auth?.user_id,
      selectedGenre,
      selectedMood
    });
    
    if (activeTab === 'trending') loadTrendingContent();
    if (activeTab === 'similar') loadSimilarMovies();
    if (activeTab === 'collaborative') loadCollaborativeRecommendations();
    if (activeTab === 'content-based') loadContentBasedRecommendations();
    if (activeTab === 'genre-based') loadGenreBasedRecommendations();
    if (activeTab === 'mood-based') loadMoodBasedRecommendations();
    if (activeTab === 'seasonal') loadSeasonalRecommendations();
  }, [activeTab, selectedGenre, selectedMood]);

  const loadPersonalizedRecommendations = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logRecommendationFeature('API_CALL_START', {
        endpoint: 'getPersonalizedRecommendations',
        userId: auth.user_id
      });
      
      const response = await recommendationsApi.getPersonalizedRecommendations();
      const duration = Date.now() - startTime;
      
      logRecommendationFeature('API_CALL_SUCCESS', {
        endpoint: 'getPersonalizedRecommendations',
        userId: auth.user_id,
        duration: `${duration}ms`,
        recommendationsCount: response.data?.recommendations?.length || 0,
        hasConfidenceScores: response.data?.recommendations?.some(r => r.confidence_score)
      });
      
      setPersonalizedRecs(response.data.recommendations || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logRecommendationFeature('API_CALL_ERROR', {
        endpoint: 'getPersonalizedRecommendations',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load personalized recommendations');
    } finally {
      setLoading(false);
      logRecommendationFeature('PERSONALIZED_RECS_LOAD_END', {
        userId: auth.user_id,
        count: personalizedRecs.length
      });
    }
  };

  const loadTrendingContent = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logRecommendationFeature('API_CALL_START', {
        endpoint: 'getTrendingContent',
        userId: auth.user_id
      });
      
      const response = await recommendationsApi.getTrendingContent();
      const duration = Date.now() - startTime;
      
      logRecommendationFeature('API_CALL_SUCCESS', {
        endpoint: 'getTrendingContent',
        userId: auth.user_id,
        duration: `${duration}ms`,
        contentCount: response.data?.content?.length || 0
      });
      
      setTrendingContent(response.data.content || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logRecommendationFeature('API_CALL_ERROR', {
        endpoint: 'getTrendingContent',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load trending content');
    } finally {
      setLoading(false);
      logRecommendationFeature('TRENDING_CONTENT_LOAD_END', {
        userId: auth.user_id,
        count: trendingContent.length
      });
    }
  };

  const loadSimilarMovies = async () => {
    if (!selectedMovie) return;
    
    try {
      setLoading(true);
      const response = await recommendationsApi.getSimilarMovies(selectedMovie.movie_id);
      setSimilarMovies(response.data.similar_movies || []);
    } catch (err) {
      setError('Failed to load similar movies');
    } finally {
      setLoading(false);
    }
  };

  const loadCollaborativeRecommendations = async () => {
    try {
      setLoading(true);
      const response = await recommendationsApi.getCollaborativeRecommendations();
      setCollaborativeRecs(response.data.recommendations || []);
    } catch (err) {
      setError('Failed to load collaborative recommendations');
    } finally {
      setLoading(false);
    }
  };

  const loadContentBasedRecommendations = async () => {
    try {
      setLoading(true);
      const response = await recommendationsApi.getContentBasedRecommendations();
      setContentBasedRecs(response.data.recommendations || []);
    } catch (err) {
      setError('Failed to load content-based recommendations');
    } finally {
      setLoading(false);
    }
  };

  const loadGenreBasedRecommendations = async () => {
    if (selectedGenre === 'all') return;
    
    try {
      setLoading(true);
      const response = await recommendationsApi.getGenreBasedRecommendations(selectedGenre);
      setGenreBasedRecs(response.data.recommendations || []);
    } catch (err) {
      setError('Failed to load genre-based recommendations');
    } finally {
      setLoading(false);
    }
  };

  const loadMoodBasedRecommendations = async () => {
    if (selectedMood === 'all') return;
    
    try {
      setLoading(true);
      const response = await recommendationsApi.getMoodBasedRecommendations(selectedMood);
      setMoodBasedRecs(response.data.recommendations || []);
    } catch (err) {
      setError('Failed to load mood-based recommendations');
    } finally {
      setLoading(false);
    }
  };

  const loadSeasonalRecommendations = async () => {
    try {
      setLoading(true);
      const response = await recommendationsApi.getSeasonalRecommendations();
      setSeasonalRecs(response.data.recommendations || []);
    } catch (err) {
      setError('Failed to load seasonal recommendations');
    } finally {
      setLoading(false);
    }
  };

  const loadRecommendationPreferences = async () => {
    try {
      const response = await recommendationsApi.getRecommendationPreferences();
      setPreferences(response.data.preferences || preferences);
    } catch (err) {
      console.error('Failed to load preferences:', err);
    }
  };

  const handleUpdatePreferences = async () => {
    const startTime = Date.now();
    try {
      logRecommendationFeature('UPDATE_PREFERENCES_START', {
        userId: auth.user_id,
        preferences: {
          algorithm: preferences.recommendation_algorithm,
          minRating: preferences.preferred_ratings,
          adultContent: preferences.enable_adult_content
        }
      });
      
      await recommendationsApi.updateRecommendationPreferences(preferences);
      const duration = Date.now() - startTime;
      
      logRecommendationFeature('UPDATE_PREFERENCES_SUCCESS', {
        userId: auth.user_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Preferences updated successfully!');
      loadPersonalizedRecommendations();
    } catch (err) {
      const duration = Date.now() - startTime;
      logRecommendationFeature('UPDATE_PREFERENCES_ERROR', {
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to update preferences');
    }
  };

  const handleProvideFeedback = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logRecommendationFeature('FEEDBACK_SUBMIT_START', {
        userId: auth.user_id,
        recommendationId: feedbackForm.recommendation_id,
        rating: feedbackForm.rating,
        helpful: feedbackForm.helpful,
        hasText: !!feedbackForm.feedback_text
      });
      
      await recommendationsApi.provideRecommendationFeedback(feedbackForm);
      const duration = Date.now() - startTime;
      
      logRecommendationFeature('FEEDBACK_SUBMIT_SUCCESS', {
        userId: auth.user_id,
        recommendationId: feedbackForm.recommendation_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Feedback submitted successfully!');
      setShowFeedbackModal(false);
      setFeedbackForm({
        recommendation_id: '',
        rating: 5,
        feedback_text: '',
        helpful: true
      });
    } catch (err) {
      const duration = Date.now() - startTime;
      logRecommendationFeature('FEEDBACK_SUBMIT_ERROR', {
        userId: auth.user_id,
        recommendationId: feedbackForm.recommendation_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to submit feedback');
    }
  };

  const handleSearchMovies = async () => {
    if (!searchQuery.trim()) return;
    
    try {
      setLoading(true);
      const response = await recommendationsApi.searchRecommendations(searchQuery);
      // Handle search results
    } catch (err) {
      setError('Failed to search recommendations');
    } finally {
      setLoading(false);
    }
  };

  const clearMessages = () => {
    logRecommendationFeature('CLEAR_MESSAGES', { userId: auth.user_id });
    setError('');
    setSuccess('');
  };

  const showMovieDetails = (movie) => {
    logRecommendationFeature('MOVIE_DETAILS_OPEN', {
      userId: auth.user_id,
      movieId: movie.movie_id,
      title: movie.title,
      hasConfidenceScore: !!movie.confidence_score
    });
    setSelectedRecommendation(movie);
    setShowDetailsModal(true);
  };

  const showFeedbackDialog = (recommendation) => {
    logRecommendationFeature('FEEDBACK_DIALOG_OPEN', {
      userId: auth.user_id,
      recommendationId: recommendation.recommendation_id || recommendation.movie_id,
      title: recommendation.title
    });
    setSelectedRecommendation(recommendation);
    setFeedbackForm({
      ...feedbackForm,
      recommendation_id: recommendation.recommendation_id || recommendation.movie_id
    });
    setShowFeedbackModal(true);
  };

  const getConfidenceBadge = (confidence) => {
    if (confidence >= 0.8) return <Badge bg="success">High Confidence</Badge>;
    if (confidence >= 0.6) return <Badge bg="warning">Medium Confidence</Badge>;
    return <Badge bg="secondary">Low Confidence</Badge>;
  };

  const renderRecommendationCard = (item, index) => (
    <Col md={4} className="mb-4" key={index}>
      <Card className="h-100">
        <Card.Img 
          variant="top" 
          src={item.poster_url || '/placeholder-movie.jpg'} 
          style={{ height: '200px', objectFit: 'cover' }}
        />
        <Card.Body>
          <div className="d-flex justify-content-between align-items-start mb-2">
            <Card.Title className="h6">{item.title}</Card.Title>
            {item.confidence_score && getConfidenceBadge(item.confidence_score)}
          </div>
          <Card.Text className="small text-muted mb-2">
            {item.genre} • {item.year} • {item.duration}
          </Card.Text>
          <Card.Text className="small">
            {item.description?.substring(0, 100)}...
          </Card.Text>
          <div className="d-flex justify-content-between align-items-center mb-2">
            <Badge bg="warning">⭐ {item.rating?.toFixed(1)}</Badge>
            <small className="text-muted">
              {item.match_score && `Match: ${Math.round(item.match_score * 100)}%`}
            </small>
          </div>
          {item.reason && (
            <Alert variant="info" className="small py-2 mb-2">
              <strong>Why recommended:</strong> {item.reason}
            </Alert>
          )}
        </Card.Body>
        <Card.Footer className="bg-transparent">
          <div className="d-grid gap-2">
            <Button 
              variant="outline-primary" 
              size="sm"
              onClick={() => showMovieDetails(item)}
            >
              📋 Details
            </Button>
            <Button 
              variant="outline-info" 
              size="sm"
              onClick={() => showFeedbackDialog(item)}
            >
              💬 Feedback
            </Button>
          </div>
        </Card.Footer>
      </Card>
    </Col>
  );

  return (
    <Container fluid className="py-4">
      <Row>
        <Col lg={3}>
          {/* Preferences Card */}
          <Card className="mb-4">
            <Card.Header as="h5">⚙️ Preferences</Card.Header>
            <Card.Body>
              <Form.Group className="mb-3">
                <Form.Label>Algorithm</Form.Label>
                <Form.Select
                  value={preferences.recommendation_algorithm}
                  onChange={(e) => setPreferences({...preferences, recommendation_algorithm: e.target.value})}
                >
                  <option value="hybrid">Hybrid</option>
                  <option value="collaborative">Collaborative</option>
                  <option value="content-based">Content-Based</option>
                  <option value="deep-learning">Deep Learning</option>
                </Form.Select>
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Min Rating</Form.Label>
                <Form.Range
                  min="1"
                  max="10"
                  value={preferences.preferred_ratings}
                  onChange={(e) => setPreferences({...preferences, preferred_ratings: parseInt(e.target.value)})}
                />
                <small className="text-muted">{preferences.preferred_ratings}+ stars</small>
              </Form.Group>

              <Form.Check
                type="checkbox"
                label="Include adult content"
                checked={preferences.enable_adult_content}
                onChange={(e) => setPreferences({...preferences, enable_adult_content: e.target.checked})}
                className="mb-3"
              />

              <Button 
                variant="primary" 
                size="sm" 
                className="w-100"
                onClick={handleUpdatePreferences}
              >
                Update Preferences
              </Button>
            </Card.Body>
          </Card>

          {/* Quick Filters */}
          <Card className="mb-4">
            <Card.Header as="h5">🎭 Genre Filter</Card.Header>
            <Card.Body>
              <Form.Select
                value={selectedGenre}
                onChange={(e) => setSelectedGenre(e.target.value)}
                className="mb-2"
              >
                <option value="all">All Genres</option>
                <option value="action">Action</option>
                <option value="comedy">Comedy</option>
                <option value="drama">Drama</option>
                <option value="horror">Horror</option>
                <option value="romance">Romance</option>
                <option value="sci-fi">Sci-Fi</option>
                <option value="thriller">Thriller</option>
              </Form.Select>
            </Card.Body>
          </Card>

          <Card>
            <Card.Header as="h5">😊 Mood Filter</Card.Header>
            <Card.Body>
              <Form.Select
                value={selectedMood}
                onChange={(e) => setSelectedMood(e.target.value)}
              >
                <option value="all">All Moods</option>
                <option value="happy">Happy</option>
                <option value="sad">Sad</option>
                <option value="excited">Excited</option>
                <option value="relaxed">Relaxed</option>
                <option value="thoughtful">Thoughtful</option>
                <option value="adventurous">Adventurous</option>
              </Form.Select>
            </Card.Body>
          </Card>
        </Col>

        <Col lg={9}>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>🎯 Recommendations</h2>
            <div className="d-flex align-items-center">
              <InputGroup style={{ maxWidth: '300px' }} className="me-2">
                <Form.Control
                  placeholder="Search recommendations..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearchMovies()}
                />
                <Button variant="outline-primary" onClick={handleSearchMovies}>
                  Search
                </Button>
              </InputGroup>
              <Button variant="outline-info" onClick={loadPersonalizedRecommendations}>
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
          {success && (
            <Alert variant="success" dismissible onClose={clearMessages}>
              {success}
            </Alert>
          )}

          {/* Recommendation Tabs */}
          <Card>
            <Card.Header>
              <Nav variant="tabs" activeKey={activeTab} onSelect={(k) => setActiveTab(k)}>
                <Nav.Item>
                  <Nav.Link eventKey="personalized">🎯 For You</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="trending">🔥 Trending</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="similar">🎬 Similar</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="collaborative">👥 Users Like You</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="content-based">📊 Content-Based</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="genre-based">🎭 Genre-Based</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="mood-based">😊 Mood-Based</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="seasonal">🌟 Seasonal</Nav.Link>
                </Nav.Item>
              </Nav>
            </Card.Header>

            <Card.Body>
              {loading ? (
                <div className="text-center py-4">
                  <Spinner animation="border" />
                  <p className="mt-2">Loading recommendations...</p>
                </div>
              ) : (
                <Tab.Content>
                  {/* Personalized Recommendations */}
                  <Tab.Pane active={activeTab === 'personalized'}>
                    {personalizedRecs.length > 0 ? (
                      <Row>
                        {personalizedRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">
                        No personalized recommendations available. Watch more movies to get better recommendations!
                      </Alert>
                    )}
                  </Tab.Pane>

                  {/* Trending Content */}
                  <Tab.Pane active={activeTab === 'trending'}>
                    {trendingContent.length > 0 ? (
                      <Row>
                        {trendingContent.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No trending content available</Alert>
                    )}
                  </Tab.Pane>

                  {/* Similar Movies */}
                  <Tab.Pane active={activeTab === 'similar'}>
                    {!selectedMovie ? (
                      <Alert variant="info">
                        Select a movie to see similar recommendations
                      </Alert>
                    ) : similarMovies.length > 0 ? (
                      <Row>
                        {similarMovies.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No similar movies found</Alert>
                    )}
                  </Tab.Pane>

                  {/* Collaborative Recommendations */}
                  <Tab.Pane active={activeTab === 'collaborative'}>
                    {collaborativeRecs.length > 0 ? (
                      <Row>
                        {collaborativeRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">
                        No collaborative recommendations available. Rate more movies to get better suggestions!
                      </Alert>
                    )}
                  </Tab.Pane>

                  {/* Content-Based Recommendations */}
                  <Tab.Pane active={activeTab === 'content-based'}>
                    {contentBasedRecs.length > 0 ? (
                      <Row>
                        {contentBasedRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No content-based recommendations available</Alert>
                    )}
                  </Tab.Pane>

                  {/* Genre-Based Recommendations */}
                  <Tab.Pane active={activeTab === 'genre-based'}>
                    {selectedGenre === 'all' ? (
                      <Alert variant="info">Select a genre to see recommendations</Alert>
                    ) : genreBasedRecs.length > 0 ? (
                      <Row>
                        {genreBasedRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No recommendations available for this genre</Alert>
                    )}
                  </Tab.Pane>

                  {/* Mood-Based Recommendations */}
                  <Tab.Pane active={activeTab === 'mood-based'}>
                    {selectedMood === 'all' ? (
                      <Alert variant="info">Select a mood to see recommendations</Alert>
                    ) : moodBasedRecs.length > 0 ? (
                      <Row>
                        {moodBasedRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No recommendations available for this mood</Alert>
                    )}
                  </Tab.Pane>

                  {/* Seasonal Recommendations */}
                  <Tab.Pane active={activeTab === 'seasonal'}>
                    {seasonalRecs.length > 0 ? (
                      <Row>
                        {seasonalRecs.map((item, index) => renderRecommendationCard(item, index))}
                      </Row>
                    ) : (
                      <Alert variant="info">No seasonal recommendations available</Alert>
                    )}
                  </Tab.Pane>
                </Tab.Content>
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Movie Details Modal */}
      <Modal show={showDetailsModal} onHide={() => setShowDetailsModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>{selectedRecommendation?.title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {selectedRecommendation && (
            <Row>
              <Col md={4}>
                <img 
                  src={selectedRecommendation.poster_url || '/placeholder-movie.jpg'} 
                  alt={selectedRecommendation.title}
                  className="img-fluid rounded"
                />
              </Col>
              <Col md={8}>
                <h5>{selectedRecommendation.title}</h5>
                <p className="text-muted">
                  {selectedRecommendation.year} • {selectedRecommendation.genre} • {selectedRecommendation.duration}
                </p>
                <p><strong>Director:</strong> {selectedRecommendation.director}</p>
                <p><strong>Cast:</strong> {selectedRecommendation.cast}</p>
                <p><strong>Rating:</strong> ⭐ {selectedRecommendation.rating}/10</p>
                <p><strong>Description:</strong> {selectedRecommendation.description}</p>
                {selectedRecommendation.reason && (
                  <Alert variant="info">
                    <strong>Why recommended:</strong> {selectedRecommendation.reason}
                  </Alert>
                )}
                {selectedRecommendation.confidence_score && (
                  <div className="mb-3">
                    <strong>Confidence:</strong>
                    <ProgressBar 
                      now={selectedRecommendation.confidence_score * 100} 
                      label={`${Math.round(selectedRecommendation.confidence_score * 100)}%`}
                      variant="info"
                    />
                  </div>
                )}
              </Col>
            </Row>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowDetailsModal(false)}>
            Close
          </Button>
          <Button variant="primary">
            Watch Now
          </Button>
        </Modal.Footer>
      </Modal>

      {/* Feedback Modal */}
      <Modal show={showFeedbackModal} onHide={() => setShowFeedbackModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Provide Feedback</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleProvideFeedback}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Rating</Form.Label>
              <Form.Range
                min="1"
                max="5"
                value={feedbackForm.rating}
                onChange={(e) => setFeedbackForm({...feedbackForm, rating: parseInt(e.target.value)})}
              />
              <div className="d-flex justify-content-between">
                <small>Not Helpful</small>
                <small>Very Helpful</small>
              </div>
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Feedback (Optional)</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={feedbackForm.feedback_text}
                onChange={(e) => setFeedbackForm({...feedbackForm, feedback_text: e.target.value})}
                placeholder="Tell us more about this recommendation..."
              />
            </Form.Group>
            <Form.Check
              type="checkbox"
              label="This recommendation was helpful"
              checked={feedbackForm.helpful}
              onChange={(e) => setFeedbackForm({...feedbackForm, helpful: e.target.checked})}
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowFeedbackModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Submit Feedback
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </Container>
  );
};

export default Recommendations;