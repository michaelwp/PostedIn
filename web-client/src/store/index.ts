import { create } from 'zustand';
import type { Post, PostRequest, SchedulerStatus, TimezoneInfo } from '../models';
import { apiClient } from '../services/api';

interface PostState {
  posts: Post[];
  duePosts: Post[];
  selectedPost: Post | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchPosts: () => Promise<void>;
  fetchDuePosts: () => Promise<void>;
  createPost: (post: PostRequest) => Promise<void>;
  updatePost: (id: number, post: Partial<PostRequest>) => Promise<void>;
  deletePost: (id: number) => Promise<void>;
  deleteMultiplePosts: (ids: number[]) => Promise<void>;
  publishPost: (id: number) => Promise<void>;
  publishDuePosts: () => Promise<void>;
  setSelectedPost: (post: Post | null) => void;
  clearError: () => void;
}

interface SchedulerState {
  status: SchedulerStatus | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchStatus: () => Promise<void>;
  startScheduler: () => Promise<void>;
  stopScheduler: () => Promise<void>;
  clearError: () => void;
}

interface TimezoneState {
  timezone: TimezoneInfo | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchTimezone: () => Promise<void>;
  setTimezone: (timezone: string) => Promise<void>;
  clearError: () => void;
}

interface AuthState {
  isAuthenticated: boolean;
  userId: string | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  checkAuthStatus: () => Promise<void>;
  startAuth: () => Promise<string>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const usePostStore = create<PostState>((set, get) => ({
  posts: [],
  duePosts: [],
  selectedPost: null,
  isLoading: false,
  error: null,
  
  fetchPosts: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getPosts();
      if (response.success && response.data) {
        set({ posts: response.data, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to fetch posts', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch posts', isLoading: false });
    }
  },
  
  fetchDuePosts: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getDuePosts();
      if (response.success && response.data) {
        set({ duePosts: response.data, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to fetch due posts', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch due posts', isLoading: false });
    }
  },
  
  createPost: async (post: PostRequest) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.createPost(post);
      if (response.success && response.data) {
        const { posts } = get();
        set({ posts: [...posts, response.data], isLoading: false });
      } else {
        set({ error: response.error || 'Failed to create post', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to create post', isLoading: false });
    }
  },
  
  updatePost: async (id: number, postUpdate: Partial<PostRequest>) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.updatePost(id, postUpdate);
      if (response.success && response.data) {
        const { posts } = get();
        const updatedPosts = posts.map(p => p.id === id ? response.data! : p);
        set({ posts: updatedPosts, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to update post', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to update post', isLoading: false });
    }
  },
  
  deletePost: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.deletePost(id);
      if (response.success) {
        const { posts } = get();
        const updatedPosts = posts.filter(p => p.id !== id);
        set({ posts: updatedPosts, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to delete post', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete post', isLoading: false });
    }
  },
  
  deleteMultiplePosts: async (ids: number[]) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.deleteMultiplePosts(ids);
      if (response.success) {
        const { posts } = get();
        const updatedPosts = posts.filter(p => !ids.includes(p.id));
        set({ posts: updatedPosts, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to delete posts', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete posts', isLoading: false });
    }
  },
  
  publishPost: async (id: number) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.publishPost(id);
      if (response.success) {
        const { posts } = get();
        const updatedPosts = posts.map(p => 
          p.id === id ? { ...p, status: 'published' } : p
        );
        set({ posts: updatedPosts, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to publish post', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to publish post', isLoading: false });
    }
  },
  
  publishDuePosts: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.publishDuePosts();
      if (response.success) {
        // Refresh posts after publishing
        await get().fetchPosts();
        set({ isLoading: false });
      } else {
        set({ error: response.error || 'Failed to publish due posts', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to publish due posts', isLoading: false });
    }
  },
  
  setSelectedPost: (post: Post | null) => {
    set({ selectedPost: post });
  },
  
  clearError: () => {
    set({ error: null });
  },
}));

export const useSchedulerStore = create<SchedulerState>((set) => ({
  status: null,
  isLoading: false,
  error: null,
  
  fetchStatus: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getSchedulerStatus();
      if (response.success && response.data) {
        set({ status: response.data, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to fetch scheduler status', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch scheduler status', isLoading: false });
    }
  },
  
  startScheduler: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.startScheduler();
      if (response.success) {
        set({ isLoading: false });
        // Refresh status after starting
        await useSchedulerStore.getState().fetchStatus();
      } else {
        set({ error: response.error || 'Failed to start scheduler', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to start scheduler', isLoading: false });
    }
  },
  
  stopScheduler: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.stopScheduler();
      if (response.success) {
        set({ isLoading: false });
        // Refresh status after stopping
        await useSchedulerStore.getState().fetchStatus();
      } else {
        set({ error: response.error || 'Failed to stop scheduler', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to stop scheduler', isLoading: false });
    }
  },
  
  clearError: () => {
    set({ error: null });
  },
}));

export const useTimezoneStore = create<TimezoneState>((set) => ({
  timezone: null,
  isLoading: false,
  error: null,
  
  fetchTimezone: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getTimezone();
      if (response.success && response.data) {
        set({ timezone: response.data, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to fetch timezone', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch timezone', isLoading: false });
    }
  },
  
  setTimezone: async (timezone: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.setTimezone(timezone);
      if (response.success && response.data) {
        set({ timezone: response.data, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to set timezone', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to set timezone', isLoading: false });
    }
  },
  
  clearError: () => {
    set({ error: null });
  },
}));

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  userId: null,
  isLoading: false,
  error: null,
  
  checkAuthStatus: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.getAuthStatus();
      if (response.success && response.data) {
        set({ 
          isAuthenticated: response.data.authenticated,
          userId: response.data.user_id || null,
          isLoading: false 
        });
      } else {
        set({ error: response.error || 'Failed to check auth status', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to check auth status', isLoading: false });
    }
  },
  
  startAuth: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.startAuth();
      if (response.success && response.auth_url) {
        set({ isLoading: false });
        return response.auth_url;
      } else {
        set({ error: response.error || 'Failed to start auth', isLoading: false });
        throw new Error(response.error || 'Failed to start auth');
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to start auth', isLoading: false });
      throw error;
    }
  },
  
  logout: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiClient.logout();
      if (response.success) {
        set({ isAuthenticated: false, userId: null, isLoading: false });
      } else {
        set({ error: response.error || 'Failed to logout', isLoading: false });
      }
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to logout', isLoading: false });
    }
  },
  
  clearError: () => {
    set({ error: null });
  },
}));