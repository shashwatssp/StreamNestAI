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
  InputGroup,
  ListGroup,
  Modal
} from 'react-bootstrap';
import { searchApi, movieApi } from '../../api/enhancedApi';
import useAuth from '../../hooks/useAuth';

// Comprehensive logging utility
const logSearchFeature = (action, data = {}, level = 'info') => {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    feature: 'NATURAL_SEARCH',
    action,
    data,
    level,
    userId: localStorage.getItem('user_id') || 'anonymous'
  };
  
  const logMessage = `[${timestamp}] 🔍 [NATURAL_SEARCH] ${action}: ${JSON.stringify(data)}`;
  
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

const NaturalLanguageSearch = () => {
  const { auth } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Search states
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [searchHistory, setSearchHistory] = useState([]);
  const [suggestedQueries, setSuggestedQueries] = useState([]);
  const [searchFilters, setSearchFilters] = useState({
    genre: '',
    year_range: '',
    rating_min: 0,
    rating_max: 10,
    language: '',
    duration: ''
  });

  // Modal states
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [selectedMovie, setSelectedMovie] = useState(null);

  // Search analytics
  const [searchAnalytics, setSearchAnalytics] = useState(null);
  const [trendingSearches, setTrendingSearches] = useState([]);

  useEffect(() => {
    logSearchFeature('COMPONENT_MOUNT', { userId: auth?.user_id, hasAuth: !!auth?.user_id });
    logSearchFeature('INITIAL_LOAD_START', { userId: auth?.user_id });
    loadSearchHistory();
    loadSuggestedQueries();
    loadTrendingSearches();
    loadSearchAnalytics();
  }, []);

  const loadSearchHistory = async () => {
    const startTime = Date.now();
    try {
      logSearchFeature('API_CALL_START', {
        endpoint: 'getSearchHistory',
        userId: auth?.user_id
      });
      
      const response = await searchApi.getSearchHistory();
      const duration = Date.now() - startTime;
      
      logSearchFeature('API_CALL_SUCCESS', {
        endpoint: 'getSearchHistory',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        historyCount: response.data?.history?.length || 0
      });
      
      setSearchHistory(response.data.history || []);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSearchFeature('API_CALL_ERROR', {
        endpoint: 'getSearchHistory',
        userId: auth?.user_id,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
    }
  };

  const loadSuggestedQueries = async () => {
    try {
      // Mock suggested queries - would come from API
      setSuggestedQueries([
        'action movies from the 90s',
        'comedy movies with high ratings',
        'sci-fi movies directed by Christopher Nolan',
        'horror movies that are actually scary',
        'romantic movies with happy endings',
        'thriller movies with plot twists',
        'adventure movies for family',
        'drama movies based on true stories'
      ]);
    } catch (err) {
      console.error('Failed to load suggested queries:', err);
    }
  };

  const loadTrendingSearches = async () => {
    try {
      // Mock trending searches - would come from API
      setTrendingSearches([
        'Marvel movies',
        'Oscar winners 2024',
        'Netflix original series',
        'Horror movies 2024',
        'Best action movies'
      ]);
    } catch (err) {
      console.error('Failed to load trending searches:', err);
    }
  };

  const loadSearchAnalytics = async () => {
    try {
      // Mock analytics - would come from API
      setSearchAnalytics({
        total_searches: 1250,
        unique_queries: 450,
        avg_results_per_search: 12.5,
        success_rate: 85.2,
        popular_genres: ['Action', 'Comedy', 'Drama', 'Thriller', 'Sci-Fi'],
        search_trends: {
          daily: '+15%',
          weekly: '+8%',
          monthly: '+23%'
        }
      });
    } catch (err) {
      console.error('Failed to load search analytics:', err);
    }
  };

  const handleNaturalLanguageSearch = async (e) => {
    e.preventDefault();
    if (!searchQuery.trim()) {
      logSearchFeature('SEARCH_EMPTY_QUERY', { userId: auth?.user_id }, 'warn');
      return;
    }

    const startTime = Date.now();
    try {
      setLoading(true);
      setError('');
      
      logSearchFeature('NATURAL_SEARCH_START', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        hasFilters: Object.values(searchFilters).some(v => v !== '' && v !== 0),
        filters: searchFilters
      });
      
      const response = await searchApi.naturalLanguageSearch(
        searchQuery,
        searchFilters.genre,
        searchFilters.year_range,
        searchFilters.rating_min.toString(),
        'relevance'
      );
      
      const duration = Date.now() - startTime;
      const resultsCount = response.data.results?.length || 0;
      
      logSearchFeature('NATURAL_SEARCH_SUCCESS', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        resultsCount,
        hasResults: resultsCount > 0
      });

      setSearchResults(response.data.results || []);
      
      // Add to search history
      const newHistoryItem = {
        query: searchQuery,
        timestamp: new Date().toISOString(),
        results_count: resultsCount
      };
      setSearchHistory([newHistoryItem, ...searchHistory.slice(0, 9)]);
      
      setSuccess(`Found ${resultsCount} results`);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSearchFeature('NATURAL_SEARCH_ERROR', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to perform search. Please try again.');
    } finally {
      setLoading(false);
      logSearchFeature('NATURAL_SEARCH_END', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        resultsCount: searchResults.length
      });
    }
  };

  const handleAdvancedSearch = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      setError('');
      
      logSearchFeature('ADVANCED_SEARCH_START', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        hasFilters: Object.values(searchFilters).some(v => v !== '' && v !== 0),
        filters: searchFilters
      });
      
      const response = await searchApi.advancedSearch(
        searchQuery,
        searchFilters.genre,
        searchFilters.year_range,
        searchFilters.rating_min.toString(),
        'relevance'
      );
      
      const duration = Date.now() - startTime;
      const resultsCount = response.data.results?.length || 0;
      
      logSearchFeature('ADVANCED_SEARCH_SUCCESS', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        resultsCount
      });

      setSearchResults(response.data.results || []);
      setSuccess(`Found ${resultsCount} results`);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSearchFeature('ADVANCED_SEARCH_ERROR', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to perform advanced search');
    } finally {
      setLoading(false);
    }
  };

  const handleSemanticSearch = async () => {
    const startTime = Date.now();
    try {
      setLoading(true);
      setError('');
      
      logSearchFeature('SEMANTIC_SEARCH_START', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        hasFilters: Object.values(searchFilters).some(v => v !== '' && v !== 0),
        filters: searchFilters
      });
      
      const response = await searchApi.semanticSearch(
        searchQuery,
        searchFilters.genre,
        searchFilters.year_range,
        searchFilters.rating_min.toString(),
        'relevance'
      );
      
      const duration = Date.now() - startTime;
      const resultsCount = response.data.results?.length || 0;
      
      logSearchFeature('SEMANTIC_SEARCH_SUCCESS', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        resultsCount
      });

      setSearchResults(response.data.results || []);
      setSuccess(`Found ${resultsCount} semantic matches`);
    } catch (err) {
      const duration = Date.now() - startTime;
      logSearchFeature('SEMANTIC_SEARCH_ERROR', {
        userId: auth?.user_id,
        query: searchQuery.length > 100 ? searchQuery.substring(0, 100) + '...' : searchQuery,
        duration: `${duration}ms`,
        error: err.message,
        status: err.response?.status
      }, 'error');
      setError('Failed to perform semantic search');
    } finally {
      setLoading(false);
    }
  };

  const handleSuggestedQuery = (query) => {
    logSearchFeature('SUGGESTED_QUERY_CLICK', {
      userId: auth?.user_id,
      suggestedQuery: query.length > 50 ? query.substring(0, 50) + '...' : query,
      previousQuery: searchQuery
    });
    
    setSearchQuery(query);
    // Auto-search when clicking suggested query
    setTimeout(() => {
      const form = document.getElementById('search-form');
      if (form) form.dispatchEvent(new Event('submit'));
    }, 100);
  };

  const clearSearch = () => {
    logSearchFeature('CLEAR_SEARCH', {
      userId: auth?.user_id,
      previousQuery: searchQuery,
      previousResultsCount: searchResults.length
    });
    
    setSearchQuery('');
    setSearchResults([]);
    setError('');
    setSuccess('');
  };

  const showMovieDetails = (movie) => {
    logSearchFeature('MOVIE_DETAILS_OPEN', {
      userId: auth?.user_id,
      movieId: movie._id,
      movieTitle: movie.title,
      movieYear: movie.year,
      relevanceScore: movie.relevance_score
    });
    
    setSelectedMovie(movie);
    setShowDetailsModal(true);
  };

  const getRelevanceBadge = (score) => {
    if (score >= 0.9) return <Badge bg="success">Excellent Match</Badge>;
    if (score >= 0.7) return <Badge bg="info">Good Match</Badge>;
    if (score >= 0.5) return <Badge bg="warning">Fair Match</Badge>;
    return <Badge bg="secondary">Low Match</Badge>;
  };

  const renderSearchResult = (result, index) => (
    <Col md={6} lg={4} className="mb-4" key={index}>
      <Card className="h-100">
        <Card.Img 
          variant="top" 
          src={result.poster_url || '/placeholder-movie.jpg'} 
          style={{ height: '200px', objectFit: 'cover' }}
        />
        <Card.Body>
          <div className="d-flex justify-content-between align-items-start mb-2">
            <Card.Title className="h6">{result.title}</Card.Title>
            {result.relevance_score && getRelevanceBadge(result.relevance_score)}
          </div>
          <Card.Text className="small text-muted mb-2">
            {result.genre} • {result.year} • {result.duration}
          </Card.Text>
          <Card.Text className="small">
            {result.description?.substring(0, 100)}...
          </Card.Text>
          <div className="d-flex justify-content-between align-items-center mb-2">
            <Badge bg="warning">⭐ {result.rating?.toFixed(1)}</Badge>
            <small className="text-muted">
              {result.match_score && `Match: ${Math.round(result.match_score * 100)}%`}
            </small>
          </div>
          {result.search_highlights && (
            <div className="mb-2">
              <small className="text-muted">
                <strong>Matched:</strong> {result.search_highlights.join(', ')}
              </small>
            </div>
          )}
        </Card.Body>
        <Card.Footer className="bg-transparent">
          <Button 
            variant="primary" 
            size="sm" 
            className="w-100"
            onClick={() => showMovieDetails(result)}
          >
            View Details
          </Button>
        </Card.Footer>
      </Card>
    </Col>
  );

  return (
    <Container fluid className="py-4">
      <Row>
        <Col lg={3}>
          {/* Search Analytics */}
          {searchAnalytics && (
            <Card className="mb-4">
              <Card.Header as="h6">📊 Search Analytics</Card.Header>
              <Card.Body>
                <p><strong>Total Searches:</strong> {searchAnalytics.total_searches}</p>
                <p><strong>Unique Queries:</strong> {searchAnalytics.unique_queries}</p>
                <p><strong>Avg Results:</strong> {searchAnalytics.avg_results_per_search}</p>
                <p><strong>Success Rate:</strong> {searchAnalytics.success_rate}%</p>
                <hr />
                <h6>Popular Genres</h6>
                <div className="d-flex flex-wrap gap-1">
                  {searchAnalytics.popular_genres.map((genre, index) => (
                    <Badge bg="secondary" key={index}>{genre}</Badge>
                  ))}
                </div>
              </Card.Body>
            </Card>
          )}

          {/* Trending Searches */}
          <Card className="mb-4">
            <Card.Header as="h6">🔥 Trending Searches</Card.Header>
            <Card.Body>
              <ListGroup variant="flush">
                {trendingSearches.map((trend, index) => (
                  <ListGroup.Item 
                    key={index} 
                    action 
                    onClick={() => handleSuggestedQuery(trend)}
                    className="small"
                  >
                    {trend}
                  </ListGroup.Item>
                ))}
              </ListGroup>
            </Card.Body>
          </Card>

          {/* Search History */}
          <Card>
            <Card.Header as="h6">🕐 Recent Searches</Card.Header>
            <Card.Body>
              {searchHistory.length > 0 ? (
                <ListGroup variant="flush">
                  {searchHistory.slice(0, 5).map((item, index) => (
                    <ListGroup.Item 
                      key={index} 
                      action 
                      onClick={() => handleSuggestedQuery(item.query)}
                      className="small"
                    >
                      <div>{item.query}</div>
                      <small className="text-muted">
                        {item.results_count} results • {new Date(item.timestamp).toLocaleDateString()}
                      </small>
                    </ListGroup.Item>
                  ))}
                </ListGroup>
              ) : (
                <p className="text-muted small">No search history</p>
              )}
            </Card.Body>
          </Card>
        </Col>

        <Col lg={9}>
          <div className="d-flex justify-content-between align-items-center mb-4">
            <h2>🔍 Natural Language Search</h2>
            <Button variant="outline-secondary" onClick={clearSearch}>
              Clear
            </Button>
          </div>

          {/* Alerts */}
          {error && (
            <Alert variant="danger" dismissible onClose={() => setError('')}>
              {error}
            </Alert>
          )}
          {success && (
            <Alert variant="success" dismissible onClose={() => setSuccess('')}>
              {success}
            </Alert>
          )}

          {/* Search Form */}
          <Card className="mb-4">
            <Card.Body>
              <Form id="search-form" onSubmit={handleNaturalLanguageSearch}>
                <InputGroup className="mb-3">
                  <Form.Control
                    size="lg"
                    placeholder="Search for movies using natural language... (e.g., 'action movies from the 90s with high ratings')"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                  />
                  <Button variant="primary" type="submit" disabled={loading}>
                    {loading ? <Spinner animation="border" size="sm" /> : 'Search'}
                  </Button>
                </InputGroup>

                <div className="d-flex gap-2 mb-3">
                  <Button 
                    variant="outline-info" 
                    onClick={handleAdvancedSearch}
                    disabled={loading}
                  >
                    🔬 Advanced Search
                  </Button>
                  <Button 
                    variant="outline-success" 
                    onClick={handleSemanticSearch}
                    disabled={loading}
                  >
                    🧠 Semantic Search
                  </Button>
                </div>

                {/* Search Filters */}
                <Row>
                  <Col md={3}>
                    <Form.Group className="mb-2">
                      <Form.Label>Genre</Form.Label>
                      <Form.Select
                        size="sm"
                        value={searchFilters.genre}
                        onChange={(e) => setSearchFilters({...searchFilters, genre: e.target.value})}
                      >
                        <option value="">All Genres</option>
                        <option value="action">Action</option>
                        <option value="comedy">Comedy</option>
                        <option value="drama">Drama</option>
                        <option value="horror">Horror</option>
                        <option value="romance">Romance</option>
                        <option value="sci-fi">Sci-Fi</option>
                        <option value="thriller">Thriller</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-2">
                      <Form.Label>Year Range</Form.Label>
                      <Form.Control
                        size="sm"
                        type="text"
                        placeholder="1990-2000"
                        value={searchFilters.year_range}
                        onChange={(e) => setSearchFilters({...searchFilters, year_range: e.target.value})}
                      />
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-2">
                      <Form.Label>Min Rating</Form.Label>
                      <Form.Range
                        min="0"
                        max="10"
                        value={searchFilters.rating_min}
                        onChange={(e) => setSearchFilters({...searchFilters, rating_min: parseInt(e.target.value)})}
                      />
                      <small>{searchFilters.rating_min}+ stars</small>
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-2">
                      <Form.Label>Language</Form.Label>
                      <Form.Select
                        size="sm"
                        value={searchFilters.language}
                        onChange={(e) => setSearchFilters({...searchFilters, language: e.target.value})}
                      >
                        <option value="">All Languages</option>
                        <option value="en">English</option>
                        <option value="es">Spanish</option>
                        <option value="fr">French</option>
                        <option value="de">German</option>
                        <option value="ja">Japanese</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                </Row>
              </Form>
            </Card.Body>
          </Card>

          {/* Suggested Queries */}
          <Card className="mb-4">
            <Card.Header as="h6">💡 Suggested Queries</Card.Header>
            <Card.Body>
              <div className="d-flex flex-wrap gap-2">
                {suggestedQueries.map((query, index) => (
                  <Button 
                    key={index}
                    variant="outline-primary" 
                    size="sm"
                    onClick={() => handleSuggestedQuery(query)}
                  >
                    {query}
                  </Button>
                ))}
              </div>
            </Card.Body>
          </Card>

          {/* Search Results */}
          {searchResults.length > 0 && (
            <Card>
              <Card.Header as="h6">
                📋 Search Results ({searchResults.length})
              </Card.Header>
              <Card.Body>
                <Row>
                  {searchResults.map((result, index) => renderSearchResult(result, index))}
                </Row>
              </Card.Body>
            </Card>
          )}

          {/* Empty State */}
          {!loading && searchResults.length === 0 && !searchQuery && (
            <Card className="text-center py-5">
              <Card.Body>
                <h4>🔍 Start Searching</h4>
                <p className="text-muted">
                  Use natural language to find movies. Try queries like:<br />
                  "action movies from the 90s" or "comedy movies with high ratings"
                </p>
              </Card.Body>
            </Card>
          )}
        </Col>
      </Row>

      {/* Movie Details Modal */}
      <Modal show={showDetailsModal} onHide={() => setShowDetailsModal(false)} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>{selectedMovie?.title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {selectedMovie && (
            <Row>
              <Col md={4}>
                <img 
                  src={selectedMovie.poster_url || '/placeholder-movie.jpg'} 
                  alt={selectedMovie.title}
                  className="img-fluid rounded"
                />
              </Col>
              <Col md={8}>
                <h5>{selectedMovie.title}</h5>
                <p className="text-muted">
                  {selectedMovie.year} • {selectedMovie.genre} • {selectedMovie.duration}
                </p>
                <p><strong>Director:</strong> {selectedMovie.director}</p>
                <p><strong>Cast:</strong> {selectedMovie.cast}</p>
                <p><strong>Rating:</strong> ⭐ {selectedMovie.rating}/10</p>
                <p><strong>Description:</strong> {selectedMovie.description}</p>
                {selectedMovie.relevance_score && (
                  <Alert variant="info">
                    <strong>Search Relevance:</strong> {Math.round(selectedMovie.relevance_score * 100)}% match
                  </Alert>
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
    </Container>
  );
};

export default NaturalLanguageSearch;