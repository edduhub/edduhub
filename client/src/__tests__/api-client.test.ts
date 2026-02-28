import { NetworkError } from '@/lib/api-client';

describe('NetworkError', () => {
  it('creates an error with the correct name', () => {
    const err = new NetworkError('Connection failed');
    expect(err.name).toBe('NetworkError');
  });

  it('creates an error with the correct message', () => {
    const err = new NetworkError('No internet');
    expect(err.message).toBe('No internet');
  });

  it('is an instance of Error', () => {
    const err = new NetworkError('timeout');
    expect(err).toBeInstanceOf(Error);
  });

  it('is an instance of NetworkError', () => {
    const err = new NetworkError('timeout');
    expect(err).toBeInstanceOf(NetworkError);
  });

  it('has a stack trace', () => {
    const err = new NetworkError('test');
    expect(err.stack).toBeDefined();
  });
});
