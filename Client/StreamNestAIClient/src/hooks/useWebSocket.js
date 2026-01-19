import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from './useAuth';
import WebSocketService from '../services/websocketService';

export const useWebSocket = () => {
  const { auth } = useAuth();
  const wsRef = useRef(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionState, setConnectionState] = useState('DISCONNECTED');
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [watchPartyState, setWatchPartyState] = useState(null);
  const [chatMessages, setChatMessages] = useState([]);
  const [userStatuses, setUserStatuses] = useState({});
  const [recommendations, setRecommendations] = useState([]);

  // Create stable handler functions using useCallback
  const onConnected = useCallback(() => {
    setIsConnected(true);
    setConnectionState('CONNECTED');
    
    // Subscribe to user-specific channels if authenticated
    if (auth?.user?._id) {
      wsRef.current.subscribeToNotifications(auth.user._id);
      wsRef.current.subscribeToRecommendations(auth.user._id);
    }
  }, [auth?.user?._id]);

  const onDisconnected = useCallback(() => {
    setIsConnected(false);
    setConnectionState('DISCONNECTED');
  }, []);

  const onError = useCallback((error) => {
    console.error('WebSocket error:', error);
    setConnectionState('ERROR');
  }, []);

  const onNotification = useCallback((notification) => {
    setNotifications(prev => [notification, ...prev]);
    if (!notification.read) {
      setUnreadCount(prev => prev + 1);
    }
  }, []);

  const onWatchPartyUpdate = useCallback((update) => {
    setWatchPartyState(update);
  }, []);

  const onChatMessage = useCallback((message) => {
    setChatMessages(prev => [...prev, message]);
  }, []);

  const onUserStatus = useCallback((status) => {
    setUserStatuses(prev => ({
      ...prev,
      [status.userId]: status
    }));
  }, []);

  const onRecommendationUpdate = useCallback((recommendationData) => {
    setRecommendations(recommendationData.recommendations || []);
  }, []);

  // Initialize WebSocket service
  useEffect(() => {
    wsRef.current = WebSocketService.getInstance();
    
    // Set up event listeners with stable callback references
    const setupEventListeners = () => {
      wsRef.current.on('connected', onConnected);
      wsRef.current.on('disconnected', onDisconnected);
      wsRef.current.on('error', onError);
      wsRef.current.on('notification', onNotification);
      wsRef.current.on('watchPartyUpdate', onWatchPartyUpdate);
      wsRef.current.on('chatMessage', onChatMessage);
      wsRef.current.on('userStatus', onUserStatus);
      wsRef.current.on('recommendationUpdate', onRecommendationUpdate);
    };

    setupEventListeners();

    return () => {
      // Cleanup event listeners by removing the same callback references
      if (wsRef.current) {
        wsRef.current.off('connected', onConnected);
        wsRef.current.off('disconnected', onDisconnected);
        wsRef.current.off('error', onError);
        wsRef.current.off('notification', onNotification);
        wsRef.current.off('watchPartyUpdate', onWatchPartyUpdate);
        wsRef.current.off('chatMessage', onChatMessage);
        wsRef.current.off('userStatus', onUserStatus);
        wsRef.current.off('recommendationUpdate', onRecommendationUpdate);
        wsRef.current.disconnect();
      }
    };
  }, [onConnected, onDisconnected, onError, onNotification, onWatchPartyUpdate, onChatMessage, onUserStatus, onRecommendationUpdate]);

  // Connect to WebSocket when auth changes
  useEffect(() => {
    if (auth?.token) {
      wsRef.current.connect(auth.token);
    } else {
      wsRef.current.disconnect();
    }
  }, [auth?.token]);

  // Notification methods
  const markNotificationAsRead = useCallback((notificationId) => {
    wsRef.current.markNotificationAsRead(notificationId);
    setNotifications(prev => 
      prev.map(n => n._id === notificationId ? { ...n, read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  const markAllNotificationsAsRead = useCallback(() => {
    notifications.forEach(notification => {
      if (!notification.read) {
        wsRef.current.markNotificationAsRead(notification._id);
      }
    });
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    setUnreadCount(0);
  }, [notifications]);

  // Watch party methods
  const joinWatchParty = useCallback((partyId) => {
    if (auth?.user?._id) {
      wsRef.current.joinWatchParty(partyId, auth.user._id);
    }
  }, [auth?.user?._id]);

  const leaveWatchParty = useCallback((partyId) => {
    if (auth?.user?._id) {
      wsRef.current.leaveWatchParty(partyId, auth.user._id);
    }
  }, [auth?.user?._id]);

  const sendWatchPartyUpdate = useCallback((partyId, update) => {
    wsRef.current.sendWatchPartyUpdate(partyId, update);
  }, []);

  // Chat methods
  const joinChat = useCallback((chatId) => {
    if (auth?.user?._id) {
      wsRef.current.joinChat(chatId, auth.user._id);
    }
  }, [auth?.user?._id]);

  const leaveChat = useCallback((chatId) => {
    if (auth?.user?._id) {
      wsRef.current.leaveChat(chatId, auth.user._id);
    }
  }, [auth?.user?._id]);

  const sendChatMessage = useCallback((chatId, message) => {
    if (auth?.user?._id) {
      wsRef.current.sendChatMessage(chatId, auth.user._id, message);
    }
  }, [auth?.user?._id]);

  // User status methods
  const updateUserStatus = useCallback((status) => {
    wsRef.current.updateUserStatus(status);
  }, []);

  // Utility methods
  const getConnectionState = useCallback(() => {
    return wsRef.current?.getConnectionState() || 'DISCONNECTED';
  }, []);

  const reconnect = useCallback(() => {
    if (auth?.token) {
      wsRef.current.disconnect();
      setTimeout(() => {
        wsRef.current.connect(auth.token);
      }, 1000);
    }
  }, [auth?.token]);

  return {
    // Connection state
    isConnected,
    connectionState,
    getConnectionState,
    reconnect,
    
    // Notifications
    notifications,
    unreadCount,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    
    // Watch party
    watchPartyState,
    joinWatchParty,
    leaveWatchParty,
    sendWatchPartyUpdate,
    
    // Chat
    chatMessages,
    joinChat,
    leaveChat,
    sendChatMessage,
    setChatMessages, // Allow clearing messages
    
    // User status
    userStatuses,
    updateUserStatus,
    
    // Recommendations
    recommendations,
    
    // Raw WebSocket service for advanced usage
    wsService: wsRef.current
  };
};

export default useWebSocket;