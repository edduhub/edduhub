// Custom error classes for the application
// These provide better type safety and error handling than generic Error class

/**
 * Base application error class
 * All custom errors should extend this class
 */
export class AppError extends Error {
  constructor(
    message: string,
    public code?: string,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'AppError';
    
    // Fix prototype chain for instanceof checks
    Object.setPrototypeOf(this, AppError.prototype);
  }

  /**
   * Serialize error to JSON for logging or API responses
   */
  toJSON() {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      details: this.details,
      stack: this.stack,
    };
  }
}

/**
 * API error with HTTP status code
 * Used when API requests fail
 */
export class APIError extends AppError {
  constructor(
    public status: number,
    message: string,
    code?: string,
    details?: Record<string, unknown>,
    public validationErrors?: ValidationError[]
  ) {
    super(message, code, details);
    this.name = 'APIError';
    Object.setPrototypeOf(this, APIError.prototype);
  }

  /**
   * Check if error is a client error (4xx)
   */
  isClientError(): boolean {
    return this.status >= 400 && this.status < 500;
  }

  /**
   * Check if error is a server error (5xx)
   */
  isServerError(): boolean {
    return this.status >= 500 && this.status < 600;
  }

  /**
   * Check if error is an authentication error (401)
   */
  isAuthError(): boolean {
    return this.status === 401;
  }

  /**
   * Check if error is a forbidden error (403)
   */
  isForbiddenError(): boolean {
    return this.status === 403;
  }

  /**
   * Check if error is a not found error (404)
   */
  isNotFoundError(): boolean {
    return this.status === 404;
  }
}

/**
 * Validation error for a specific field
 */
export interface ValidationError {
  field: string;
  message: string;
  value?: unknown;
}

/**
 * Network error for connection issues
 */
export class NetworkError extends AppError {
  constructor(
    message: string = 'Network connection failed',
    public originalError?: Error
  ) {
    super(message, 'NETWORK_ERROR');
    this.name = 'NetworkError';
    Object.setPrototypeOf(this, NetworkError.prototype);
  }
}

/**
 * Authentication error for login/session issues
 */
export class AuthenticationError extends AppError {
  constructor(
    message: string = 'Authentication failed',
    public reason?: 'invalid_credentials' | 'session_expired' | 'unauthorized'
  ) {
    super(message, 'AUTH_ERROR');
    this.name = 'AuthenticationError';
    Object.setPrototypeOf(this, AuthenticationError.prototype);
  }
}

/**
 * Not found error for missing resources
 */
export class NotFoundError extends AppError {
  constructor(
    resource: string,
    public resourceId?: string | number
  ) {
    super(
      resourceId 
        ? `${resource} with id "${resourceId}" not found`
        : `${resource} not found`,
      'NOT_FOUND'
    );
    this.name = 'NotFoundError';
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * Timeout error for requests that take too long
 */
export class TimeoutError extends AppError {
  constructor(
    message: string = 'Request timeout',
    public timeoutMs?: number
  ) {
    super(message, 'TIMEOUT_ERROR');
    this.name = 'TimeoutError';
    Object.setPrototypeOf(this, TimeoutError.prototype);
  }
}

/**
 * Type guard to check if error is an APIError
 */
export function isAPIError(error: unknown): error is APIError {
  return error instanceof APIError || 
    (typeof error === 'object' && 
     error !== null && 
     'name' in error && 
     (error as Error).name === 'APIError');
}

/**
 * Type guard to check if error is a NetworkError
 */
export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof NetworkError ||
    (typeof error === 'object' &&
     error !== null &&
     'name' in error &&
     (error as Error).name === 'NetworkError');
}

/**
 * Type guard to check if error is a ValidationError array
 */
export function isValidationErrors(error: unknown): error is ValidationError[] {
  return (
    Array.isArray(error) &&
    error.length > 0 &&
    typeof error[0] === 'object' &&
    error[0] !== null &&
    'field' in error[0] &&
    'message' in error[0]
  );
}

/**
 * Type guard to check if error is an AuthenticationError
 */
export function isAuthenticationError(error: unknown): error is AuthenticationError {
  return error instanceof AuthenticationError ||
    (typeof error === 'object' &&
     error !== null &&
     'name' in error &&
     (error as Error).name === 'AuthenticationError');
}

/**
 * Type guard to check if error is a NotFoundError
 */
export function isNotFoundError(error: unknown): error is NotFoundError {
  return error instanceof NotFoundError ||
    (typeof error === 'object' &&
     error !== null &&
     'name' in error &&
     (error as Error).name === 'NotFoundError');
}

/**
 * Extract error message from unknown error
 * Safe way to get message from any error type
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'string') {
    return error;
  }
  if (typeof error === 'object' && error !== null && 'message' in error) {
    return String((error as { message: unknown }).message);
  }
  return 'An unknown error occurred';
}

/**
 * Extract HTTP status code from error if available
 */
export function getErrorStatus(error: unknown): number | undefined {
  if (isAPIError(error)) {
    return error.status;
  }
  if (typeof error === 'object' && error !== null && 'status' in error) {
    const status = (error as { status: unknown }).status;
    if (typeof status === 'number') {
      return status;
    }
  }
  return undefined;
}
