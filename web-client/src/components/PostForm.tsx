import { useState } from 'react';
import type { PostRequest, PostImage } from '../models';

interface PostFormProps {
  onSubmit: (post: PostRequest) => void;
  initialData?: Partial<PostRequest>;
  isLoading?: boolean;
  isEdit?: boolean; // New prop to distinguish edit mode
}


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

export function PostForm({ onSubmit, initialData, isLoading = false, isEdit = false }: PostFormProps) {
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

  const [images, setImages] = useState<PostImage[]>(initialData?.images || []);
  const [newAltText, setNewAltText] = useState('');
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim() && dateValue && timeValue) {
      const dateTimeString = `${dateValue}T${timeValue}:00`;
      const apiDateFormat = formatForAPI(dateTimeString);
      onSubmit({
        content: content.trim(),
        scheduled_at: apiDateFormat,
        images: images.length > 0 ? images : undefined
      });
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setDateValue(e.target.value);
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTimeValue(e.target.value);
  };

  // Remove handleAddImage since it is no longer used
  const handleRemoveImage = (id: string) => {
    setImages(images.filter(img => img.id !== id));
  };
  const handleAltTextChange = (id: string, altText: string) => {
    setImages(images.map(img => img.id === id ? { ...img, altText } : img));
  };

  // Upload image to backend and add to images array
  const handleUploadImage = async () => {
    if (!selectedFile) return;
    setUploading(true);
    setUploadError(null);
    try {
      const formData = new FormData();
      formData.append('image', selectedFile);
      formData.append('altText', newAltText);
      const res = await fetch('/api/v1/posts/upload-image', {
        method: 'POST',
        body: formData,
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error || 'Upload failed');
      setImages([...images, { id: data.urn, altText: data.altText }]);
      setSelectedFile(null);
      setNewAltText('');
    } catch (err: any) {
      setUploadError(err.message || 'Upload failed');
    } finally {
      setUploading(false);
    }
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
      
      <div className="form-group">
        <label htmlFor="images">Images & Alt Text</label>
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.5rem', flexWrap: 'wrap' }}>
          <input
            type="file"
            accept="image/*"
            onChange={e => setSelectedFile(e.target.files?.[0] || null)}
            disabled={isLoading || uploading}
          />
          <input
            type="text"
            value={newAltText}
            onChange={e => setNewAltText(e.target.value)}
            placeholder="Alt text (optional)"
            disabled={isLoading || uploading}
          />
          <button type="button" onClick={handleUploadImage} disabled={isLoading || uploading || !selectedFile}>
            {uploading ? 'Uploading...' : 'Upload to LinkedIn'}
          </button>
        </div>
        {uploadError && <div style={{ color: 'red', marginBottom: 4 }}>{uploadError}</div>}
        {images.length > 0 && (
          <ul style={{ paddingLeft: 20 }}>
            {images.map((img, idx) => (
              <li key={img.id + idx} style={{ marginBottom: 4 }}>
                <span style={{ wordBreak: 'break-all' }}>{img.id}</span>
                <input
                  type="text"
                  value={img.altText}
                  onChange={e => handleAltTextChange(img.id, e.target.value)}
                  placeholder="Alt text"
                  style={{ marginLeft: 8, width: 180 }}
                  disabled={isLoading}
                />
                <button type="button" onClick={() => handleRemoveImage(img.id)} style={{ marginLeft: 8 }} disabled={isLoading}>
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
        <small className="form-help">Upload images to LinkedIn. Alt text is recommended for accessibility.</small>
      </div>
      
      <button type="submit" disabled={isLoading || !content.trim() || !dateValue || !timeValue}>
        {isLoading ? 'Saving...' : isEdit ? 'Update Post' : 'Schedule Post'}
      </button>
    </form>
  );
}