import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('redis_errors');
export let cacheHitRate = new Rate('cache_hits');

// Test configuration
export let options = {
  stages: [
    { duration: '1m', target: 200 },  // Ramp up to 200 users
    { duration: '3m', target: 200 },  // Stay at 200 users
    { duration: '1m', target: 500 },  // Ramp up to 500 users
    { duration: '3m', target: 500 },  // Stay at 500 users
    { duration: '1m', target: 1000 }, // Ramp up to 1000 users
    { duration: '3m', target: 1000 }, // Stay at 1000 users
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 95% of requests under 300ms
    http_req_failed: ['rate<0.05'],   // Error rate under 5%
    redis_errors: ['rate<0.05'],      // Redis error rate under 5%
    cache_hits: ['rate>0.7'],         // Cache hit rate above 70%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data
const movieIds = [
  '503a16d82f4e953d39a1e503',
  '503a16d82f4e953d39a1e504',
  '503a16d82f4e953d39a1e505',
  '503a16d82f4e953d39a1e506',
  '503a16d82f4e953d39a1e507',
];

let authToken = '';

export function setup() {
  console.log('Setting up Redis stress test...');
  
  // Login to get auth token
  const loginResponse = http.post(`${BASE_URL}/api/v1/users/login`, JSON.stringify({
    email: 'test1@example.com',
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (loginResponse.status === 200) {
    const loginData = JSON.parse(loginResponse.body);
    authToken = loginData.token;
    console.log('Authentication successful for Redis stress test');
  }
  
  return { authToken };
}

export default function(data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.authToken}`,
  };

  // Test 1: Cache Movie Details (should hit cache after first request)
  const movieId = movieIds[Math.floor(Math.random() * movieIds.length)];
  let movieResponse = http.get(`${BASE_URL}/api/v2/movies/${movieId}`, { headers });
  
  const isCacheHit = movieResponse.headers['X-Cache'] === 'HIT';
  if (isCacheHit) {
    cacheHitRate.add(1);
  }
  
  check(movieResponse, {
    'movie details status is 200': (r) => r.status === 200,
    'movie details response time < 200ms': (r) => r.timings.duration < 200,
    'movie details has cache header': (r) => r.headers['X-Cache'] !== undefined,
  }) || errorRate.add(1);

  // Test 2: Rate Limiting Test (should not be rate limited under normal load)
  let rateLimitResponse = http.get(`${BASE_URL}/api/v2/movies`, { headers });
  check(rateLimitResponse, {
    'rate limit test status is 200': (r) => r.status === 200,
    'rate limit test not blocked': (r) => r.status !== 429,
    'rate limit headers present': (r) => {
      return r.headers['X-RateLimit-Limit'] && 
             r.headers['X-RateLimit-Remaining'] && 
             r.headers['X-RateLimit-Reset'];
    },
  }) || errorRate.add(1);

  // Test 3: Concurrent Cache Access (multiple requests to same resource)
  const concurrentRequests = Array(5).fill(0).map(() => 
    http.get(`${BASE_URL}/api/v2/movies/${movieId}`, { headers })
  );
  
  concurrentRequests.forEach((response, index) => {
    check(response, {
      [`concurrent request ${index} status is 200`]: (r) => r.status === 200,
      [`concurrent request ${index} response time < 300ms`]: (r) => r.timings.duration < 300,
    }) || errorRate.add(1);
  });

  // Test 4: Analytics Events (tests Redis pub/sub)
  let analyticsResponse = http.post(`${BASE_URL}/api/v2/analytics/events`, JSON.stringify({
    user_id: 'test_user_id',
    event_type: 'movie_view',
    metadata: {
      movie_id: movieId,
      duration: Math.floor(Math.random() * 180) + 60, // 60-240 seconds
      timestamp: new Date().toISOString(),
      quality: 'HD',
    },
  }), { headers });
  
  check(analyticsResponse, {
    'analytics event status is 200': (r) => r.status === 200,
    'analytics event response time < 150ms': (r) => r.timings.duration < 150,
  }) || errorRate.add(1);

  // Test 5: Recommendation Cache Test
  let recommendationResponse = http.get(`${BASE_URL}/api/v2/recommendations/user/test_user_id`, { headers });
  check(recommendationResponse, {
    'recommendations status is 200': (r) => r.status === 200,
    'recommendations response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 6: Search with Caching
  let searchResponse = http.get(`${BASE_URL}/api/v2/search/movies?q=action&page=1&limit=10`, { headers });
  check(searchResponse, {
    'search status is 200': (r) => r.status === 200,
    'search response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Test 7: Social Feed Cache Test
  let feedResponse = http.get(`${BASE_URL}/api/v2/social/feed?page=1&limit=20`, { headers });
  check(feedResponse, {
    'social feed status is 200': (r) => r.status === 200,
    'social feed response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 8: Dashboard Analytics Cache Test
  let dashboardResponse = http.get(`${BASE_URL}/api/v2/analytics/dashboard`, { headers });
  check(dashboardResponse, {
    'dashboard status is 200': (r) => r.status === 200,
    'dashboard response time < 800ms': (r) => r.timings.duration < 800,
  }) || errorRate.add(1);

  // Simulate realistic user behavior
  sleep(Math.random() * 2 + 0.5); // Random sleep between 0.5-2.5 seconds
}

export function teardown(data) {
  console.log('Redis stress test completed');
  console.log(`Cache hit rate: ${(cacheHitRate.rate * 100).toFixed(2)}%`);
  console.log(`Error rate: ${(errorRate.rate * 100).toFixed(2)}%`);
}

// Custom function to test rate limiting boundaries
export function handleRateLimitBoundary() {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Send rapid requests to test rate limiting
  for (let i = 0; i < 150; i++) {
    let response = http.get(`${BASE_URL}/api/v2/movies`, { headers });
    
    if (response.status === 429) {
      console.log(`Rate limit triggered after ${i} requests`);
      check(response, {
        'rate limit response has retry-after header': (r) => r.headers['Retry-After'] !== undefined,
        'rate limit response has proper status': (r) => r.status === 429,
      });
      break;
    }
  }
}