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
    { duration: '2m', target: 300 }, // Ramp up to 300 users
    { duration: '5m', target: 300 }, // Stay at 300 users
    { duration: '2m', target: 500 }, // Ramp up to 500 users
    { duration: '5m', target: 500 }, // Stay at 500 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],    // Error rate under 10%
    errors: ['rate<0.1'],             // Custom error rate under 10%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data
const testUsers = [
  { email: 'test1@example.com', password: 'password123' },
  { email: 'test2@example.com', password: 'password123' },
  { email: 'test3@example.com', password: 'password123' },
];

let authToken = '';
let userId = '';

export function setup() {
  // Setup: Create test users and get auth tokens
  console.log('Setting up test environment...');
  
  // Register test users
  for (let user of testUsers) {
    const registerResponse = http.post(`${BASE_URL}/api/v1/users/register`, JSON.stringify({
      username: user.email.split('@')[0],
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (registerResponse.status !== 200 && registerResponse.status !== 409) {
      console.log(`User registration failed: ${registerResponse.status}`);
    }
  }
  
  // Login and get token
  const loginResponse = http.post(`${BASE_URL}/api/v1/users/login`, JSON.stringify({
    email: testUsers[0].email,
    password: testUsers[0].password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  // Validate login response
  if (loginResponse.status !== 200) {
    throw new Error(`Login failed with status ${loginResponse.status}: ${loginResponse.body}`);
  }
  
  const loginData = JSON.parse(loginResponse.body);
  
  // Validate token and user ID are present
  if (!loginData.token) {
    throw new Error('Login response missing token');
  }
  
  if (!loginData.user || !loginData.user.id) {
    throw new Error('Login response missing user ID');
  }
  
  authToken = loginData.token;
  userId = loginData.user.id;
  console.log('Authentication successful');
  
  return { authToken, userId };
}

export default function(data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.authToken}`,
  };

  // Test 1: Health Check
  let healthResponse = http.get(`${BASE_URL}/health`, { headers });
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Test 2: Get Movies (Original API)
  let moviesResponse = http.get(`${BASE_URL}/api/v1/movies`, { headers });
  check(moviesResponse, {
    'movies list status is 200': (r) => r.status === 200,
    'movies list response time < 300ms': (r) => r.timings.duration < 300,
    'movies list has data': (r) => JSON.parse(r.body).data && JSON.parse(r.body).data.length > 0,
  }) || errorRate.add(1);

  // Test 3: Enhanced Movies API (v2)
  let enhancedMoviesResponse = http.get(`${BASE_URL}/api/v2/movies?page=1&limit=10`, { headers });
  check(enhancedMoviesResponse, {
    'enhanced movies status is 200': (r) => r.status === 200,
    'enhanced movies response time < 400ms': (r) => r.timings.duration < 400,
    'enhanced movies has pagination': (r) => {
      const data = JSON.parse(r.body);
      return data.pagination && data.data;
    },
  }) || errorRate.add(1);

  // Test 4: Get Recommendations
  let recommendationsResponse = http.get(`${BASE_URL}/api/v2/recommendations/user/${data.userId}`, { headers });
  check(recommendationsResponse, {
    'recommendations status is 200': (r) => r.status === 200,
    'recommendations response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Test 5: Analytics Event Tracking
  let analyticsResponse = http.post(`${BASE_URL}/api/v2/analytics/events`, JSON.stringify({
    user_id: data.userId,
    event_type: 'movie_view',
    metadata: {
      movie_id: 'test_movie_id',
      duration: 120,
      timestamp: new Date().toISOString(),
    },
  }), { headers });
  check(analyticsResponse, {
    'analytics event status is 200': (r) => r.status === 200,
    'analytics event response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);

  // Test 6: Social Features - Get User Profile
  let profileResponse = http.get(`${BASE_URL}/api/v2/social/users/${data.userId}`, { headers });
  check(profileResponse, {
    'user profile status is 200': (r) => r.status === 200,
    'user profile response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 7: Get Activity Feed
  let feedResponse = http.get(`${BASE_URL}/api/v2/social/feed?page=1&limit=20`, { headers });
  check(feedResponse, {
    'activity feed status is 200': (r) => r.status === 200,
    'activity feed response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 8: Search Movies (Enhanced)
  let searchResponse = http.get(`${BASE_URL}/api/v2/search/movies?q=action&page=1&limit=10`, { headers });
  check(searchResponse, {
    'search movies status is 200': (r) => r.status === 200,
    'search movies response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Test 9: Get Analytics Dashboard
  let dashboardResponse = http.get(`${BASE_URL}/api/v2/analytics/dashboard`, { headers });
  check(dashboardResponse, {
    'analytics dashboard status is 200': (r) => r.status === 200,
    'analytics dashboard response time < 1000ms': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);

  // Test 10: Metrics Endpoint
  let metricsResponse = http.get(`${BASE_URL}/metrics`);
  check(metricsResponse, {
    'metrics endpoint status is 200': (r) => r.status === 200,
    'metrics endpoint response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);

  // Simulate user think time
  sleep(Math.random() * 3 + 1); // Random sleep between 1-4 seconds
}

export function teardown(data) {
  console.log('Test completed. Cleaning up...');
  // Cleanup logic if needed
}