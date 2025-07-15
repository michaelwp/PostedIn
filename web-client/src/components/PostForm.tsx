import { useState } from 'react';
import type { PostRequest } from '../models';

interface PostFormProps {
  onSubmit: (post: PostRequest) => void;
  initialData?: Partial<PostRequest>;
  isLoading?: boolean;
}

const formatDateTimeLocal = (dateStr: string) => {
  if (!dateStr) return '';
  
  try {
    // Parse the date regardless of format (ISO, space-separated, etc.)
    const date = new Date(dateStr);
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid date string:', dateStr);
      return '';
    }
    
    // Convert to local datetime-local format (YYYY-MM-DDTHH:MM)
    // This removes timezone info and converts to local time
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  } catch (error) {
    console.error('Error formatting date:', error, dateStr);
    return '';
  }
};

const formatForAPI = (dateTimeLocal: string) => {
  if (!dateTimeLocal) return '';
  
  try {
    // dateTimeLocal format is YYYY-MM-DDTHH:MM:SS
    // Backend expects YYYY-MM-DD HH:MM format
    // So we need to replace 'T' with space and remove seconds
    
    if (dateTimeLocal.includes('T')) {
      // Split date and time parts
      const [datePart, timePart] = dateTimeLocal.split('T');
      const timeWithoutSeconds = timePart.split(':').slice(0, 2).join(':'); // Take only HH:MM
      return `${datePart} ${timeWithoutSeconds}`;
    }
    
    return dateTimeLocal; // Already in correct format
  } catch (error) {
    console.error('Error formatting for API:', error, dateTimeLocal);
    return '';
  }
};

export function PostForm({ onSubmit, initialData, isLoading = false }: PostFormProps) {
  const [content, setContent] = useState(initialData?.content || '');
  
  // Split datetime into separate date and time values
  const [dateValue, setDateValue] = useState(() => {
    if (initialData?.scheduled_at) {
      const date = new Date(initialData.scheduled_at);
      return date.toISOString().split('T')[0]; // YYYY-MM-DD format
    }
    return '';
  });
  
  const [timeValue, setTimeValue] = useState(() => {
    if (initialData?.scheduled_at) {
      const date = new Date(initialData.scheduled_at);
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      return `${hours}:${minutes}`; // HH:MM format
    }
    return '';
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim() && dateValue && timeValue) {
      // Combine date and time into API format
      const dateTimeString = `${dateValue}T${timeValue}:00`; // Add seconds
      const apiDateFormat = formatForAPI(dateTimeString);
      console.log('Sending to API:', apiDateFormat);
      onSubmit({ content: content.trim(), scheduled_at: apiDateFormat });
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setDateValue(e.target.value);
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTimeValue(e.target.value);
  };

  return (
    <form onSubmit={handleSubmit} className="post-form">
      <div className="form-group">
        <label htmlFor="content">Post Content</label>
        <textarea
          id="content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Enter your LinkedIn post content..."
          required
          rows={4}
          disabled={isLoading}
        />
      </div>
      
      <div className="form-group">
        <label>Scheduled Date & Time</label>
        <div className="datetime-inputs">
          <div className="date-input">
            <label htmlFor="scheduledDate">Date</label>
            <input
              type="date"
              id="scheduledDate"
              value={dateValue}
              onChange={handleDateChange}
              required
              disabled={isLoading}
            />
          </div>
          <div className="time-input">
            <label htmlFor="scheduledTime">Time</label>
            <input
              type="time"
              id="scheduledTime"
              value={timeValue}
              onChange={handleTimeChange}
              required
              disabled={isLoading}
            />
          </div>
        </div>
        <small className="form-help">
          Format: YYYY-MM-DD HH:MM (e.g., 2025-07-16 14:30)
        </small>
      </div>
      
      <button type="submit" disabled={isLoading || !content.trim() || !dateValue || !timeValue}>
        {isLoading ? 'Saving...' : 'Schedule Post'}
      </button>
    </form>
  );
}