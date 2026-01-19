import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('errors');

// Test configuration
export let options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 200 }, // Ramp up to 200 users
    { duration: '5m', target: 200 }, // Stay at 200 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],    // Error rate under 10%
    errors: ['rate<0.1'],             // Custom error rate under 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test data
const users = [
  { email: 'test1@example.com', password: 'password123' },
  { email: 'test2@example.com', password: 'password123' },
  { email: 'test3@example.com', password: 'password123' },
];

export function setup() {
  // Setup - create test users and login
  let tokens = [];
  
  users.forEach(user => {
    let registerResponse = http.post(`${BASE_URL}/register`, JSON.stringify({
      username: user.email.split('@')[0],
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    let loginResponse = http.post(`${BASE_URL}/login`, JSON.stringify({
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (loginResponse.status === 200) {
      tokens.push(loginResponse.json('token'));
    }
  });
  
  return { tokens };
}

export default function(data) {
  // Pick a random user token
  let token = data.tokens[Math.floor(Math.random() * data.tokens.length)];
  let headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  // Test 1: Health check
  let healthResponse = http.get(`${BASE_URL}/health`, { headers });
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Test 2: Get movies
  let moviesResponse = http.get(`${BASE_URL}/movies`, { headers });
  check(moviesResponse, {
    'movies status is 200': (r) => r.status === 200,
    'movies response time < 300ms': (r) => r.timings.duration < 300,
    'movies has data': (r) => r.json().length > 0,
  }) || errorRate.add(1);

  // Test 3: Get specific movie
  if (moviesResponse.json().length > 0) {
    let movieId = moviesResponse.json()[0]._id;
    let movieResponse = http.get(`${BASE_URL}/movies/${movieId}`, { headers });
    check(movieResponse, {
      'movie detail status is 200': (r) => r.status === 200,
      'movie detail response time < 200ms': (r) => r.timings.duration < 200,
    }) || errorRate.add(1);
  }

  // Test 4: Search movies
  let searchResponse = http.get(`${BASE_URL}/movies/search?query=action`, { headers });
  check(searchResponse, {
    'search status is 200': (r) => r.status === 200,
    'search response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 5: Get recommendations
  let recommendationsResponse = http.get(`${BASE_URL}/recommendations/personalized`, { headers });
  check(recommendationsResponse, {
    'recommendations status is 200': (r) => r.status === 200,
    'recommendations response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Test 6: Get user profile
  let profileResponse = http.get(`${BASE_URL}/profile/me`, { headers });
  check(profileResponse, {
    'profile status is 200': (r) => r.status === 200,
    'profile response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 7: Get activity feed
  let activityResponse = http.get(`${BASE_URL}/activity/feed?limit=20`, { headers });
  check(activityResponse, {
    'activity feed status is 200': (r) => r.status === 200,
    'activity feed response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 8: Analytics tracking
  let analyticsResponse = http.post(`${BASE_URL}/analytics/track`, JSON.stringify({
    event: 'movie_view',
    properties: {
      movie_id: 'test_movie_id',
      duration: 120,
    },
  }), { headers });
  check(analyticsResponse, {
    'analytics tracking status is 200': (r) => r.status === 200,
    'analytics tracking response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);

  // Test 9: Metrics endpoint
  let metricsResponse = http.get(`${BASE_URL}/metrics`, { headers });
  check(metricsResponse, {
    'metrics status is 200': (r) => r.status === 200,
    'metrics response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  sleep(1);
}

export function teardown(data) {
  // Cleanup - remove test users if needed
  console.log('Load test completed');
}