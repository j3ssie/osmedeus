// Generic API response types

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
  };
}

export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, string>;
}

export interface ApiResponse<T> {
  data: T;
  error?: ApiError;
}
