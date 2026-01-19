import ws from 'k6/ws';
import { check } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('websocket_errors');

// Test configuration
export let options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp up to 50 concurrent WebSocket connections
    { duration: '3m', target: 50 },   // Stay at 50 connections
    { duration: '1m', target: 100 },  // Ramp up to 100 connections
    { duration: '3m', target: 100 },  // Stay at 100 connections
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    websocket_errors: ['rate<0.1'],    // WebSocket error rate under 10%
  },
};

const BASE_URL = 'ws://localhost:8080';

export default function() {
  const url = `${BASE_URL}/ws`;
  const params = { tags: { ws_type: 'binary' } };

  const response = ws.connect(url, params, function(socket) {
    socket.on('open', function() {
      console.log('WebSocket connection established');
      
      // Send authentication message
      socket.send(JSON.stringify({
        type: 'auth',
        token: 'test_token_here', // You'll need to replace with actual token
      }));
      
      // Join a watch party room
      socket.send(JSON.stringify({
        type: 'join_room',
        room: 'watch_party_test',
      }));
      
      // Send chat messages periodically
      let messageCount = 0;
      const messageInterval = setInterval(() => {
        if (messageCount < 10) {
          socket.send(JSON.stringify({
            type: 'chat_message',
            room: 'watch_party_test',
            message: `Test message ${messageCount}`,
            user_id: 'test_user',
          }));
          messageCount++;
        } else {
          clearInterval(messageInterval);
        }
      }, 2000);
    });

    socket.on('message', function(message) {
      try {
        const data = JSON.parse(message);
        console.log('Received message:', data.type);
        
        check(data, {
          'message has type': (msg) => msg.type !== undefined,
          'message is valid JSON': (msg) => typeof msg === 'object',
        }) || errorRate.add(1);
      } catch (e) {
        console.log('Failed to parse message:', e);
        errorRate.add(1);
      }
    });

    socket.on('error', function(error) {
      console.log('WebSocket error:', error);
      errorRate.add(1);
    });

    socket.on('close', function() {
      console.log('WebSocket connection closed');
    });

    // Keep connection alive for test duration
    socket.setTimeout(function() {
      console.log('Closing WebSocket connection after timeout');
      socket.close();
    }, 30000); // 30 seconds
  });

  check(response, {
    'WebSocket connection status is 101': (r) => r && r.status === 101,
  }) || errorRate.add(1);
}

export function teardown() {
  console.log('WebSocket performance test completed');
}