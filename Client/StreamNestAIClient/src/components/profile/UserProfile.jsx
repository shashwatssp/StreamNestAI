import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';
import useAuth from '../../hooks/useAuth';
import './UserProfile.css';

const UserProfile = () => {
  const { userId } = useParams();
  const { auth } = useAuth();
  const navigate = useNavigate();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isFollowing, setIsFollowing] = useState(false);
  const [activeTab, setActiveTab] = useState('watchlist');
  const [userStats, setUserStats] = useState(null);

  useEffect(() => {
    fetchUserProfile();
    fetchUserStats();
  }, [userId]);

  const fetchUserProfile = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`/api/v1/social/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${auth.token}`
        }
      });
      setProfile(response.data.user);
      setIsFollowing(response.data.user.isFollowing);
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to fetch user profile');
    } finally {
      setLoading(false);
    }
  };

  const fetchUserStats = async () => {
    try {
      const response = await axios.get(`/api/v1/social/users/${userId}/stats`, {
        headers: {
          'Authorization': `Bearer ${auth.token}`
        }
      });
      setUserStats(response.data);
    } catch (err) {
      console.error('Failed to fetch user stats:', err);
    }
  };

  const handleFollow = async () => {
    try {
      if (isFollowing) {
        // Unfollow - use DELETE method
        await axios.delete(`/api/v2/follow/${userId}`, {
          headers: {
            'Authorization': `Bearer ${auth.token}`
          }
        });
      } else {
        // Follow - use POST method
        await axios.post(`/api/v2/follow/${userId}`, {}, {
          headers: {
            'Authorization': `Bearer ${auth.token}`
          }
        });
      }
      setIsFollowing(!isFollowing);
      fetchUserStats(); // Refresh stats
    } catch (err) {
      console.error('Failed to follow/unfollow:', err);
    }
  };

  const handleSendMessage = () => {
    // Navigate to messages or open message modal
    navigate(`/messages/${userId}`);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="user-profile">
        <div className="profile-loading">
          <div className="spinner"></div>
          <p>Loading profile...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="user-profile">
        <div className="profile-error">
          <p>{error}</p>
          <button onClick={() => navigate(-1)} className="back-button">
            Go Back
          </button>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="user-profile">
        <div className="profile-error">
          <p>User not found</p>
          <button onClick={() => navigate(-1)} className="back-button">
            Go Back
          </button>
        </div>
      </div>
    );
  }

  const isOwnProfile = auth.user?.id === userId;

  return (
    <div className="user-profile">
      <div className="profile-header">
        <div className="profile-cover">
          <div className="cover-image"></div>
          <div className="profile-info">
            <div className="profile-avatar">
              <img
                src={profile.avatar || `https://ui-avatars.com/api/?name=${profile.username}&background=4fc3f7&color=000&size=128`}
                alt={profile.username}
              />
              {profile.isOnline && <div className="online-indicator"></div>}
            </div>
            <div className="profile-details">
              <h1 className="profile-name">{profile.displayName || profile.username}</h1>
              <p className="profile-username">@{profile.username}</p>
              <p className="profile-bio">{profile.bio || 'No bio available'}</p>
              <div className="profile-meta">
                <span className="join-date">
                  Joined {formatDate(profile.createdAt)}
                </span>
                {profile.location && (
                  <span className="location">📍 {profile.location}</span>
                )}
                {profile.website && (
                  <a href={profile.website} target="_blank" rel="noopener noreferrer" className="website">
                    🔗 {profile.website}
                  </a>
                )}
              </div>
            </div>
            <div className="profile-actions">
              {!isOwnProfile && (
                <>
                  <button
                    onClick={handleFollow}
                    className={`follow-button ${isFollowing ? 'following' : 'follow'}`}
                  >
                    {isFollowing ? 'Following' : 'Follow'}
                  </button>
                  <button
                    onClick={handleSendMessage}
                    className="message-button"
                  >
                    Message
                  </button>
                </>
              )}
              {isOwnProfile && (
                <button
                  onClick={() => navigate('/settings/profile')}
                  className="edit-profile-button"
                >
                  Edit Profile
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="profile-stats">
        <div className="stat-item">
          <div className="stat-value">{userStats?.followersCount || 0}</div>
          <div className="stat-label">Followers</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{userStats?.followingCount || 0}</div>
          <div className="stat-label">Following</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{userStats?.moviesWatched || 0}</div>
          <div className="stat-label">Movies Watched</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{userStats?.reviewsCount || 0}</div>
          <div className="stat-label">Reviews</div>
        </div>
        <div className="stat-item">
          <div className="stat-value">{userStats?.watchlistCount || 0}</div>
          <div className="stat-label">Watchlist</div>
        </div>
      </div>

      <div className="profile-content">
        <div className="content-tabs">
          <button
            className={`tab-button ${activeTab === 'watchlist' ? 'active' : ''}`}
            onClick={() => setActiveTab('watchlist')}
          >
            Watchlist
          </button>
          <button
            className={`tab-button ${activeTab === 'watched' ? 'active' : ''}`}
            onClick={() => setActiveTab('watched')}
          >
            Watched
          </button>
          <button
            className={`tab-button ${activeTab === 'reviews' ? 'active' : ''}`}
            onClick={() => setActiveTab('reviews')}
          >
            Reviews
          </button>
          <button
            className={`tab-button ${activeTab === 'followers' ? 'active' : ''}`}
            onClick={() => setActiveTab('followers')}
          >
            Followers
          </button>
          <button
            className={`tab-button ${activeTab === 'following' ? 'active' : ''}`}
            onClick={() => setActiveTab('following')}
          >
            Following
          </button>
        </div>

        <div className="content-area">
          {activeTab === 'watchlist' && (
            <div className="movie-grid">
              {profile.watchlist?.length > 0 ? (
                profile.watchlist.map((movie) => (
                  <div key={movie.id} className="movie-card">
                    <img
                      src={movie.poster || `https://via.placeholder.com/200x300?text=${movie.title}`}
                      alt={movie.title}
                      className="movie-poster"
                    />
                    <div className="movie-info">
                      <h4 className="movie-title">{movie.title}</h4>
                      <p className="movie-year">{movie.year}</p>
                    </div>
                  </div>
                ))
              ) : (
                <div className="empty-state">
                  <p>No movies in watchlist</p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'watched' && (
            <div className="movie-grid">
              {profile.watchedMovies?.length > 0 ? (
                profile.watchedMovies.map((movie) => (
                  <div key={movie.id} className="movie-card">
                    <img
                      src={movie.poster || `https://via.placeholder.com/200x300?text=${movie.title}`}
                      alt={movie.title}
                      className="movie-poster"
                    />
                    <div className="movie-info">
                      <h4 className="movie-title">{movie.title}</h4>
                      <p className="movie-year">{movie.year}</p>
                      {movie.rating && (
                        <div className="movie-rating">
                          ⭐ {movie.rating.toFixed(1)}
                        </div>
                      )}
                    </div>
                  </div>
                ))
              ) : (
                <div className="empty-state">
                  <p>No watched movies</p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'reviews' && (
            <div className="reviews-list">
              {profile.reviews?.length > 0 ? (
                profile.reviews.map((review) => (
                  <div key={review.id} className="review-card">
                    <div className="review-header">
                      <h4 className="review-movie-title">{review.movieTitle}</h4>
                      <div className="review-rating">
                        {'⭐'.repeat(Math.floor(review.rating))}
                      </div>
                    </div>
                    <p className="review-content">{review.content}</p>
                    <p className="review-date">{formatDate(review.createdAt)}</p>
                  </div>
                ))
              ) : (
                <div className="empty-state">
                  <p>No reviews yet</p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'followers' && (
            <div className="users-list">
              {profile.followers?.length > 0 ? (
                profile.followers.map((user) => (
                  <div key={user.id} className="user-card">
                    <img
                      src={user.avatar || `https://ui-avatars.com/api/?name=${user.username}&background=4fc3f7&color=000&size=48`}
                      alt={user.username}
                      className="user-avatar"
                    />
                    <div className="user-info">
                      <h4 className="user-name">{user.displayName || user.username}</h4>
                      <p className="user-username">@{user.username}</p>
                    </div>
                    <button
                      onClick={() => navigate(`/profile/${user.id}`)}
                      className="view-profile-button"
                    >
                      View Profile
                    </button>
                  </div>
                ))
              ) : (
                <div className="empty-state">
                  <p>No followers yet</p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'following' && (
            <div className="users-list">
              {profile.following?.length > 0 ? (
                profile.following.map((user) => (
                  <div key={user.id} className="user-card">
                    <img
                      src={user.avatar || `https://ui-avatars.com/api/?name=${user.username}&background=4fc3f7&color=000&size=48`}
                      alt={user.username}
                      className="user-avatar"
                    />
                    <div className="user-info">
                      <h4 className="user-name">{user.displayName || user.username}</h4>
                      <p className="user-username">@{user.username}</p>
                    </div>
                    <button
                      onClick={() => navigate(`/profile/${user.id}`)}
                      className="view-profile-button"
                    >
                      View Profile
                    </button>
                  </div>
                ))
              ) : (
                <div className="empty-state">
                  <p>Not following anyone yet</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default UserProfile;