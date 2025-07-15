import { useEffect } from 'react';
import { useSchedulerStore } from '../store';

export function SchedulerStatus() {
  const { status, isLoading, error, fetchStatus, startScheduler, stopScheduler, clearError } = useSchedulerStore();

  useEffect(() => {
    fetchStatus();
    // Refresh status every 30 seconds
    const interval = setInterval(fetchStatus, 30000);
    return () => clearInterval(interval);
  }, [fetchStatus]);

  const handleToggleScheduler = async () => {
    if (status?.running) {
      await stopScheduler();
    } else {
      await startScheduler();
    }
  };

  if (error) {
    return (
      <div className="scheduler-status error">
        <p>Error: {error}</p>
        <button onClick={clearError}>Dismiss</button>
      </div>
    );
  }

  return (
    <div className="scheduler-status">
      <div className="status-header">
        <h3>Auto-Scheduler</h3>
        <button 
          onClick={handleToggleScheduler}
          disabled={isLoading}
          className={status?.running ? 'stop-btn' : 'start-btn'}
        >
          {isLoading ? 'Loading...' : status?.running ? 'Stop' : 'Start'}
        </button>
      </div>
      
      {status && (
        <div className="status-details">
          <div className="status-item">
            <span className="label">Status:</span>
            <span className={`value ${status.running ? 'running' : 'stopped'}`}>
              {status.running ? 'Running' : 'Stopped'}
            </span>
          </div>
          
          <div className="status-item">
            <span className="label">Scheduled Posts:</span>
            <span className="value">{status.posts_count}</span>
          </div>
          
          {status.next_execution && (
            <div className="status-item">
              <span className="label">Next Execution:</span>
              <span className="value">{new Date(status.next_execution).toLocaleString()}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}