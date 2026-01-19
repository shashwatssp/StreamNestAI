import React, { useState, useEffect, useRef } from 'react';
import { useWebSocket } from '../../hooks/useWebSocket';
import './NotificationCenter.css';

const NotificationCenter = () => {
  const {
    notifications,
    unreadCount,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    isConnected
  } = useWebSocket();

  const [isOpen, setIsOpen] = useState(false);
  const [filter, setFilter] = useState('all'); // all, unread, read
  const notificationRef = useRef(null);

  // Close notifications when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (notificationRef.current && !notificationRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const filteredNotifications = notifications.filter(notification => {
    if (filter === 'unread') return !notification.read;
    if (filter === 'read') return notification.read;
    return true;
  });

  const getNotificationIcon = (type) => {
    switch (type) {
      case 'follow':
        return '👤';
      case 'like':
        return '❤️';
      case 'comment':
        return '💬';
      case 'mention':
        return '@';
      case 'watch_party':
        return '🎬';
      case 'recommendation':
        return '⭐';
      case 'system':
        return '🔔';
      default:
        return '📢';
    }
  };

  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  };

  const handleNotificationClick = (notification) => {
    if (!notification.read) {
      markNotificationAsRead(notification._id);
    }
    
    // Handle navigation based on notification type
    if (notification.actionUrl) {
      window.location.href = notification.actionUrl;
    }
    
    setIsOpen(false);
  };

  return (
    <div className="notification-center" ref={notificationRef}>
      {/* Notification Bell */}
      <button
        className={`notification-bell ${isOpen ? 'open' : ''} ${!isConnected ? 'disconnected' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        title={isConnected ? 'Notifications' : 'Connection lost'}
      >
        <span className="bell-icon">🔔</span>
        {unreadCount > 0 && (
          <span className="notification-badge">{unreadCount > 99 ? '99+' : unreadCount}</span>
        )}
        {!isConnected && <span className="connection-indicator">⚠️</span>}
      </button>

      {/* Notification Dropdown */}
      {isOpen && (
        <div className="notification-dropdown">
          <div className="notification-header">
            <h3>Notifications</h3>
            <div className="notification-actions">
              {unreadCount > 0 && (
                <button
                  className="mark-all-read"
                  onClick={markAllNotificationsAsRead}
                >
                  Mark all read
                </button>
              )}
              <button
                className="close-button"
                onClick={() => setIsOpen(false)}
              >
                ✕
              </button>
            </div>
          </div>

          {/* Filter Tabs */}
          <div className="notification-filters">
            <button
              className={`filter-tab ${filter === 'all' ? 'active' : ''}`}
              onClick={() => setFilter('all')}
            >
              All ({notifications.length})
            </button>
            <button
              className={`filter-tab ${filter === 'unread' ? 'active' : ''}`}
              onClick={() => setFilter('unread')}
            >
              Unread ({unreadCount})
            </button>
            <button
              className={`filter-tab ${filter === 'read' ? 'active' : ''}`}
              onClick={() => setFilter('read')}
            >
              Read
            </button>
          </div>

          {/* Notification List */}
          <div className="notification-list">
            {filteredNotifications.length === 0 ? (
              <div className="no-notifications">
                <p>No notifications</p>
                {filter === 'unread' && (
                  <button onClick={() => setFilter('all')}>
                    View all notifications
                  </button>
                )}
              </div>
            ) : (
              filteredNotifications.map((notification) => (
                <div
                  key={notification._id}
                  className={`notification-item ${!notification.read ? 'unread' : ''}`}
                  onClick={() => handleNotificationClick(notification)}
                >
                  <div className="notification-icon">
                    {getNotificationIcon(notification.type)}
                  </div>
                  <div className="notification-content">
                    <div className="notification-message">
                      {notification.message}
                    </div>
                    <div className="notification-meta">
                      <span className="notification-time">
                        {formatTime(notification.timestamp)}
                      </span>
                      {notification.sender && (
                        <span className="notification-sender">
                          from {notification.sender.name}
                        </span>
                      )}
                    </div>
                  </div>
                  {!notification.read && (
                    <div className="unread-indicator" />
                  )}
                </div>
              ))
            )}
          </div>

          {/* Connection Status */}
          {!isConnected && (
            <div className="connection-status">
              <span>⚠️ Connection lost. Reconnecting...</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default NotificationCenter;