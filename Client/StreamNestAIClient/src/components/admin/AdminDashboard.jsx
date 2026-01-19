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
  Modal,
  InputGroup,
  Nav,
  Tab
} from 'react-bootstrap';
import { adminApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

const AdminDashboard = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // State for different sections
  const [systemStats, setSystemStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [movies, setMovies] = useState([]);
  const [reviews, setReviews] = useState([]);
  const [activities, setActivities] = useState([]);
  const [systemHealth, setSystemHealth] = useState(null);

  // Modal states
  const [showUserModal, setShowUserModal] = useState(false);
  const [showMovieModal, setShowMovieModal] = useState(false);
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);

  // Form states
  const [userForm, setUserForm] = useState({
    username: '',
    email: '',
    display_name: '',
    role: 'user',
    is_active: true
  });
  const [movieForm, setMovieForm] = useState({
    title: '',
    description: '',
    genre: '',
    release_year: '',
    director: '',
    cast: '',
    rating: '',
    poster_url: ''
  });
  const [reviewForm, setReviewForm] = useState({
    movie_id: '',
    user_id: '',
    rating: 5,
    review_text: '',
    is_approved: true
  });

  // Search and filter states
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState('all');

  useEffect(() => {
    if (auth?.user_id) {
      loadSystemStats();
      loadSystemHealth();
    }
  }, [auth]);

  useEffect(() => {
    if (activeTab === 'users') loadUsers();
    if (activeTab === 'movies') loadMovies();
    if (activeTab === 'reviews') loadReviews();
    if (activeTab === 'activities') loadActivities();
  }, [activeTab]);

  const loadSystemStats = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getSystemStats();
      setSystemStats(response.data.stats);
    } catch (err) {
      setError('Failed to load system statistics');
    } finally {
      setLoading(false);
    }
  };

  const loadSystemHealth = async () => {
    try {
      const response = await adminApi.getSystemHealth();
      setSystemHealth(response.data);
    } catch (err) {
      console.error('Failed to load system health:', err);
    }
  };

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getAllUsers();
      setUsers(response.data.users || []);
    } catch (err) {
      setError('Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const loadMovies = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getAllMovies();
      setMovies(response.data.movies || []);
    } catch (err) {
      setError('Failed to load movies');
    } finally {
      setLoading(false);
    }
  };

  const loadReviews = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getAllReviews();
      setReviews(response.data.reviews || []);
    } catch (err) {
      setError('Failed to load reviews');
    } finally {
      setLoading(false);
    }
  };

  const loadActivities = async () => {
    try {
      setLoading(true);
      const response = await adminApi.getAllActivities();
      setActivities(response.data.activities || []);
    } catch (err) {
      setError('Failed to load activities');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async (e) => {
    e.preventDefault();
    try {
      await adminApi.createUser(userForm);
      setSuccess('User created successfully!');
      setShowUserModal(false);
      loadUsers();
      setUserForm({
        username: '',
        email: '',
        display_name: '',
        role: 'user',
        is_active: true
      });
    } catch (err) {
      setError('Failed to create user');
    }
  };

  const handleUpdateUser = async (userId, userData) => {
    try {
      await adminApi.updateUser(userId, userData);
      setSuccess('User updated successfully!');
      loadUsers();
    } catch (err) {
      setError('Failed to update user');
    }
  };

  const handleDeleteUser = async (userId) => {
    if (window.confirm('Are you sure you want to delete this user?')) {
      try {
        await adminApi.deleteUser(userId);
        setSuccess('User deleted successfully!');
        loadUsers();
      } catch (err) {
        setError('Failed to delete user');
      }
    }
  };

  const handleCreateMovie = async (e) => {
    e.preventDefault();
    try {
      await adminApi.createMovie(movieForm);
      setSuccess('Movie created successfully!');
      setShowMovieModal(false);
      loadMovies();
      setMovieForm({
        title: '',
        description: '',
        genre: '',
        release_year: '',
        director: '',
        cast: '',
        rating: '',
        poster_url: ''
      });
    } catch (err) {
      setError('Failed to create movie');
    }
  };

  const handleUpdateReview = async (reviewId, reviewData) => {
    try {
      await adminApi.updateReview(reviewId, reviewData);
      setSuccess('Review updated successfully!');
      loadReviews();
    } catch (err) {
      setError('Failed to update review');
    }
  };

  const handleDeleteReview = async (reviewId) => {
    if (window.confirm('Are you sure you want to delete this review?')) {
      try {
        await adminApi.deleteReview(reviewId);
        setSuccess('Review deleted successfully!');
        loadReviews();
      } catch (err) {
        setError('Failed to delete review');
      }
    }
  };

  const clearMessages = () => {
    setError('');
    setSuccess('');
  };

  const filteredUsers = users.filter(user => 
    user.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredMovies = movies.filter(movie => 
    movie.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    movie.genre.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Container fluid className="py-4">
      <Row>
        <Col>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>🛡️ Admin Dashboard</h2>
            <div>
              <Button variant="outline-info" onClick={loadSystemHealth} className="me-2">
                🔄 Refresh Health
              </Button>
              <Button variant="outline-primary" onClick={loadSystemStats}>
                🔄 Refresh Stats
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

          {/* System Overview */}
          {systemStats && (
            <Row className="mb-4">
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h3 className="text-primary">{systemStats.total_users}</h3>
                    <p className="mb-0">Total Users</p>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h3 className="text-success">{systemStats.total_movies}</h3>
                    <p className="mb-0">Total Movies</p>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h3 className="text-warning">{systemStats.total_reviews}</h3>
                    <p className="mb-0">Total Reviews</p>
                  </Card.Body>
                </Card>
              </Col>
              <Col md={3}>
                <Card className="text-center">
                  <Card.Body>
                    <h3 className="text-info">{systemStats.total_activities}</h3>
                    <p className="mb-0">Activities</p>
                  </Card.Body>
                </Card>
              </Col>
            </Row>
          )}

          {/* System Health */}
          {systemHealth && (
            <Card className="mb-4">
              <Card.Header as="h5">🏥 System Health</Card.Header>
              <Card.Body>
                <Row>
                  <Col md={3}>
                    <Badge bg={systemHealth.database.status === 'healthy' ? 'success' : 'danger'}>
                      Database: {systemHealth.database.status}
                    </Badge>
                  </Col>
                  <Col md={3}>
                    <Badge bg={systemHealth.redis.status === 'healthy' ? 'success' : 'danger'}>
                      Redis: {systemHealth.redis.status}
                    </Badge>
                  </Col>
                  <Col md={3}>
                    <Badge bg={systemHealth.api.status === 'healthy' ? 'success' : 'danger'}>
                      API: {systemHealth.api.status}
                    </Badge>
                  </Col>
                  <Col md={3}>
                    <Badge bg={systemHealth.websocket.status === 'healthy' ? 'success' : 'danger'}>
                      WebSocket: {systemHealth.websocket.status}
                    </Badge>
                  </Col>
                </Row>
              </Card.Body>
            </Card>
          )}

          {/* Main Content Tabs */}
          <Card>
            <Card.Header>
              <Nav variant="tabs" activeKey={activeTab} onSelect={(k) => setActiveTab(k)}>
                <Nav.Item>
                  <Nav.Link eventKey="users">👥 Users</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="movies">🎬 Movies</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="reviews">⭐ Reviews</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="activities">📋 Activities</Nav.Link>
                </Nav.Item>
              </Nav>
            </Card.Header>

            <Card.Body>
              <Tab.Content>
                {/* Users Tab */}
                <Tab.Pane active={activeTab === 'users'}>
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <InputGroup style={{ maxWidth: '300px' }}>
                      <Form.Control
                        placeholder="Search users..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                      />
                    </InputGroup>
                    <Button variant="primary" onClick={() => setShowUserModal(true)}>
                      ➕ Add User
                    </Button>
                  </div>

                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>Username</th>
                          <th>Email</th>
                          <th>Display Name</th>
                          <th>Role</th>
                          <th>Status</th>
                          <th>Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredUsers.map((user) => (
                          <tr key={user.user_id}>
                            <td>{user.username}</td>
                            <td>{user.email}</td>
                            <td>{user.display_name}</td>
                            <td>
                              <Badge bg={user.role === 'admin' ? 'danger' : 'primary'}>
                                {user.role}
                              </Badge>
                            </td>
                            <td>
                              <Badge bg={user.is_active ? 'success' : 'secondary'}>
                                {user.is_active ? 'Active' : 'Inactive'}
                              </Badge>
                            </td>
                            <td>
                              <Button 
                                variant="outline-warning" 
                                size="sm" 
                                className="me-1"
                                onClick={() => handleUpdateUser(user.user_id, { is_active: !user.is_active })}
                              >
                                Toggle
                              </Button>
                              <Button 
                                variant="outline-danger" 
                                size="sm"
                                onClick={() => handleDeleteUser(user.user_id)}
                              >
                                Delete
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </Tab.Pane>

                {/* Movies Tab */}
                <Tab.Pane active={activeTab === 'movies'}>
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <InputGroup style={{ maxWidth: '300px' }}>
                      <Form.Control
                        placeholder="Search movies..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                      />
                    </InputGroup>
                    <Button variant="primary" onClick={() => setShowMovieModal(true)}>
                      ➕ Add Movie
                    </Button>
                  </div>

                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>Title</th>
                          <th>Genre</th>
                          <th>Year</th>
                          <th>Director</th>
                          <th>Rating</th>
                          <th>Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredMovies.map((movie) => (
                          <tr key={movie.movie_id}>
                            <td>{movie.title}</td>
                            <td>{movie.genre}</td>
                            <td>{movie.release_year}</td>
                            <td>{movie.director}</td>
                            <td>{movie.rating}</td>
                            <td>
                              <Button variant="outline-primary" size="sm">
                                Edit
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </Tab.Pane>

                {/* Reviews Tab */}
                <Tab.Pane active={activeTab === 'reviews'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>Movie</th>
                          <th>User</th>
                          <th>Rating</th>
                          <th>Review</th>
                          <th>Status</th>
                          <th>Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {reviews.map((review) => (
                          <tr key={review.review_id}>
                            <td>{review.movie_title}</td>
                            <td>{review.user_display_name}</td>
                            <td>
                              <Badge bg="warning">⭐ {review.rating}</Badge>
                            </td>
                            <td>{review.review_text?.substring(0, 50)}...</td>
                            <td>
                              <Badge bg={review.is_approved ? 'success' : 'secondary'}>
                                {review.is_approved ? 'Approved' : 'Pending'}
                              </Badge>
                            </td>
                            <td>
                              <Button 
                                variant="outline-warning" 
                                size="sm" 
                                className="me-1"
                                onClick={() => handleUpdateReview(review.review_id, { is_approved: !review.is_approved })}
                              >
                                Toggle
                              </Button>
                              <Button 
                                variant="outline-danger" 
                                size="sm"
                                onClick={() => handleDeleteReview(review.review_id)}
                              >
                                Delete
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </Tab.Pane>

                {/* Activities Tab */}
                <Tab.Pane active={activeTab === 'activities'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : (
                    <Table striped bordered hover responsive>
                      <thead>
                        <tr>
                          <th>User</th>
                          <th>Type</th>
                          <th>Title</th>
                          <th>Description</th>
                          <th>Date</th>
                        </tr>
                      </thead>
                      <tbody>
                        {activities.map((activity, index) => (
                          <tr key={index}>
                            <td>{activity.user_display_name}</td>
                            <td>
                              <Badge bg="primary">{activity.type}</Badge>
                            </td>
                            <td>{activity.title}</td>
                            <td>{activity.description?.substring(0, 50)}...</td>
                            <td>{new Date(activity.created_at).toLocaleDateString()}</td>
                          </tr>
                        ))}
                      </tbody>
                    </Table>
                  )}
                </Tab.Pane>
              </Tab.Content>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Create User Modal */}
      <Modal show={showUserModal} onHide={() => setShowUserModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Create New User</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleCreateUser}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Username</Form.Label>
              <Form.Control
                type="text"
                value={userForm.username}
                onChange={(e) => setUserForm({...userForm, username: e.target.value})}
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Email</Form.Label>
              <Form.Control
                type="email"
                value={userForm.email}
                onChange={(e) => setUserForm({...userForm, email: e.target.value})}
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Display Name</Form.Label>
              <Form.Control
                type="text"
                value={userForm.display_name}
                onChange={(e) => setUserForm({...userForm, display_name: e.target.value})}
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Role</Form.Label>
              <Form.Select
                value={userForm.role}
                onChange={(e) => setUserForm({...userForm, role: e.target.value})}
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </Form.Select>
            </Form.Group>
            <Form.Check
              type="checkbox"
              label="Active"
              checked={userForm.is_active}
              onChange={(e) => setUserForm({...userForm, is_active: e.target.checked})}
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowUserModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Create User
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Create Movie Modal */}
      <Modal show={showMovieModal} onHide={() => setShowMovieModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Add New Movie</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleCreateMovie}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Title</Form.Label>
              <Form.Control
                type="text"
                value={movieForm.title}
                onChange={(e) => setMovieForm({...movieForm, title: e.target.value})}
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={movieForm.description}
                onChange={(e) => setMovieForm({...movieForm, description: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Genre</Form.Label>
              <Form.Control
                type="text"
                value={movieForm.genre}
                onChange={(e) => setMovieForm({...movieForm, genre: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Release Year</Form.Label>
              <Form.Control
                type="number"
                value={movieForm.release_year}
                onChange={(e) => setMovieForm({...movieForm, release_year: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Director</Form.Label>
              <Form.Control
                type="text"
                value={movieForm.director}
                onChange={(e) => setMovieForm({...movieForm, director: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Cast</Form.Label>
              <Form.Control
                type="text"
                value={movieForm.cast}
                onChange={(e) => setMovieForm({...movieForm, cast: e.target.value})}
                placeholder="Comma-separated actor names"
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Rating</Form.Label>
              <Form.Control
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={movieForm.rating}
                onChange={(e) => setMovieForm({...movieForm, rating: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Poster URL</Form.Label>
              <Form.Control
                type="url"
                value={movieForm.poster_url}
                onChange={(e) => setMovieForm({...movieForm, poster_url: e.target.value})}
              />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowMovieModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Add Movie
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </Container>
  );
};

export default AdminDashboard;