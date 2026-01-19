import React, { useState, useEffect, useRef } from 'react';
import { useWebSocket } from '../../hooks/useWebSocket';
import './WatchParty.css';

const WatchParty = ({ movieId, partyId, onLeave }) => {
  const {
    watchPartyState,
    joinWatchParty,
    leaveWatchParty,
    sendWatchPartyUpdate,
    isConnected,
    userStatuses
  } = useWebSocket();

  const [isHost, setIsHost] = useState(false);
  const [participants, setParticipants] = useState([]);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [chatOpen, setChatOpen] = useState(false);
  const [chatMessage, setChatMessage] = useState('');
  const [showControls, setShowControls] = useState(false);
  const videoRef = useRef(null);
  const controlsTimeoutRef = useRef(null);

  // Initialize watch party
  useEffect(() => {
    if (partyId && isConnected) {
      joinWatchParty(partyId);
    }
    
    return () => {
      if (partyId) {
        leaveWatchParty(partyId);
      }
    };
  }, [partyId, isConnected, joinWatchParty, leaveWatchParty]);

  // Handle watch party state updates
  useEffect(() => {
    if (watchPartyState) {
      setParticipants(watchPartyState.participants || []);
      setIsHost(watchPartyState.isHost || false);
      
      // Sync video state if not host
      if (!isHost && videoRef.current) {
        if (watchPartyState.isPlaying !== undefined) {
          if (watchPartyState.isPlaying && videoRef.current.paused) {
            videoRef.current.play();
          } else if (!watchPartyState.isPlaying && !videoRef.current.paused) {
            videoRef.current.pause();
          }
        }
        
        if (watchPartyState.currentTime !== undefined) {
          const timeDiff = Math.abs(videoRef.current.currentTime - watchPartyState.currentTime);
          if (timeDiff > 2) { // Only seek if difference is significant
            videoRef.current.currentTime = watchPartyState.currentTime;
          }
        }
      }
    }
  }, [watchPartyState, isHost]);

  // Video event handlers
  const handleVideoPlay = () => {
    if (isHost) {
      setIsPlaying(true);
      sendWatchPartyUpdate(partyId, {
        isPlaying: true,
        currentTime: videoRef.current.currentTime
      });
    }
  };

  const handleVideoPause = () => {
    if (isHost) {
      setIsPlaying(false);
      sendWatchPartyUpdate(partyId, {
        isPlaying: false,
        currentTime: videoRef.current.currentTime
      });
    }
  };

  const handleVideoSeek = () => {
    if (isHost) {
      setCurrentTime(videoRef.current.currentTime);
      sendWatchPartyUpdate(partyId, {
        currentTime: videoRef.current.currentTime
      });
    }
  };

  const handleVideoTimeUpdate = () => {
    setCurrentTime(videoRef.current.currentTime);
  };

  const handleVideoLoadedMetadata = () => {
    setDuration(videoRef.current.duration);
  };

  // Control handlers
  const handlePlayPause = () => {
    if (videoRef.current) {
      if (videoRef.current.paused) {
        videoRef.current.play();
      } else {
        videoRef.current.pause();
      }
    }
  };

  const handleSeek = (time) => {
    if (videoRef.current && isHost) {
      videoRef.current.currentTime = time;
    }
  };

  const handleVolumeChange = (newVolume) => {
    if (videoRef.current) {
      videoRef.current.volume = newVolume;
      setVolume(newVolume);
    }
  };

  const handleFullscreen = () => {
    if (videoRef.current) {
      if (videoRef.current.requestFullscreen) {
        videoRef.current.requestFullscreen();
      }
    }
  };

  // Chat handlers
  const handleSendMessage = (e) => {
    e.preventDefault();
    if (chatMessage.trim()) {
      // Send chat message via WebSocket
      sendWatchPartyUpdate(partyId, {
        type: 'chat',
        message: chatMessage.trim()
      });
      setChatMessage('');
    }
  };

  // Show/hide controls
  const handleMouseMove = () => {
    setShowControls(true);
    
    if (controlsTimeoutRef.current) {
      clearTimeout(controlsTimeoutRef.current);
    }
    
    controlsTimeoutRef.current = setTimeout(() => {
      if (isPlaying) {
        setShowControls(false);
      }
    }, 3000);
  };

  const formatTime = (time) => {
    const minutes = Math.floor(time / 60);
    const seconds = Math.floor(time % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  const getParticipantStatus = (userId) => {
    return userStatuses[userId]?.status || 'online';
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'online': return '#4caf50';
      case 'away': return '#ff9800';
      case 'busy': return '#f44336';
      default: return '#9e9e9e';
    }
  };

  return (
    <div className="watch-party">
      {/* Video Container */}
      <div 
        className="video-container"
        onMouseMove={handleMouseMove}
        onMouseLeave={() => isPlaying && setShowControls(false)}
      >
        <video
          ref={videoRef}
          className="video-player"
          onPlay={handleVideoPlay}
          onPause={handleVideoPause}
          onSeeked={handleVideoSeek}
          onTimeUpdate={handleVideoTimeUpdate}
          onLoadedMetadata={handleVideoLoadedMetadata}
          controls={false}
        >
          <source src={`/api/v1/movies/${movieId}/stream`} type="video/mp4" />
          Your browser does not support the video tag.
        </video>

        {/* Video Controls */}
        <div className={`video-controls ${showControls ? 'visible' : ''}`}>
          <div className="progress-bar">
            <input
              type="range"
              min="0"
              max={duration}
              value={currentTime}
              onChange={(e) => handleSeek(parseFloat(e.target.value))}
              disabled={!isHost}
              className="progress-slider"
            />
            <div className="time-display">
              {formatTime(currentTime)} / {formatTime(duration)}
            </div>
          </div>

          <div className="control-buttons">
            <button
              onClick={handlePlayPause}
              disabled={!isHost}
              className="control-button"
              title={isHost ? 'Play/Pause' : 'Only host can control playback'}
            >
              {isPlaying ? '⏸️' : '▶️'}
            </button>

            <div className="volume-control">
              <button className="control-button">🔊</button>
              <input
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={volume}
                onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                className="volume-slider"
              />
            </div>

            <button
              onClick={handleFullscreen}
              className="control-button"
              title="Fullscreen"
            >
              ⛶
            </button>

            <button
              onClick={() => setChatOpen(!chatOpen)}
              className="control-button chat-toggle"
              title="Toggle chat"
            >
              💬
            </button>
          </div>
        </div>

        {/* Host Indicator */}
        {isHost && (
          <div className="host-indicator">
            👑 You are the host
          </div>
        )}

        {/* Connection Status */}
        {!isConnected && (
          <div className="connection-status">
            ⚠️ Connection lost. Reconnecting...
          </div>
        )}
      </div>

      {/* Participants Panel */}
      <div className="participants-panel">
        <h3>Participants ({participants.length})</h3>
        <div className="participants-list">
          {participants.map((participant) => (
            <div key={participant._id} className="participant-item">
              <div
                className="status-indicator"
                style={{ backgroundColor: getStatusColor(getParticipantStatus(participant._id)) }}
              />
              <img
                src={participant.avatar || '/default-avatar.png'}
                alt={participant.name}
                className="participant-avatar"
              />
              <span className="participant-name">
                {participant.name}
                {participant._id === watchPartyState?.hostId && ' 👑'}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Chat Panel */}
      {chatOpen && (
        <div className="chat-panel">
          <div className="chat-header">
            <h3>Party Chat</h3>
            <button
              onClick={() => setChatOpen(false)}
              className="close-chat"
            >
              ✕
            </button>
          </div>
          
          <div className="chat-messages">
            {/* Chat messages would be rendered here from WebSocket state */}
            <div className="chat-message system">
              Welcome to the watch party!
            </div>
          </div>

          <form onSubmit={handleSendMessage} className="chat-input">
            <input
              type="text"
              value={chatMessage}
              onChange={(e) => setChatMessage(e.target.value)}
              placeholder="Type a message..."
              className="message-input"
            />
            <button type="submit" className="send-button">
              Send
            </button>
          </form>
        </div>
      )}

      {/* Leave Party Button */}
      <div className="party-actions">
        <button onClick={onLeave} className="leave-party-button">
          Leave Watch Party
        </button>
      </div>
    </div>
  );
};

export default WatchParty;