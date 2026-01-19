import axiosClient from './axiosConfig';

// Social Features API - Only endpoints that exist in server
export const socialApi = {
  // Get user profile
  getUserProfile: (userId) => axiosClient.get(`/api/v2/profile/${userId}`),
  
  // Update user profile
  updateProfile: (profileData) => axiosClient.put('/api/v2/profile', profileData),
  
  // Follow user
  followUser: (userId) => axiosClient.post(`/api/v2/follow/${userId}`),
  
  // Unfollow user
  unfollowUser: (userId) => axiosClient.delete(`/api/v2/follow/${userId}`),
  
  // Get activity feed
  getActivityFeed: (limit = 20, offset = 0) =>
    axiosClient.get(`/api/v2/activity/feed?limit=${limit}&offset=${offset}`),
  
  // Get social stats
  getSocialStats: () => axiosClient.get('/api/v2/social/stats'),
  
  // Get followers
  getFollowers: (userId) => axiosClient.get(`/api/v2/social/followers/${userId}`),
  
  // Get following
  getFollowing: (userId) => axiosClient.get(`/api/v2/social/following/${userId}`),
  
  // Search users
  searchUsers: (query) => axiosClient.get(`/api/v2/social/search?q=${encodeURIComponent(query)}`),
  
  // Create activity
  createActivity: (activityData) => axiosClient.post('/api/v2/social/activity', activityData),
  
  // Security audit log (available endpoint)
  getSecurityAuditLog: () => axiosClient.get('/api/v2/security/audit-log'),
};

// Watchlist API - Only endpoints that exist in server
export const watchlistApi = {
  // Get user watchlists
  getUserWatchlists: (userId) => axiosClient.get(`/api/v2/watchlist/user/${userId}`),
  
  // Get public watchlists
  getPublicWatchlists: () => axiosClient.get('/api/v2/watchlist/public'),
  
  // Get trending watchlists
  getTrendingWatchlists: () => axiosClient.get('/api/v2/watchlist/trending'),
  
  // Add movie to watchlist
  addMovieToWatchlist: (movieId) => axiosClient.post(`/api/v2/watchlist/add/${movieId}`),
  
  // Remove movie from watchlist
  removeMovieFromWatchlist: (movieId) => axiosClient.delete(`/api/v2/watchlist/remove/${movieId}`),
};

// Recommendations API - Only endpoints that exist in server
export const recommendationsApi = {
  // Get personalized recommendations
  getPersonalizedRecommendations: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/personalized?${queryParams}`);
  },
  
  // Get trending content (public endpoint)
  getTrendingContent: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/trending?${queryParams}`);
  },
  
  // Get similar movies
  getSimilarMovies: (movieId, params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/similar/${movieId}?${queryParams}`);
  },
  
  // Get collaborative recommendations
  getCollaborativeRecommendations: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/collaborative?${queryParams}`);
  },
  
  // Get content-based recommendations
  getContentBasedRecommendations: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/content-based?${queryParams}`);
  },
  
  // Get genre-based recommendations
  getGenreBasedRecommendations: (genre, params = {}) => {
    const queryParams = new URLSearchParams({ genre, ...params }).toString();
    return axiosClient.get(`/api/v2/recommendations/genre?${queryParams}`);
  },
  
  // Get mood-based recommendations
  getMoodBasedRecommendations: (mood, params = {}) => {
    const queryParams = new URLSearchParams({ mood, ...params }).toString();
    return axiosClient.get(`/api/v2/recommendations/mood?${queryParams}`);
  },
  
  // Get seasonal recommendations
  getSeasonalRecommendations: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/recommendations/seasonal?${queryParams}`);
  },
  
  // Get recommendation preferences
  getRecommendationPreferences: () => {
    return axiosClient.get('/api/v2/recommendations/preferences');
  },
  
  // Update recommendation preferences
  updateRecommendationPreferences: (preferences) => {
    return axiosClient.put('/api/v2/recommendations/preferences', preferences);
  },
  
  // Provide recommendation feedback
  provideRecommendationFeedback: (feedback) => {
    return axiosClient.post('/api/v2/recommendations/feedback', feedback);
  },
  
  // Search recommendations
  searchRecommendations: (query, params = {}) => {
    const queryParams = new URLSearchParams({ q: query, ...params }).toString();
    return axiosClient.get(`/api/v2/recommendations/search?${queryParams}`);
  },
  
  // Legacy methods for backward compatibility
  getPersonalized: (params = {}) => {
    return recommendationsApi.getPersonalizedRecommendations(params);
  },
  
  getTrending: (params = {}) => {
    return recommendationsApi.getTrendingContent(params);
  },
};

