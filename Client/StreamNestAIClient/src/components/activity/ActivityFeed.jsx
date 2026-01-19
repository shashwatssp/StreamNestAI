import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import useAuth from '../../hooks/useAuth';
import useWebSocket from '../../hooks/useWebSocket';
import './ActivityFeed.css';

const ActivityFeed = () => {
  const { auth } = useAuth();
  const { isConnected } = useWebSocket();
  const [activities, setActivities] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const observerRef = useRef();

  useEffect(() => {
    fetchActivities();
  }, [filter, page]);

  useEffect(() => {
    // Set up intersection observer for infinite scroll
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loading) {
          loadMore();
        }
      },
      { threshold: 0.1 }
    );

    if (observerRef.current) {
      observer.observe(observerRef.current);
    }

    return () => {
      if (observerRef.current) {
        observer.unobserve(observerRef.current);
      }
    };
  }, [hasMore, loading]);

  const fetchActivities = async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setRefreshing(true);
        setPage(1);
        setActivities([]);
      } else {
        setLoading(true);
      }

      const response = await axios.get(`/api/v1/social/activity/feed`, {
        params: {
          filter,
          page: isRefresh ? 1 : page,
          limit: 20
        },
        headers: {
          'Authorization': `Bearer ${auth.token}`
        }
      });

      const newActivities = response.data.activities || [];
      
      if (isRefresh) {
        setActivities(newActivities);
      } else {
        setActivities(prev => [...prev, ...newActivities]);
      }
      
      setHasMore(newActivities.length === 20);
      setError(null);
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to fetch activities');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const loadMore = () => {
    if (!loading && hasMore) {
      setPage(prev => prev + 1);
    }
  };

  const handleRefresh = () => {
    fetchActivities(true);
  };

  const handleLike = async (activityId) => {
    try {
      await axios.post(`/api/v1/social/activity/${activityId}/like`, {}, {
        headers: {
          'Authorization': `Bearer ${auth.token}`
        }
      });

      setActivities(prev => prev.map(activity => {
        if (activity.id === activityId) {
          return {
            ...activity,
            isLiked: !activity.isLiked,
            likesCount: activity.isLiked ? activity.likesCount - 1 : activity.likesCount + 1
          };
        }
        return activity;
      }));
    } catch (err) {
      console.error('Failed to like activity:', err);
    }
  };

  const handleComment = async (activityId, comment) => {
    try {
      const response = await axios.post(`/api/v1/social/activity/${activityId}/comment`, {
        content: comment
      }, {
        headers: {
          'Authorization': `Bearer ${auth.token}`
        }
      });

      setActivities(prev => prev.map(activity => {
        if (activity.id === activityId) {
          return {
            ...activity,
            commentsCount: activity.commentsCount + 1,
            comments: [response.data.comment, ...(activity.comments || [])]
          };
        }
        return activity;
      }));
    } catch (err) {
      console.error('Failed to comment:', err);
    }
  };

  const getActivityIcon = (type) => {
    const icons = {
      'movie_watched': '🎬',
      'movie_reviewed': '⭐',
      'movie_added_to_watchlist': '📝',
      'followed_user': '👥',
      'joined_watch_party': '🎉',
      'created_watch_party': '🎥',
      'liked_review': '❤️',
      'commented_review': '💬'
    };
    return icons[type] || '📌';
  };

  const getActivityText = (activity) => {
    const { type, user, targetMovie, targetUser, metadata } = activity;
    
    switch (type) {
      case 'movie_watched':
        return `${user.displayName || user.username} watched ${targetMovie?.title || 'a movie'}`;
      case 'movie_reviewed':
        return `${user.displayName || user.username} reviewed ${targetMovie?.title || 'a movie'}`;
      case 'movie_added_to_watchlist':
        return `${user.displayName || user.username} added ${targetMovie?.title || 'a movie'} to watchlist`;
      case 'followed_user':
        return `${user.displayName || user.username} started following ${targetUser?.displayName || targetUser?.username}`;
      case 'joined_watch_party':
        return `${user.displayName || user.username} joined a watch party for ${targetMovie?.title || 'a movie'}`;
      case 'created_watch_party':
        return `${user.displayName || user.username} created a watch party for ${targetMovie?.title || 'a movie'}`;
      case 'liked_review':
        return `${user.displayName || user.username} liked a review`;
      case 'commented_review':
        return `${user.displayName || user.username} commented on a review`;
      default:
        return `${user.displayName || user.username} performed an action`;
    }
  };

  const formatTimeAgo = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);

    if (diffInSeconds < 60) return 'just now';
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
    if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d ago`;
    return date.toLocaleDateString();
  };

  if (loading && activities.length === 0) {
    return (
      <div className="activity-feed">
        <div className="feed-header">
          <h2>Activity Feed</h2>
        </div>
        <div className="feed-loading">
          <div className="spinner"></div>
          <p>Loading activities...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="activity-feed">
      <div className="feed-header">
        <h2>Activity Feed</h2>
        <div className="feed-controls">
          <div className="connection-status">
            <div className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}></div>
            <span>{isConnected ? 'Live' : 'Offline'}</span>
          </div>
          <button
            onClick={handleRefresh}
            className={`refresh-button ${refreshing ? 'refreshing' : ''}`}
            disabled={refreshing}
          >
            🔄
          </button>
        </div>
      </div>

      <div className="feed-filters">
        <button
          className={`filter-button ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          All Activity
        </button>
        <button
          className={`filter-button ${filter === 'movies' ? 'active' : ''}`}
          onClick={() => setFilter('movies')}
        >
          Movies
        </button>
        <button
          className={`filter-button ${filter === 'social' ? 'active' : ''}`}
          onClick={() => setFilter('social')}
        >
          Social
        </button>
        <button
          className={`filter-button ${filter === 'following' ? 'active' : ''}`}
          onClick={() => setFilter('following')}
        >
          Following
        </button>
      </div>

      <div className="feed-content">
        {error && (
          <div className="feed-error">
            <p>{error}</p>
            <button onClick={handleRefresh} className="retry-button">
              Retry
            </button>
          </div>
        )}

        {activities.length === 0 && !loading && !error && (
          <div className="empty-feed">
            <div className="empty-icon">📭</div>
            <h3>No activities yet</h3>
            <p>Follow more users or watch some movies to see activity here!</p>
          </div>
        )}

        <div className="activities-list">
          {activities.map((activity) => (
            <div key={activity.id} className="activity-item">
              <div className="activity-avatar">
                <img
                  src={activity.user.avatar || `https://ui-avatars.com/api/?name=${activity.user.username}&background=4fc3f7&color=000&size=48`}
                  alt={activity.user.username}
                />
              </div>
              <div className="activity-content">
                <div className="activity-header">
                  <div className="activity-icon">
                    {getActivityIcon(activity.type)}
                  </div>
                  <div className="activity-text">
                    <p>{getActivityText(activity)}</p>
                    <span className="activity-time">{formatTimeAgo(activity.createdAt)}</span>
                  </div>
                </div>

                {activity.metadata?.reviewContent && (
                  <div className="activity-review">
                    <p>"{activity.metadata.reviewContent}"</p>
                    {activity.metadata.rating && (
                      <div className="review-rating">
                        {'⭐'.repeat(Math.floor(activity.metadata.rating))}
                      </div>
                    )}
                  </div>
                )}

                {activity.targetMovie && (
                  <div className="activity-movie">
                    <img
                      src={activity.targetMovie.poster || `https://via.placeholder.com/60x90?text=${activity.targetMovie.title}`}
                      alt={activity.targetMovie.title}
                      className="movie-thumbnail"
                    />
                    <div className="movie-info">
                      <h4>{activity.targetMovie.title}</h4>
                      <p>{activity.targetMovie.year}</p>
                    </div>
                  </div>
                )}

                <div className="activity-actions">
                  <button
                    onClick={() => handleLike(activity.id)}
                    className={`action-button ${activity.isLiked ? 'liked' : ''}`}
                  >
                    {activity.isLiked ? '❤️' : '🤍'} {activity.likesCount || 0}
                  </button>
                  <button className="action-button">
                    💬 {activity.commentsCount || 0}
                  </button>
                  <button className="action-button">
                    🔗 Share
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {loading && activities.length > 0 && (
          <div className="feed-loading-more">
            <div className="spinner"></div>
            <p>Loading more...</p>
          </div>
        )}

        {!hasMore && activities.length > 0 && (
          <div className="feed-end">
            <p>You've reached the end!</p>
          </div>
        )}

        <div ref={observerRef} className="scroll-trigger"></div>
      </div>
    </div>
  );
};

export default ActivityFeed;