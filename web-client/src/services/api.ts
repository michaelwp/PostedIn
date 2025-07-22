import type {
  PostRequest,
  Post,
  DeletePostResponse,
  DeleteMultiplePostsResponse,
  PublishPostResponse,
  PublishDuePostsResponse,
  SchedulerStatus,
  TimezoneInfo,
  ApiResponse
} from '../models';

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080/api/v1') {
    this.baseURL = baseURL;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Post endpoints
  async getPosts(): Promise<ApiResponse<Post[]>> {
    return this.request<ApiResponse<Post[]>>('/posts');
  }

  async getPost(id: number): Promise<ApiResponse<Post>> {
    return this.request<ApiResponse<Post>>(`/posts/${id}`);
  }

  // Supports images field in PostRequest
  async createPost(post: PostRequest): Promise<ApiResponse<Post>> {
    return this.request<ApiResponse<Post>>('/posts', {
      method: 'POST',
      body: JSON.stringify(post),
    });
  }

  // Supports images field in PostRequest
  async updatePost(id: number, post: Partial<PostRequest>): Promise<ApiResponse<Post>> {
    return this.request<ApiResponse<Post>>(`/posts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(post),
    });
  }

  async deletePost(id: number): Promise<DeletePostResponse> {
    return this.request<DeletePostResponse>(`/posts/${id}`, {
      method: 'DELETE',
    });
  }

  async deleteMultiplePosts(ids: number[]): Promise<DeleteMultiplePostsResponse> {
    return this.request<DeleteMultiplePostsResponse>('/posts', {
      method: 'DELETE',
      body: JSON.stringify({ ids }),
    });
  }

  async getDuePosts(): Promise<ApiResponse<Post[]>> {
    return this.request<ApiResponse<Post[]>>('/posts/due');
  }

  async publishPost(id: number): Promise<PublishPostResponse> {
    return this.request<PublishPostResponse>(`/posts/${id}/publish`, {
      method: 'POST',
    });
  }

  async publishDuePosts(): Promise<PublishDuePostsResponse> {
    return this.request<PublishDuePostsResponse>('/posts/publish-due', {
      method: 'POST',
    });
  }

  // Scheduler endpoints
  async getSchedulerStatus(): Promise<ApiResponse<SchedulerStatus>> {
    return this.request<ApiResponse<SchedulerStatus>>('/scheduler/status');
  }

  async startScheduler(): Promise<ApiResponse<never>> {
    return this.request<ApiResponse<never>>('/scheduler/start', {
      method: 'POST',
    });
  }

  async stopScheduler(): Promise<ApiResponse<never>> {
    return this.request<ApiResponse<never>>('/scheduler/stop', {
      method: 'POST',
    });
  }

  // Timezone endpoints
  async getTimezone(): Promise<ApiResponse<TimezoneInfo>> {
    return this.request<ApiResponse<TimezoneInfo>>('/timezone');
  }

  async setTimezone(timezone: string): Promise<ApiResponse<TimezoneInfo>> {
    return this.request<ApiResponse<TimezoneInfo>>('/timezone', {
      method: 'POST',
      body: JSON.stringify({ location: timezone }),
    });
  }

  // Auth endpoints
  async getAuthStatus(): Promise<ApiResponse<{ authenticated: boolean; user_id?: string }>> {
    return this.request<ApiResponse<{ authenticated: boolean; user_id?: string }>>('/auth/status');
  }

  async startAuth(): Promise<{ success: boolean; auth_url: string; error?: string }> {
    return this.request<{ success: boolean; auth_url: string; error?: string }>('/auth/linkedin');
  }

  async logout(): Promise<ApiResponse<never>> {
    return this.request<ApiResponse<never>>('/auth/logout', {
      method: 'POST',
    });
  }
}

export const apiClient = new ApiClient();
export default apiClient;