// Analytics API - Only endpoints that exist in server
export const analyticsApi = {
  // Track event
  trackEvent: (eventData) => axiosClient.post('/api/v2/analytics/track', eventData),
  
  // Get overview (public endpoint)
  getOverview: () => axiosClient.get('/api/v2/analytics/overview'),
  
  // Get real-time metrics
  getRealTimeMetrics: () => axiosClient.get('/api/v2/analytics/realtime'),
  
  // Get user analytics
  getUserAnalytics: (userId, params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/analytics/users/${userId}?${queryParams}`);
  },
  
  // Get content analytics
  getContentAnalytics: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/analytics/content?${queryParams}`);
  },
  
  // Get engagement metrics
  getEngagementMetrics: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/analytics/engagement?${queryParams}`);
  },
  
  // Get conversion metrics
  getConversionMetrics: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/analytics/conversions?${queryParams}`);
  },
};

// Monitoring API - Only endpoints that exist in server
export const monitoringApi = {
  // Get health status (public endpoint)
  getHealth: () => axiosClient.get('/api/v2/health'),
  
  // Get system health
  getSystemHealth: () => axiosClient.get('/api/v2/monitoring/health'),
  
  // Get performance metrics
  getPerformanceMetrics: () => axiosClient.get('/api/v2/monitoring/performance'),
  
  // Get error logs
  getErrorLogs: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/monitoring/errors?${queryParams}`);
  },
  
  // Get API metrics
  getApiMetrics: () => axiosClient.get('/api/v2/monitoring/api'),
  
  // Get database stats
  getDatabaseStats: () => axiosClient.get('/api/v2/monitoring/database'),
  
  // Get cache stats
  getCacheStats: () => axiosClient.get('/api/v2/monitoring/cache'),
  
  // Get server metrics
  getServerMetrics: () => axiosClient.get('/api/v2/monitoring/server'),
};

// Admin API - Only endpoints that exist in server
export const adminApi = {
  // Get admin dashboard
  getDashboard: () => axiosClient.get('/api/v2/admin/dashboard'),
  
  // Get users
  getUsers: () => axiosClient.get('/api/v2/admin/users'),
};

// Natural Language Search API - Only endpoints that exist in server
export const searchApi = {
  // Get search history
  getSearchHistory: () => axiosClient.get('/api/v2/search/history'),
  
  // Natural language search
  naturalLanguageSearch: (query, genre = '', year = '', rating = '', sort = 'relevance') => {
    const params = new URLSearchParams({
      q: query,
      genre: genre,
      year: year,
      rating: rating,
      sort: sort
    }).toString();
    return axiosClient.get(`/api/v2/search/natural?${params}`);
  },
  
  // Advanced search
  advancedSearch: (query, genre = '', year = '', rating = '', sort = 'relevance') => {
    const params = new URLSearchParams({
      q: query,
      genre: genre,
      year: year,
      rating: rating,
      sort: sort
    }).toString();
    return axiosClient.get(`/api/v2/search/advanced?${params}`);
  },
  
  // Semantic search
  semanticSearch: (query, genre = '', year = '', rating = '', sort = 'relevance') => {
    const params = new URLSearchParams({
      q: query,
      genre: genre,
      year: year,
      rating: rating,
      sort: sort
    }).toString();
    return axiosClient.get(`/api/v2/search/semantic?${params}`);
  },
  
  // Enhanced movies search (available endpoint)
  searchMovies: (query, genre = '', year = '', sort = 'relevance') => {
    const params = new URLSearchParams({
      q: query,
      genre: genre,
      year: year,
      sort: sort
    }).toString();
    return axiosClient.get(`/api/v2/movies/search?${params}`);
  },
};

// Enhanced Movie API - Only endpoints that exist in server
export const movieApi = {
  // Get movies with filters
  getMovies: (params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/movies?${queryParams}`);
  },
  
  // Get movie details
  getMovie: (imdbId, params = {}) => {
    const queryParams = new URLSearchParams(params).toString();
    return axiosClient.get(`/api/v2/movies/${imdbId}?${queryParams}`);
  },
  
  // Search movies
  searchMovies: (query, genre = '', year = '', sort = 'relevance') => {
    const params = new URLSearchParams({
      q: query,
      genre: genre,
      year: year,
      sort: sort
    }).toString();
    return axiosClient.get(`/api/v2/movies/search?${params}`);
  },
};

// WebSocket API (for real-time features)
export const wsApi = {
  // Connect to WebSocket
  connect: (token) => {
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
    return new WebSocket(wsUrl, [], {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
  },
  
  // Send message
  sendMessage: (ws, message) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message));
    }
  },
};

export default {
  social: socialApi,
  watchlist: watchlistApi,
  recommendations: recommendationsApi,
  analytics: analyticsApi,
  monitoring: monitoringApi,
  admin: adminApi,
  search: searchApi,
  movie: movieApi,
  ws: wsApi,
};