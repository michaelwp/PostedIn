export interface Post {
  id: number;
  content: string;
  status: string;
  scheduled_at: string;
  created_at: string;
  cron_entry_id?: number;
}

export interface PostRequest {
  content: string;
  scheduled_at: string;
}

export interface DeletePostsRequest {
  ids: number[];
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface CreatePostResponse extends ApiResponse<Post> {}
export interface GetPostsResponse extends ApiResponse<Post[]> {}
export interface GetPostResponse extends ApiResponse<Post> {}
export interface UpdatePostResponse extends ApiResponse<Post> {}

export interface DeletePostResponse extends ApiResponse<never> {
  deleted_id?: number;
  message?: string;
}

export interface DeleteMultiplePostsResponse extends ApiResponse<never> {
  deleted_ids?: number[];
  count?: number;
  message?: string;
}

export interface PublishPostResponse extends ApiResponse<never> {
  published_id?: number;
  message?: string;
}

export interface PublishDuePostsResponse extends ApiResponse<never> {
  published?: number[];
  failed?: number[];
  message?: string;
}

export interface SchedulerStatus {
  running: boolean;
  posts_count: number;
  next_execution?: string;
}

export interface TimezoneInfo {
  location: string;
  offset: string;
  info: string;
  // Legacy fields for backward compatibility
  timezone?: string;
  current_time?: string;
  utc_offset?: string;
}