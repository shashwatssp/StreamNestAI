import { useParams } from 'react-router-dom';
import './StreamMovie.css';

const StreamMovie = () => {
  const { yt_id } = useParams();

  if (!yt_id) {
    return <div style={{ color: 'white', padding: '20px' }}>No video found</div>;
  }

  return (
    <div className="video-container">
      <iframe
        width="100%"
        height="100%"
        src={`https://www.youtube.com/embed/${yt_id}`}
        title="YouTube video player"
        frameBorder="0"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
      />
    </div>
  );
};

export default StreamMovie;
