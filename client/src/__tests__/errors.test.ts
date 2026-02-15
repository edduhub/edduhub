import {
  AppError,
  APIError,
  NetworkError,
  AuthenticationError,
  NotFoundError,
  TimeoutError,
  isAPIError,
  isNetworkError,
  isValidationErrors,
  isAuthenticationError,
  isNotFoundError,
  getErrorMessage,
  getErrorStatus,
  ValidationError,
} from '@/lib/errors';

describe('AppError', () => {
  it('creates error with message', () => {
    const error = new AppError('Test error');
    expect(error.message).toBe('Test error');
    expect(error.name).toBe('AppError');
  });

  it('creates error with code', () => {
    const error = new AppError('Test error', 'ERROR_CODE');
    expect(error.code).toBe('ERROR_CODE');
  });

  it('creates error with details', () => {
    const error = new AppError('Test error', 'ERROR_CODE', { key: 'value' });
    expect(error.details).toEqual({ key: 'value' });
  });

  it('serializes to JSON', () => {
    const error = new AppError('Test error', 'ERROR_CODE', { key: 'value' });
    const json = error.toJSON();
    expect(json.message).toBe('Test error');
    expect(json.code).toBe('ERROR_CODE');
    expect(json.details).toEqual({ key: 'value' });
  });
});

describe('APIError', () => {
  it('creates error with status and message', () => {
    const error = new APIError(404, 'Not found');
    expect(error.status).toBe(404);
    expect(error.message).toBe('Not found');
    expect(error.name).toBe('APIError');
  });

  it('isClientError returns true for 4xx', () => {
    expect(new APIError(400, 'Bad request').isClientError()).toBe(true);
    expect(new APIError(401, 'Unauthorized').isClientError()).toBe(true);
    expect(new APIError(403, 'Forbidden').isClientError()).toBe(true);
    expect(new APIError(404, 'Not found').isClientError()).toBe(true);
    expect(new APIError(422, 'Unprocessable').isClientError()).toBe(true);
  });

  it('isClientError returns false for 5xx', () => {
    expect(new APIError(500, 'Server error').isClientError()).toBe(false);
    expect(new APIError(503, 'Service unavailable').isClientError()).toBe(false);
  });

  it('isServerError returns true for 5xx', () => {
    expect(new APIError(500, 'Server error').isServerError()).toBe(true);
    expect(new APIError(503, 'Service unavailable').isServerError()).toBe(true);
  });

  it('isServerError returns false for 4xx', () => {
    expect(new APIError(400, 'Bad request').isServerError()).toBe(false);
  });

  it('isAuthError returns true for 401', () => {
    expect(new APIError(401, 'Unauthorized').isAuthError()).toBe(true);
  });

  it('isAuthError returns false for other status', () => {
    expect(new APIError(400, 'Bad request').isAuthError()).toBe(false);
    expect(new APIError(403, 'Forbidden').isAuthError()).toBe(false);
  });

  it('isForbiddenError returns true for 403', () => {
    expect(new APIError(403, 'Forbidden').isForbiddenError()).toBe(true);
  });

  it('isNotFoundError returns true for 404', () => {
    expect(new APIError(404, 'Not found').isNotFoundError()).toBe(true);
  });

  it('handles validation errors', () => {
    const validationErrors: ValidationError[] = [
      { field: 'email', message: 'Invalid email' },
    ];
    const error = new APIError(400, 'Validation failed', 'VALIDATION_ERROR', undefined, validationErrors);
    expect(error.validationErrors).toEqual(validationErrors);
  });
});

describe('NetworkError', () => {
  it('creates error with default message', () => {
    const error = new NetworkError();
    expect(error.message).toBe('Network connection failed');
    expect(error.code).toBe('NETWORK_ERROR');
  });

  it('creates error with custom message', () => {
    const error = new NetworkError('Connection refused');
    expect(error.message).toBe('Connection refused');
  });

  it('stores original error', () => {
    const original = new Error('Original error');
    const error = new NetworkError('Failed', original);
    expect(error.originalError).toBe(original);
  });
});

