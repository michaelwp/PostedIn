import { useState, useEffect } from 'react';
import { useTimezoneStore } from '../store';

const getSystemTimezone = () => {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
};

const commonTimezones = [
  'System',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'Europe/London',
  'Europe/Berlin',
  'Europe/Paris',
  'Asia/Tokyo',
  'Asia/Singapore',
  'Asia/Jakarta',
  'Asia/Bangkok',
  'Australia/Sydney',
  'UTC',
];

export function TimezoneSettings() {
  const { timezone, isLoading, error, fetchTimezone, setTimezone, clearError } = useTimezoneStore();
  const [selectedTimezone, setSelectedTimezone] = useState('');

  useEffect(() => {
    fetchTimezone();
  }, [fetchTimezone]);

  useEffect(() => {
    if (timezone) {
      // Handle the actual API response structure
      const currentTz = timezone.location || timezone.timezone;
      if (currentTz === getSystemTimezone()) {
        setSelectedTimezone('System');
      } else {
        setSelectedTimezone(currentTz);
      }
    }
  }, [timezone]);

  const handleTimezoneChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedTimezone) {
      const actualTimezone = selectedTimezone === 'System' ? getSystemTimezone() : selectedTimezone;
      const currentTz = timezone?.location || timezone?.timezone;
      if (actualTimezone !== currentTz) {
        await setTimezone(actualTimezone);
      }
    }
  };

  if (error) {
    return (
      <div className="timezone-settings error">
        <p>Error: {error}</p>
        <button onClick={clearError}>Dismiss</button>
      </div>
    );
  }

  return (
    <div className="timezone-settings">
      <h3>Timezone Settings</h3>
      
      {timezone && (
        <div className="current-timezone">
          <p><strong>Current Timezone:</strong> {timezone.info || `${timezone.location} (${timezone.offset})`}</p>
          <p><strong>Current Time:</strong> {new Date().toLocaleString('en-US', { 
            timeZone: timezone.location || timezone.timezone,
            dateStyle: 'medium',
            timeStyle: 'medium'
          })}</p>
          <p><strong>UTC Offset:</strong> {timezone.offset || timezone.utc_offset}</p>
        </div>
      )}

      <form onSubmit={handleTimezoneChange} className="timezone-form">
        <div className="form-group">
          <label htmlFor="timezone">Select Timezone:</label>
          <select
            id="timezone"
            value={selectedTimezone}
            onChange={(e) => setSelectedTimezone(e.target.value)}
            disabled={isLoading}
          >
            <option value="">Select a timezone...</option>
            {commonTimezones.map(tz => (
              <option key={tz} value={tz}>
                {tz === 'System' ? `System (${getSystemTimezone()})` : tz}
              </option>
            ))}
          </select>
        </div>
        
        <button 
          type="submit" 
          disabled={isLoading || !selectedTimezone || selectedTimezone === (timezone?.location || timezone?.timezone)}
        >
          {isLoading ? 'Updating...' : 'Update Timezone'}
        </button>
      </form>
    </div>
  );
}