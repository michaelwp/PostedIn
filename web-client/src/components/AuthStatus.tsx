import { useEffect } from 'react';
import { useAuthStore } from '../store';

export function AuthStatus() {
  const { isAuthenticated, userId, isLoading, error, checkAuthStatus, startAuth, logout, clearError } = useAuthStore();

  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  const handleLogin = async () => {
    try {
      const authUrl = await startAuth();
      window.open(authUrl, '_blank');
    } catch (error) {
      console.error('Failed to start auth:', error);
    }
  };

  const handleLogout = async () => {
    await logout();
  };

  if (error) {
    return (
      <div className="auth-status error">
        <p>Error: {error}</p>
        <button onClick={clearError}>Dismiss</button>
      </div>
    );
  }

  return (
    <div className="auth-status">
      <div className="auth-header">
        <h3>LinkedIn Authentication</h3>
        {isLoading && <span className="loading">Loading...</span>}
      </div>
      
      {isAuthenticated ? (
        <div className="authenticated">
          <div className="user-info">
            <span className="status-indicator authenticated"></span>
            <span>Connected</span>
            {userId && <span className="user-id">User ID: {userId}</span>}
          </div>
          <button onClick={handleLogout} disabled={isLoading}>
            Logout
          </button>
        </div>
      ) : (
        <div className="not-authenticated">
          <div className="user-info">
            <span className="status-indicator not-authenticated"></span>
            <span>Not Connected</span>
          </div>
          <button onClick={handleLogin} disabled={isLoading}>
            Connect to LinkedIn
          </button>
          <p className="auth-note">
            You need to connect to LinkedIn to schedule and publish posts.
          </p>
        </div>
      )}
    </div>
  );
}