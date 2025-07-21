import { useState, useEffect } from 'react';
import type { Post, PostImage } from '../models';

interface PostListProps {
  posts: Post[];
  onDelete: (id: number) => void;
  onDeleteMultiple: (ids: number[]) => void;
  onPublish: (id: number) => void;
  onEdit: (post: Post) => void;
  isLoading?: boolean;
}

function LinkedInImage({ urn, alt }: { urn: string, alt: string }) {
  const [url, setUrl] = useState<string | null>(null);
  useEffect(() => {
    if (!urn.startsWith('urn:li:image:')) {
      setUrl(urn);
      return;
    }
    fetch(`/api/v1/posts/image-url?urn=${encodeURIComponent(urn)}`)
      .then(res => res.json())
      .then(data => {
        if (data.success) setUrl(data.url);
      });
  }, [urn]);
  if (!url) return <span>Loading image...</span>;
  return <img src={url} alt={alt} style={{ maxWidth: 80, maxHeight: 80, borderRadius: 4, border: '1px solid #eee', objectFit: 'cover' }} />;
}

export function PostList({ posts, onDelete, onDeleteMultiple, onPublish, onEdit, isLoading = false }: PostListProps) {
  const [selectedIds, setSelectedIds] = useState<number[]>([]);

  const handleSelectAll = (checked: boolean) => {
    setSelectedIds(checked ? posts.map(p => p.id) : []);
  };

  const handleSelectPost = (id: number, checked: boolean) => {
    setSelectedIds(prev => 
      checked 
        ? [...prev, id]
        : prev.filter(selectedId => selectedId !== id)
    );
  };

  const handleDeleteSelected = () => {
    if (selectedIds.length > 0) {
      onDeleteMultiple(selectedIds);
      setSelectedIds([]);
    }
  };

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'published': return 'green';
      case 'scheduled': return 'blue';
      case 'failed': return 'red';
      default: return 'gray';
    }
  };

  const isDue = (scheduledAt: string) => {
    return new Date(scheduledAt) <= new Date();
  };

  if (posts.length === 0) {
    return (
      <div className="empty-state">
        <p>No posts found. Create your first scheduled post!</p>
      </div>
    );
  }

  return (
    <div className="post-list">
      <div className="list-header">
        <div className="select-all">
          <input
            type="checkbox"
            checked={selectedIds.length === posts.length}
            onChange={(e) => handleSelectAll(e.target.checked)}
            disabled={isLoading}
          />
          <label>Select All</label>
        </div>
        
        {selectedIds.length > 0 && (
          <button 
            onClick={handleDeleteSelected}
            disabled={isLoading}
            className="delete-selected"
          >
            Delete Selected ({selectedIds.length})
          </button>
        )}
      </div>

      <div className="posts">
        {posts.map((post) => (
          <div key={post.id} className="post-item">
            <div className="post-select">
              <input
                type="checkbox"
                checked={selectedIds.includes(post.id)}
                onChange={(e) => handleSelectPost(post.id, e.target.checked)}
                disabled={isLoading}
              />
            </div>
            
            <div className="post-content">
              <div className="post-header">
                <span className="post-id">#{post.id}</span>
                <span 
                  className="post-status" 
                  style={{ color: getStatusColor(post.status) }}
                >
                  {post.status}
                </span>
                {isDue(post.scheduled_at) && post.status === 'scheduled' && (
                  <span className="due-badge">DUE</span>
                )}
              </div>
              
              <div className="post-text">
                {post.content}
              </div>
              {post.images && post.images.length > 0 && (
                <div className="post-images" style={{ marginTop: 8, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  {post.images.map((img: PostImage, idx: number) => (
                    <LinkedInImage key={img.id + idx} urn={img.id} alt={img.altText || `Post image ${idx + 1}`} />
                  ))}
                </div>
              )}
              
              <div className="post-meta">
                <span>Scheduled: {formatDateTime(post.scheduled_at)}</span>
                <span>Created: {formatDateTime(post.created_at)}</span>
              </div>
            </div>
            
            <div className="post-actions">
              <button 
                onClick={() => onEdit(post)}
                disabled={isLoading || post.status === 'published'}
                className="edit-btn"
              >
                Edit
              </button>
              <button 
                onClick={() => onPublish(post.id)}
                disabled={isLoading || post.status === 'published'}
                className="publish-btn"
              >
                Publish Now
              </button>
              <button 
                onClick={() => onDelete(post.id)}
                disabled={isLoading}
                className="delete-btn"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}