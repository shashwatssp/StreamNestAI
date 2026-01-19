import React from 'react';
import { Spinner, Card } from 'react-bootstrap';

const LoadingSpinner = ({ message = 'Loading...', size = 'lg', fullScreen = false }) => {
  const spinnerComponent = (
    <div className={`d-flex flex-column align-items-center justify-content-center ${fullScreen ? 'min-vh-100' : 'py-5'}`}>
      <Spinner animation="border" variant="primary" size={size} className="mb-3" />
      <p className="text-muted">{message}</p>
    </div>
  );

  if (fullScreen) {
    return spinnerComponent;
  }

  return (
    <Card className="text-center p-4">
      <Card.Body>
        {spinnerComponent}
      </Card.Body>
    </Card>
  );
};

export default LoadingSpinner;