describe('AuthenticationError', () => {
  it('creates error with default message', () => {
    const error = new AuthenticationError();
    expect(error.message).toBe('Authentication failed');
    expect(error.code).toBe('AUTH_ERROR');
  });

  it('creates error with reason', () => {
    const error = new AuthenticationError('Invalid credentials', 'invalid_credentials');
    expect(error.reason).toBe('invalid_credentials');
  });
});

describe('NotFoundError', () => {
  it('creates error with resource name', () => {
    const error = new NotFoundError('User');
    expect(error.message).toBe('User not found');
  });

  it('creates error with resource id', () => {
    const error = new NotFoundError('User', 123);
    expect(error.message).toBe('User with id "123" not found');
    expect(error.resourceId).toBe(123);
  });
});

describe('TimeoutError', () => {
  it('creates error with default message', () => {
    const error = new TimeoutError();
    expect(error.message).toBe('Request timeout');
    expect(error.code).toBe('TIMEOUT_ERROR');
  });

  it('creates error with timeout value', () => {
    const error = new TimeoutError('Timeout after 5s', 5000);
    expect(error.timeoutMs).toBe(5000);
  });
});

describe('Type guards', () => {
  describe('isAPIError', () => {
    it('returns true for APIError instance', () => {
      const error = new APIError(400, 'Bad request');
      expect(isAPIError(error)).toBe(true);
    });

    it('returns false for other errors', () => {
      expect(isAPIError(new Error('Regular error'))).toBe(false);
      expect(isAPIError('string error')).toBe(false);
      expect(isAPIError(null)).toBe(false);
    });
  });

  describe('isNetworkError', () => {
    it('returns true for NetworkError instance', () => {
      const error = new NetworkError('Network failed');
      expect(isNetworkError(error)).toBe(true);
    });
  });

  describe('isAuthenticationError', () => {
    it('returns true for AuthenticationError instance', () => {
      const error = new AuthenticationError('Auth failed');
      expect(isAuthenticationError(error)).toBe(true);
    });
  });

  describe('isValidationErrors', () => {
    it('returns true for validation error array', () => {
      const errors: ValidationError[] = [{ field: 'email', message: 'Invalid' }];
      expect(isValidationErrors(errors)).toBe(true);
    });

    it('returns false for non-validation arrays', () => {
      expect(isValidationErrors([])).toBe(false);
      expect(isValidationErrors(['string'])).toBe(false);
      expect(isValidationErrors(null)).toBe(false);
    });
  });

  describe('isNotFoundError type guard', () => {
    it('returns true for NotFoundError instance', () => {
      const error = new NotFoundError('User', 123);
      expect(isNotFoundError(error)).toBe(true);
    });

    it('returns false for other errors', () => {
      expect(isNotFoundError(new Error('Regular error'))).toBe(false);
    });
  });
});

describe('getErrorMessage', () => {
  it('extracts message from Error', () => {
    expect(getErrorMessage(new Error('test'))).toBe('test');
  });

  it('returns string as is', () => {
    expect(getErrorMessage('string error')).toBe('string error');
  });

  it('extracts message from object', () => {
    expect(getErrorMessage({ message: 'object error' })).toBe('object error');
  });

  it('returns default for unknown', () => {
    expect(getErrorMessage(123)).toBe('An unknown error occurred');
  });
});

describe('getErrorStatus', () => {
  it('extracts status from APIError', () => {
    expect(getErrorStatus(new APIError(404, 'Not found'))).toBe(404);
  });

  it('extracts status from object with status', () => {
    expect(getErrorStatus({ status: 500 })).toBe(500);
  });

  it('returns undefined for errors without status', () => {
    expect(getErrorStatus(new Error('test'))).toBeUndefined();
  });
});
