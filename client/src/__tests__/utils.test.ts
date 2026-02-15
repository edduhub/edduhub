import { cn, formatDistanceToNow } from '@/lib/utils';

describe('cn utility', () => {
  it('merges class names correctly', () => {
    const result = cn('foo', 'bar');
    expect(result).toBe('foo bar');
  });

  it('handles conditional class names', () => {
    const result = cn('foo', false && 'bar', 'baz');
    expect(result).toBe('foo baz');
  });

  it('handles undefined and null values', () => {
    const result = cn('foo', undefined, null, 'bar');
    expect(result).toBe('foo bar');
  });

  it('merges tailwind classes with same utility (last one wins)', () => {
    const result = cn('px-2 px-4');
    expect(result).toBe('px-4');
  });

  it('handles array inputs', () => {
    const result = cn(['foo', 'bar']);
    expect(result).toBe('foo bar');
  });

  it('handles object inputs for conditional classes', () => {
    const result = cn({ 'foo-bar': true, 'baz-qux': false });
    expect(result).toBe('foo-bar');
  });

  it('handles mixed inputs', () => {
    const result = cn('base', ['array'], { conditional: true }, false && 'nope');
    expect(result).toBe('base array conditional');
  });
});

describe('formatDistanceToNow', () => {
  beforeEach(() => {
    // Mock Date to return a fixed time
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2024-01-15T12:00:00Z'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('returns "just now" for dates less than a minute ago', () => {
    const date = new Date('2024-01-15T11:59:30Z'); // 30 seconds ago
    expect(formatDistanceToNow(date)).toBe('just now');
  });

  it('returns minutes ago for dates less than an hour ago', () => {
    const date = new Date('2024-01-15T11:30:00Z'); // 30 minutes ago
    expect(formatDistanceToNow(date)).toBe('30m ago');
  });

  it('returns hours ago for dates less than a day ago', () => {
    const date = new Date('2024-01-15T08:00:00Z'); // 4 hours ago
    expect(formatDistanceToNow(date)).toBe('4h ago');
  });

  it('returns days ago for dates less than a week ago', () => {
    const date = new Date('2024-01-10T12:00:00Z'); // 5 days ago
    expect(formatDistanceToNow(date)).toBe('5d ago');
  });

  it('handles string date inputs', () => {
    const date = '2024-01-14T12:00:00Z'; // 1 day ago
    expect(formatDistanceToNow(date)).toBe('1d ago');
  });

  it('handles ISO string date inputs', () => {
    const date = '2024-01-13T12:00:00Z'; // 2 days ago
    expect(formatDistanceToNow(date)).toBe('2d ago');
  });

  it('returns correct format for 1 minute ago', () => {
    const date = new Date('2024-01-15T11:59:00Z'); // 1 minute ago
    expect(formatDistanceToNow(date)).toBe('1m ago');
  });

  it('returns correct format for 1 hour ago', () => {
    const date = new Date('2024-01-15T11:00:00Z'); // 1 hour ago
    expect(formatDistanceToNow(date)).toBe('1h ago');
  });

  it('returns correct format for 1 day ago', () => {
    const date = new Date('2024-01-14T12:00:00Z'); // 1 day ago
    expect(formatDistanceToNow(date)).toBe('1d ago');
  });
});
