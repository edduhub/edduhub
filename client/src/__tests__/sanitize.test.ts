const mockSanitize = jest.fn((input: string) => input);
const mockAddHook = jest.fn();
const mockRemoveHook = jest.fn();

jest.mock('dompurify', () => ({
  __esModule: true,
  default: {
    sanitize: mockSanitize,
    addHook: mockAddHook,
    removeHook: mockRemoveHook,
  },
}));

import {
  sanitizeHtml,
  sanitizeText,
  sanitizeInput,
  sanitizeRichText,
  sanitizeUrl,
  configureDOMPurify,
} from '@/lib/sanitize';

describe('sanitizeHtml', () => {
  beforeEach(() => {
    mockSanitize.mockClear();
  });

  it('sanitizes HTML input', () => {
    sanitizeHtml('<script>alert("xss")</script>');
    expect(mockSanitize).toHaveBeenCalledWith('<script>alert("xss")</script>');
  });
});

describe('sanitizeText', () => {
  beforeEach(() => {
    mockSanitize.mockClear();
  });

  it('removes all HTML tags', () => {
    sanitizeText('<p>Hello</p>');
    expect(mockSanitize).toHaveBeenCalledWith('<p>Hello</p>', { ALLOWED_TAGS: [] });
  });
});

describe('sanitizeInput', () => {
  beforeEach(() => {
    mockSanitize.mockClear().mockImplementation((input: string) => input);
  });

  it('sanitizes with empty tags and attributes', () => {
    sanitizeInput('  test  ');
    expect(mockSanitize).toHaveBeenCalledWith('  test  ', { ALLOWED_TAGS: [], ALLOWED_ATTR: [] });
  });
});

describe('sanitizeRichText', () => {
  beforeEach(() => {
    mockSanitize.mockClear();
  });

  it('allows basic formatting tags', () => {
    sanitizeRichText('<p>Hello <strong>world</strong></p>');
    expect(mockSanitize).toHaveBeenCalledWith('<p>Hello <strong>world</strong></p>', {
      ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 'ul', 'ol', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'a', 'blockquote', 'code', 'pre'],
      ALLOWED_ATTR: ['href', 'title', 'class'],
      ALLOW_DATA_ATTR: false,
    });
  });
});

describe('sanitizeUrl', () => {
  beforeEach(() => {
    mockSanitize.mockClear();
  });

  it('sanitizes URL and blocks dangerous protocols', () => {
    const result = sanitizeUrl('javascript:alert(1)');
    expect(result).toBe('#');
  });

  it('sanitizes data: protocol', () => {
    const result = sanitizeUrl('data:text/html,<script>alert(1)</script>');
    expect(result).toBe('#');
  });

  it('allows safe URLs', () => {
    const result = sanitizeUrl('https://example.com');
    expect(result).toBe('https://example.com');
  });

  it('allows http URLs', () => {
    const result = sanitizeUrl('http://example.com');
    expect(result).toBe('http://example.com');
  });
});

describe('configureDOMPurify', () => {
  beforeEach(() => {
    mockAddHook.mockClear();
  });

  it('adds hooks to DOMPurify', () => {
    configureDOMPurify();
    expect(mockAddHook).toHaveBeenCalledTimes(2);
  });
});
