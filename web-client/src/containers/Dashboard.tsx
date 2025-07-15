import { useState, useEffect } from 'react';
import { usePostStore } from '../store';
import { PostForm } from '../components/PostForm';
import { PostList } from '../components/PostList';
import { SchedulerStatus } from '../components/SchedulerStatus';
import { TimezoneSettings } from '../components/TimezoneSettings';
import { AuthStatus } from '../components/AuthStatus';
import type { PostRequest, Post } from '../models';

export function Dashboard() {
  const {
    posts,
    selectedPost,
    isLoading,
    error,
    fetchPosts,
    createPost,
    updatePost,
    deletePost,
    deleteMultiplePosts,
    publishPost,
    setSelectedPost,
    clearError
  } = usePostStore();

  const [isEditing, setIsEditing] = useState(false);
  const [activeTab, setActiveTab] = useState<'posts' | 'settings'>('posts');

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  const handleCreatePost = async (postData: PostRequest) => {
    await createPost(postData);
    if (!error) {
      setIsEditing(false);
    }
  };

  const handleUpdatePost = async (postData: PostRequest) => {
    if (selectedPost) {
      await updatePost(selectedPost.id, postData);
      if (!error) {
        setIsEditing(false);
        setSelectedPost(null);
      }
    }
  };

  const handleEditPost = (post: Post) => {
    setSelectedPost(post);
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setSelectedPost(null);
  };

  const handleDeletePost = async (id: number) => {
    if (confirm('Are you sure you want to delete this post?')) {
      await deletePost(id);
    }
  };

  const handleDeleteMultiplePosts = async (ids: number[]) => {
    if (confirm(`Are you sure you want to delete ${ids.length} posts?`)) {
      await deleteMultiplePosts(ids);
    }
  };

  const handlePublishPost = async (id: number) => {
    if (confirm('Are you sure you want to publish this post now?')) {
      await publishPost(id);
    }
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="header-brand">
          <div className="logo">
            <img src="/postedin-logo.png" alt="PostedIn Sloth Logo" className="logo-image" />
          </div>
          <h1>PostedIn - LinkedIn Post Scheduler</h1>
        </div>
        <nav className="dashboard-nav">
          <button 
            onClick={() => setActiveTab('posts')}
            className={activeTab === 'posts' ? 'active' : ''}
          >
            <span className="nav-icon">📝</span>
            Posts
          </button>
          <button 
            onClick={() => setActiveTab('settings')}
            className={activeTab === 'settings' ? 'active' : ''}
          >
            <span className="nav-icon">⚙️</span>
            Settings
          </button>
        </nav>
      </header>

      {error && (
        <div className="error-banner">
          <p>{error}</p>
          <button onClick={clearError}>×</button>
        </div>
      )}

      <div className="dashboard-content">
        {activeTab === 'posts' ? (
          <div className="posts-section">
            <div className="sidebar">
              <SchedulerStatus />
              <AuthStatus />
            </div>
            
            <div className="main-content">
              <div className="content-header">
                <h2>Scheduled Posts</h2>
                <button 
                  onClick={() => setIsEditing(true)}
                  disabled={isLoading}
                  className="create-post-btn"
                >
                  Create New Post
                </button>
              </div>

              {isEditing && (
                <div className="post-form-container">
                  <div className="form-header">
                    <h3>{selectedPost ? 'Edit Post' : 'Create New Post'}</h3>
                    <button onClick={handleCancelEdit}>Cancel</button>
                  </div>
                  <PostForm
                    onSubmit={selectedPost ? handleUpdatePost : handleCreatePost}
                    initialData={selectedPost ? {
                      content: selectedPost.content,
                      scheduled_at: selectedPost.scheduled_at
                    } : undefined}
                    isLoading={isLoading}
                  />
                </div>
              )}

              <PostList
                posts={posts}
                onDelete={handleDeletePost}
                onDeleteMultiple={handleDeleteMultiplePosts}
                onPublish={handlePublishPost}
                onEdit={handleEditPost}
                isLoading={isLoading}
              />
            </div>
          </div>
        ) : (
          <div className="settings-section">
            <h2>Settings</h2>
            <div className="settings-grid">
              <TimezoneSettings />
              <AuthStatus />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}