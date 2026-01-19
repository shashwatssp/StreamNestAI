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
  Nav,
  Tab,
  Dropdown
} from 'react-bootstrap';
import { watchlistApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

// Comprehensive logging utility
const logWatchlistFeature = (action, data = {}, level = 'info') => {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    feature: 'WATCHLIST_MANAGER',
    action,
    data,
    level,
    userId: localStorage.getItem('user_id') || 'anonymous'
  };
  
  const logMessage = `[${timestamp}] 📝 [WATCHLIST_MANAGER] ${action}: ${JSON.stringify(data)}`;
  
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

const WatchlistManager = () => {
  const { auth } = useAuth();
  const [activeTab, setActiveTab] = useState('my-watchlists');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // State for watchlists
  const [myWatchlists, setMyWatchlists] = useState([]);
  const [publicWatchlists, setPublicWatchlists] = useState([]);
  const [sharedWatchlists, setSharedWatchlists] = useState([]);
  const [trendingWatchlists, setTrendingWatchlists] = useState([]);
  const [currentWatchlist, setCurrentWatchlist] = useState(null);
  const [watchlistItems, setWatchlistItems] = useState([]);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [showAddItemModal, setShowAddItemModal] = useState(false);

  // Form states
  const [watchlistForm, setWatchlistForm] = useState({
    name: '',
    description: '',
    is_public: false,
    is_collaborative: false,
    tags: []
  });
  const [addItemForm, setAddItemForm] = useState({
    movie_id: '',
    notes: '',
    priority: 'medium'
  });

  // Search and filter states
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [filterGenre, setFilterGenre] = useState('all');

  useEffect(() => {
    logWatchlistFeature('COMPONENT_MOUNT', { userId: auth?.user_id, hasAuth: !!auth?.user_id });
    if (auth?.user_id) {
      logWatchlistFeature('INITIAL_LOAD_START', { userId: auth.user_id });
      loadMyWatchlists();
      loadPublicWatchlists();
      loadTrendingWatchlists();
    } else {
      logWatchlistFeature('NO_AUTH_USER', {}, 'warn');
    }
  }, [auth]);

  useEffect(() => {
    logWatchlistFeature('TAB_CHANGE', { activeTab, userId: auth?.user_id });
    if (activeTab === 'shared') loadSharedWatchlists();
  }, [activeTab]);

  const loadMyWatchlists = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      logWatchlistFeature('API_CALL_START', {
        endpoint: 'getUserWatchlists',
        userId: auth.user_id
      });
      
      const response = await watchlistApi.getUserWatchlists(auth.user_id);
      const duration = Date.now() - startTime;
      
      logWatchlistFeature('API_CALL_SUCCESS', {
        endpoint: 'getUserWatchlists',
        userId: auth.user_id,
        duration: `${duration}ms`,
        watchlistsCount: response.data?.watchlists?.length || 0,
        hasPublicWatchlists: response.data?.watchlists?.some(w => w.is_public),
        hasCollaborativeWatchlists: response.data?.watchlists?.some(w => w.is_collaborative)
      });
      
      setMyWatchlists(response.data.watchlists || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logWatchlistFeature('API_CALL_ERROR', {
        endpoint: 'getUserWatchlists',
        userId: auth.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to load your watchlists');
    } finally {
      setLoading(false);
      logWatchlistFeature('MY_WATCHLISTS_LOAD_END', {
        userId: auth.user_id,
        count: myWatchlists.length
      });
    }
  };

  const loadPublicWatchlists = async () => {
    try {
      setLoading(true);
      const response = await watchlistApi.getPublicWatchlists();
      setPublicWatchlists(response.data.watchlists || []);
    } catch (err) {
      setError('Failed to load public watchlists');
    } finally {
      setLoading(false);
    }
  };

  const loadSharedWatchlists = async () => {
    try {
      setLoading(true);
      const response = await watchlistApi.getSharedWatchlists();
      setSharedWatchlists(response.data.watchlists || []);
    } catch (err) {
      setError('Failed to load shared watchlists');
    } finally {
      setLoading(false);
    }
  };

  const loadTrendingWatchlists = async () => {
    try {
      setLoading(true);
      const response = await watchlistApi.getTrendingWatchlists();
      setTrendingWatchlists(response.data.watchlists || []);
    } catch (err) {
      setError('Failed to load trending watchlists');
    } finally {
      setLoading(false);
    }
  };

  const loadWatchlistItems = async (watchlistId) => {
    try {
      setLoading(true);
      const response = await watchlistApi.getWatchlistItems(watchlistId);
      setWatchlistItems(response.data.items || []);
    } catch (err) {
      setError('Failed to load watchlist items');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateWatchlist = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logWatchlistFeature('CREATE_WATCHLIST_START', {
        userId: auth.user_id,
        watchlistData: {
          name: watchlistForm.name,
          hasDescription: !!watchlistForm.description,
          isPublic: watchlistForm.is_public,
          isCollaborative: watchlistForm.is_collaborative,
          tagsCount: watchlistForm.tags.length
        }
      });
      
      await watchlistApi.createWatchlist(watchlistForm);
      const duration = Date.now() - startTime;
      
      logWatchlistFeature('CREATE_WATCHLIST_SUCCESS', {
        userId: auth.user_id,
        watchlistName: watchlistForm.name,
        duration: `${duration}ms`
      });
      
      setSuccess('Watchlist created successfully!');
      setShowCreateModal(false);
      loadMyWatchlists();
      setWatchlistForm({
        name: '',
        description: '',
        is_public: false,
        is_collaborative: false,
        tags: []
      });
    } catch (err) {
      const duration = Date.now() - startTime;
      logWatchlistFeature('CREATE_WATCHLIST_ERROR', {
        userId: auth.user_id,
        watchlistName: watchlistForm.name,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to create watchlist');
    }
  };

  const handleUpdateWatchlist = async (watchlistId, updateData) => {
    try {
      await watchlistApi.updateWatchlist(watchlistId, updateData);
      setSuccess('Watchlist updated successfully!');
      setShowEditModal(false);
      loadMyWatchlists();
    } catch (err) {
      setError('Failed to update watchlist');
    }
  };

  const handleDeleteWatchlist = async (watchlistId) => {
    if (window.confirm('Are you sure you want to delete this watchlist?')) {
      const startTime = Date.now();
      try {
        logWatchlistFeature('DELETE_WATCHLIST_START', {
          userId: auth.user_id,
          watchlistId
        });
        
        await watchlistApi.deleteWatchlist(watchlistId);
        const duration = Date.now() - startTime;
        
        logWatchlistFeature('DELETE_WATCHLIST_SUCCESS', {
          userId: auth.user_id,
          watchlistId,
          duration: `${duration}ms`
        });
        
        setSuccess('Watchlist deleted successfully!');
        loadMyWatchlists();
      } catch (err) {
        const duration = Date.now() - startTime;
        logWatchlistFeature('DELETE_WATCHLIST_ERROR', {
          userId: auth.user_id,
          watchlistId,
          duration: `${duration}ms`,
          error: err.message,
          status: err.response?.status
        }, 'error');
        setError('Failed to delete watchlist');
      }
    } else {
      logWatchlistFeature('DELETE_WATCHLIST_CANCELLED', {
        userId: auth.user_id,
        watchlistId
      });
    }
  };

  const handleAddItem = async (e) => {
    e.preventDefault();
    const startTime = Date.now();
    try {
      logWatchlistFeature('ADD_ITEM_START', {
        userId: auth.user_id,
        watchlistId: currentWatchlist.watchlist_id,
        watchlistName: currentWatchlist.name,
        itemData: {
          movieId: addItemForm.movie_id,
          hasNotes: !!addItemForm.notes,
          priority: addItemForm.priority
        }
      });
      
      await watchlistApi.addItemToWatchlist(currentWatchlist.watchlist_id, addItemForm);
      const duration = Date.now() - startTime;
      
      logWatchlistFeature('ADD_ITEM_SUCCESS', {
        userId: auth.user_id,
        watchlistId: currentWatchlist.watchlist_id,
        movieId: addItemForm.movie_id,
        duration: `${duration}ms`
      });
      
      setSuccess('Item added to watchlist!');
      setShowAddItemModal(false);
      loadWatchlistItems(currentWatchlist.watchlist_id);
      setAddItemForm({
        movie_id: '',
        notes: '',
        priority: 'medium'
      });
    } catch (err) {
      const duration = Date.now() - startTime;
      logWatchlistFeature('ADD_ITEM_ERROR', {
        userId: auth.user_id,
        watchlistId: currentWatchlist.watchlist_id,
        movieId: addItemForm.movie_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to add item to watchlist');
    }
  };

  const handleRemoveItem = async (itemId) => {
    try {
      await watchlistApi.removeItemFromWatchlist(currentWatchlist.watchlist_id, itemId);
      setSuccess('Item removed from watchlist!');
      loadWatchlistItems(currentWatchlist.watchlist_id);
    } catch (err) {
      setError('Failed to remove item from watchlist');
    }
  };

  const handleShareWatchlist = async (shareData) => {
    try {
      await watchlistApi.shareWatchlist(currentWatchlist.watchlist_id, shareData);
      setSuccess('Watchlist shared successfully!');
      setShowShareModal(false);
    } catch (err) {
      setError('Failed to share watchlist');
    }
  };

  const handleLikeWatchlist = async (watchlistId) => {
    try {
      await watchlistApi.likeWatchlist(watchlistId);
      setSuccess('Watchlist liked!');
      // Refresh the appropriate list
      if (activeTab === 'my-watchlists') loadMyWatchlists();
      if (activeTab === 'public') loadPublicWatchlists();
      if (activeTab === 'trending') loadTrendingWatchlists();
    } catch (err) {
      setError('Failed to like watchlist');
    }
  };

  const handleFollowWatchlist = async (watchlistId) => {
    try {
      await watchlistApi.followWatchlist(watchlistId);
      setSuccess('You are now following this watchlist!');
      loadSharedWatchlists();
    } catch (err) {
      setError('Failed to follow watchlist');
    }
  };

  const clearMessages = () => {
    logWatchlistFeature('CLEAR_MESSAGES', { userId: auth.user_id });
    setError('');
    setSuccess('');
  };

  const openWatchlist = (watchlist) => {
    logWatchlistFeature('WATCHLIST_OPEN', {
      userId: auth.user_id,
      watchlistId: watchlist.watchlist_id,
      watchlistName: watchlist.name,
      itemCount: watchlist.item_count,
      isPublic: watchlist.is_public
    });
    setCurrentWatchlist(watchlist);
    loadWatchlistItems(watchlist.watchlist_id);
    setActiveTab('watchlist-details');
  };

  const getPriorityBadge = (priority) => {
    const variants = {
      high: 'danger',
      medium: 'warning',
      low: 'success'
    };
    return <Badge bg={variants[priority] || 'secondary'}>{priority}</Badge>;
  };

  const renderWatchlistCard = (watchlist, showActions = true) => (
    <Col md={4} className="mb-4" key={watchlist.watchlist_id}>
      <Card className="h-100">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-start mb-2">
            <Card.Title className="h6">{watchlist.name}</Card.Title>
            <div>
              {watchlist.is_public && <Badge bg="info" className="me-1">Public</Badge>}
              {watchlist.is_collaborative && <Badge bg="warning">Collaborative</Badge>}
            </div>
          </div>
          <Card.Text className="small text-muted mb-2">
            {watchlist.description}
          </Card.Text>
          <div className="d-flex justify-content-between align-items-center mb-2">
            <small className="text-muted">
              {watchlist.item_count} items • by {watchlist.created_by_display_name}
            </small>
          </div>
          <div className="d-flex justify-content-between align-items-center mb-3">
            <div>
              <Button 
                variant="outline-danger" 
                size="sm" 
                className="me-2"
                onClick={() => handleLikeWatchlist(watchlist.watchlist_id)}
              >
                ❤️ {watchlist.likes_count}
              </Button>
              <small className="text-muted">
                👁️ {watchlist.views_count} views
              </small>
            </div>
            <small className="text-muted">
              {new Date(watchlist.created_at).toLocaleDateString()}
            </small>
          </div>
          {watchlist.tags && watchlist.tags.length > 0 && (
            <div className="mb-2">
              {watchlist.tags.map((tag, index) => (
                <Badge bg="secondary" className="me-1" key={index}>
                  {tag}
                </Badge>
              ))}
            </div>
          )}
        </Card.Body>
        <Card.Footer className="bg-transparent">
          <div className="d-grid gap-2">
            <Button 
              variant="primary" 
              size="sm"
              onClick={() => openWatchlist(watchlist)}
            >
              View Items
            </Button>
            {showActions && (
              <div className="d-flex gap-2">
                <Button 
                  variant="outline-info" 
                  size="sm"
                  onClick={() => handleFollowWatchlist(watchlist.watchlist_id)}
                >
                  Follow
                </Button>
                <Button 
                  variant="outline-success" 
                  size="sm"
                  onClick={() => {
                    setCurrentWatchlist(watchlist);
                    setShowShareModal(true);
                  }}
                >
                  Share
                </Button>
              </div>
            )}
          </div>
        </Card.Footer>
      </Card>
    </Col>
  );

  return (
    <Container fluid className="py-4">
      <Row>
        <Col>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>📝 Watchlist Manager</h2>
            <div className="d-flex align-items-center">
              <Button 
                variant="primary" 
                className="me-2"
                onClick={() => setShowCreateModal(true)}
              >
                ➕ Create Watchlist
              </Button>
              <Button variant="outline-info" onClick={loadMyWatchlists}>
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

          {/* Main Content Tabs */}
          <Card>
            <Card.Header>
              <Nav variant="tabs" activeKey={activeTab} onSelect={(k) => setActiveTab(k)}>
                <Nav.Item>
                  <Nav.Link eventKey="my-watchlists">📋 My Watchlists</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="public">🌍 Public Watchlists</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="shared">🤝 Shared with Me</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link eventKey="trending">🔥 Trending</Nav.Link>
                </Nav.Item>
                {currentWatchlist && (
                  <Nav.Item>
                    <Nav.Link eventKey="watchlist-details">
                      📂 {currentWatchlist.name}
                    </Nav.Link>
                  </Nav.Item>
                )}
              </Nav>
            </Card.Header>

            <Card.Body>
              <Tab.Content>
                {/* My Watchlists Tab */}
                <Tab.Pane active={activeTab === 'my-watchlists'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : myWatchlists.length > 0 ? (
                    <Row>
                      {myWatchlists.map((watchlist) => (
                        <Col md={4} className="mb-4" key={watchlist.watchlist_id}>
                          <Card className="h-100">
                            <Card.Body>
                              <div className="d-flex justify-content-between align-items-start mb-2">
                                <Card.Title className="h6">{watchlist.name}</Card.Title>
                                <div>
                                  {watchlist.is_public && <Badge bg="info" className="me-1">Public</Badge>}
                                  {watchlist.is_collaborative && <Badge bg="warning">Collaborative</Badge>}
                                </div>
                              </div>
                              <Card.Text className="small text-muted mb-2">
                                {watchlist.description}
                              </Card.Text>
                              <div className="d-flex justify-content-between align-items-center mb-2">
                                <small className="text-muted">
                                  {watchlist.item_count} items
                                </small>
                                <small className="text-muted">
                                  {new Date(watchlist.created_at).toLocaleDateString()}
                                </small>
                              </div>
                            </Card.Body>
                            <Card.Footer className="bg-transparent">
                              <div className="d-grid gap-2">
                                <Button 
                                  variant="primary" 
                                  size="sm"
                                  onClick={() => openWatchlist(watchlist)}
                                >
                                  View Items
                                </Button>
                                <div className="d-flex gap-2">
                                  <Button 
                                    variant="outline-warning" 
                                    size="sm"
                                    onClick={() => {
                                      setCurrentWatchlist(watchlist);
                                      setShowEditModal(true);
                                    }}
                                  >
                                    Edit
                                  </Button>
                                  <Button 
                                    variant="outline-danger" 
                                    size="sm"
                                    onClick={() => handleDeleteWatchlist(watchlist.watchlist_id)}
                                  >
                                    Delete
                                  </Button>
                                </div>
                              </div>
                            </Card.Footer>
                          </Card>
                        </Col>
                      ))}
                    </Row>
                  ) : (
                    <Alert variant="info">
                      You haven't created any watchlists yet. Create your first one to get started!
                    </Alert>
                  )}
                </Tab.Pane>

                {/* Public Watchlists Tab */}
                <Tab.Pane active={activeTab === 'public'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : publicWatchlists.length > 0 ? (
                    <Row>
                      {publicWatchlists.map((watchlist) => renderWatchlistCard(watchlist))}
                    </Row>
                  ) : (
                    <Alert variant="info">No public watchlists available</Alert>
                  )}
                </Tab.Pane>

                {/* Shared Watchlists Tab */}
                <Tab.Pane active={activeTab === 'shared'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : sharedWatchlists.length > 0 ? (
                    <Row>
                      {sharedWatchlists.map((watchlist) => renderWatchlistCard(watchlist))}
                    </Row>
                  ) : (
                    <Alert variant="info">No watchlists have been shared with you</Alert>
                  )}
                </Tab.Pane>

                {/* Trending Watchlists Tab */}
                <Tab.Pane active={activeTab === 'trending'}>
                  {loading ? (
                    <div className="text-center py-4">
                      <Spinner animation="border" />
                    </div>
                  ) : trendingWatchlists.length > 0 ? (
                    <Row>
                      {trendingWatchlists.map((watchlist, index) => (
                        <Col md={4} className="mb-4" key={watchlist.watchlist_id}>
                          <Card className="h-100 border-warning">
                            <Card.Header className="bg-warning text-dark">
                              <strong>🏆 #{index + 1} Trending</strong>
                            </Card.Header>
                            <Card.Body>
                              <div className="d-flex justify-content-between align-items-start mb-2">
                                <Card.Title className="h6">{watchlist.name}</Card.Title>
                                <Badge bg="info">Public</Badge>
                              </div>
                              <Card.Text className="small text-muted mb-2">
                                {watchlist.description}
                              </Card.Text>
                              <div className="d-flex justify-content-between align-items-center mb-2">
                                <small className="text-muted">
                                  {watchlist.item_count} items • by {watchlist.created_by_display_name}
                                </small>
                              </div>
                              <div className="d-flex justify-content-between align-items-center mb-3">
                                <div>
                                  <Button 
                                    variant="outline-danger" 
                                    size="sm" 
                                    className="me-2"
                                    onClick={() => handleLikeWatchlist(watchlist.watchlist_id)}
                                  >
                                    ❤️ {watchlist.likes_count}
                                  </Button>
                                  <small className="text-muted">
                                    👁️ {watchlist.views_count} views
                                  </small>
                                </div>
                                <small className="text-muted">
                                  {new Date(watchlist.created_at).toLocaleDateString()}
                                </small>
                              </div>
                            </Card.Body>
                            <Card.Footer className="bg-transparent">
                              <div className="d-grid gap-2">
                                <Button 
                                  variant="primary" 
                                  size="sm"
                                  onClick={() => openWatchlist(watchlist)}
                                >
                                  View Items
                                </Button>
                                <div className="d-flex gap-2">
                                  <Button 
                                    variant="outline-info" 
                                    size="sm"
                                    onClick={() => handleFollowWatchlist(watchlist.watchlist_id)}
                                  >
                                    Follow
                                  </Button>
                                  <Button 
                                    variant="outline-success" 
                                    size="sm"
                                    onClick={() => {
                                      setCurrentWatchlist(watchlist);
                                      setShowShareModal(true);
                                    }}
                                  >
                                    Share
                                  </Button>
                                </div>
                              </div>
                            </Card.Footer>
                          </Card>
                        </Col>
                      ))}
                    </Row>
                  ) : (
                    <Alert variant="info">No trending watchlists available</Alert>
                  )}
                </Tab.Pane>

                {/* Watchlist Details Tab */}
                <Tab.Pane active={activeTab === 'watchlist-details'}>
                  {currentWatchlist ? (
                    <div>
                      <div className="d-flex justify-content-between align-items-center mb-4">
                        <div>
                          <h4>{currentWatchlist.name}</h4>
                          <p className="text-muted">{currentWatchlist.description}</p>
                        </div>
                        <Button 
                          variant="primary"
                          onClick={() => setShowAddItemModal(true)}
                        >
                          ➕ Add Item
                        </Button>
                      </div>

                      {loading ? (
                        <div className="text-center py-4">
                          <Spinner animation="border" />
                        </div>
                      ) : watchlistItems.length > 0 ? (
                        <ListGroup>
                          {watchlistItems.map((item) => (
                            <ListGroup.Item key={item.item_id} className="d-flex justify-content-between align-items-center">
                              <div className="d-flex align-items-center">
                                <img 
                                  src={item.poster_url || '/placeholder-movie.jpg'} 
                                  alt={item.title}
                                  style={{ width: '50px', height: '75px', objectFit: 'cover', marginRight: '15px' }}
                                />
                                <div>
                                  <h6 className="mb-1">{item.title}</h6>
                                  <small className="text-muted">{item.genre} • {item.year}</small>
                                  {item.notes && (
                                    <p className="small mb-0 mt-1">{item.notes}</p>
                                  )}
                                </div>
                              </div>
                              <div className="d-flex align-items-center">
                                {getPriorityBadge(item.priority)}
                                <Button 
                                  variant="outline-danger" 
                                  size="sm" 
                                  className="ms-3"
                                  onClick={() => handleRemoveItem(item.item_id)}
                                >
                                  Remove
                                </Button>
                              </div>
                            </ListGroup.Item>
                          ))}
                        </ListGroup>
                      ) : (
                        <Alert variant="info">
                          This watchlist is empty. Add some items to get started!
                        </Alert>
                      )}
                    </div>
                  ) : (
                    <Alert variant="info">Select a watchlist to view its items</Alert>
                  )}
                </Tab.Pane>
              </Tab.Content>
            </Card.Body>
          </Card>
        </Col>
      </Row>

      {/* Create Watchlist Modal */}
      <Modal show={showCreateModal} onHide={() => setShowCreateModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Create New Watchlist</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleCreateWatchlist}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Name</Form.Label>
              <Form.Control
                type="text"
                value={watchlistForm.name}
                onChange={(e) => setWatchlistForm({...watchlistForm, name: e.target.value})}
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Description</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={watchlistForm.description}
                onChange={(e) => setWatchlistForm({...watchlistForm, description: e.target.value})}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Tags (comma-separated)</Form.Label>
              <Form.Control
                type="text"
                value={watchlistForm.tags.join(', ')}
                onChange={(e) => setWatchlistForm({...watchlistForm, tags: e.target.value.split(',').map(tag => tag.trim()).filter(tag => tag)})}
                placeholder="action, adventure, classic"
              />
            </Form.Group>
            <Form.Check
              type="checkbox"
              label="Public watchlist"
              checked={watchlistForm.is_public}
              onChange={(e) => setWatchlistForm({...watchlistForm, is_public: e.target.checked})}
              className="mb-2"
            />
            <Form.Check
              type="checkbox"
              label="Collaborative watchlist"
              checked={watchlistForm.is_collaborative}
              onChange={(e) => setWatchlistForm({...watchlistForm, is_collaborative: e.target.checked})}
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Create Watchlist
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Add Item Modal */}
      <Modal show={showAddItemModal} onHide={() => setShowAddItemModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Add Item to Watchlist</Modal.Title>
        </Modal.Header>
        <Form onSubmit={handleAddItem}>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Movie ID</Form.Label>
              <Form.Control
                type="text"
                value={addItemForm.movie_id}
                onChange={(e) => setAddItemForm({...addItemForm, movie_id: e.target.value})}
                placeholder="Enter movie ID"
                required
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Notes</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={addItemForm.notes}
                onChange={(e) => setAddItemForm({...addItemForm, notes: e.target.value})}
                placeholder="Add your notes about this movie..."
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Priority</Form.Label>
              <Form.Select
                value={addItemForm.priority}
                onChange={(e) => setAddItemForm({...addItemForm, priority: e.target.value})}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </Form.Select>
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={() => setShowAddItemModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" type="submit">
              Add Item
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </Container>
  );
};

export default WatchlistManager;