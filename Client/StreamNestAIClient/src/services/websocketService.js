class WebSocketService {
  constructor() {
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectInterval = 5000;
    this.heartbeatInterval = null;
    this.messageQueue = [];
    this.eventListeners = new Map();
    this.isConnected = false;
    this.url = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';
  }

  connect(token = null) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected');
      return;
    }

    try {
      // Build WebSocket URL with token if provided
      const wsUrl = token ? `${this.url}?token=${token}` : this.url;
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        
        // Send queued messages
        this.flushMessageQueue();
        
        // Start heartbeat
        this.startHeartbeat();
        
        // Emit connection event
        this.emit('connected', {});
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleMessage(data);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.isConnected = false;
        this.stopHeartbeat();
        
        // Emit disconnection event
        this.emit('disconnected', { code: event.code, reason: event.reason });
        
        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.emit('error', { error });
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.emit('error', { error });
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    this.isConnected = false;
    this.stopHeartbeat();
  }

  send(message) {
    if (!this.isConnected || !this.ws) {
      console.log('WebSocket not connected, queuing message');
      this.messageQueue.push(message);
      return;
    }

    try {
      this.ws.send(JSON.stringify(message));
    } catch (error) {
      console.error('Error sending WebSocket message:', error);
      this.messageQueue.push(message);
    }
  }

  handleMessage(data) {
    const { type, payload } = data;
    
    switch (type) {
      case 'notification':
        this.emit('notification', payload);
        break;
      case 'watch_party_update':
        this.emit('watchPartyUpdate', payload);
        break;
      case 'chat_message':
        this.emit('chatMessage', payload);
        break;
      case 'user_status':
        this.emit('userStatus', payload);
        break;
      case 'recommendation_update':
        this.emit('recommendationUpdate', payload);
        break;
      case 'heartbeat':
        // Respond to server heartbeat
        this.send({ type: 'heartbeat_response' });
        break;
      default:
        console.log('Unknown message type:', type);
        this.emit('message', data);
    }
  }

  // Notification methods
  subscribeToNotifications(userId) {
    this.send({
      type: 'subscribe',
      payload: { channel: `notifications:${userId}` }
    });
  }

  unsubscribeFromNotifications(userId) {
    this.send({
      type: 'unsubscribe',
      payload: { channel: `notifications:${userId}` }
    });
  }

  markNotificationAsRead(notificationId) {
    this.send({
      type: 'mark_notification_read',
      payload: { notificationId }
    });
  }

  // Watch party methods
  joinWatchParty(partyId, userId) {
    this.send({
      type: 'join_watch_party',
      payload: { partyId, userId }
    });
  }

  leaveWatchParty(partyId, userId) {
    this.send({
      type: 'leave_watch_party',
      payload: { partyId, userId }
    });
  }

  sendWatchPartyUpdate(partyId, update) {
    this.send({
      type: 'watch_party_update',
      payload: { partyId, ...update }
    });
  }

  // Chat methods
  joinChat(chatId, userId) {
    this.send({
      type: 'join_chat',
      payload: { chatId, userId }
    });
  }

  leaveChat(chatId, userId) {
    this.send({
      type: 'leave_chat',
      payload: { chatId, userId }
    });
  }

  sendChatMessage(chatId, userId, message) {
    this.send({
      type: 'chat_message',
      payload: { chatId, userId, message, timestamp: new Date().toISOString() }
    });
  }

  // User status methods
  updateUserStatus(status) {
    this.send({
      type: 'update_user_status',
      payload: { status, timestamp: new Date().toISOString() }
    });
  }

  // Recommendation methods
  subscribeToRecommendations(userId) {
    this.send({
      type: 'subscribe',
      payload: { channel: `recommendations:${userId}` }
    });
  }

  // Event listener methods
  on(event, callback) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event).push(callback);
  }

  off(event, callback) {
    if (this.eventListeners.has(event)) {
      const listeners = this.eventListeners.get(event);
      const index = listeners.indexOf(callback);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  emit(event, data) {
    if (this.eventListeners.has(event)) {
      this.eventListeners.get(event).forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error('Error in event listener:', error);
        }
      });
    }
  }

  // Utility methods
  attemptReconnect() {
    this.reconnectAttempts++;
    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, this.reconnectInterval);
  }

  flushMessageQueue() {
    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift();
      this.send(message);
    }
  }

  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        this.send({ type: 'heartbeat' });
      }
    }, 30000); // Send heartbeat every 30 seconds
  }

  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  getConnectionState() {
    if (!this.ws) return 'DISCONNECTED';
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'CONNECTING';
      case WebSocket.OPEN:
        return 'CONNECTED';
      case WebSocket.CLOSING:
        return 'CLOSING';
      case WebSocket.CLOSED:
        return 'DISCONNECTED';
      default:
        return 'UNKNOWN';
    }
  }

  // Static method to get singleton instance
  static getInstance() {
    if (!WebSocketService.instance) {
      WebSocketService.instance = new WebSocketService();
    }
    return WebSocketService.instance;
  }
}

export default WebSocketService;