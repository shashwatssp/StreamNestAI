import React from 'react';
import { Alert, Button } from 'react-bootstrap';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    // You can also log the error to an error reporting service
    console.error('Error caught by boundary:', error, errorInfo);
    this.setState({
      error: error,
      errorInfo: errorInfo
    });
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="container mt-5">
          <Alert variant="danger">
            <Alert.Heading>Oops! Something went wrong</Alert.Heading>
            <p>
              An unexpected error occurred while loading this component. 
              Please try again or contact support if the problem persists.
            </p>
            <hr />
            <div className="d-flex justify-content-between">
              <Button variant="outline-danger" onClick={this.handleRetry}>
                Try Again
              </Button>
              <Button variant="outline-secondary" onClick={() => window.location.href = '/'}>
                Go Home
              </Button>
            </div>
            {process.env.NODE_ENV === 'development' && (
              <details className="mt-3">
                <summary>Error Details</summary>
                <pre className="mt-2 p-3 bg-light">
                  {this.state.error && this.state.error.toString()}
                  <br />
                  {this.state.errorInfo?.componentStack || 'No component stack available'}
                </pre>
              </details>
            )}
          </Alert>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;