import DOMPurify from 'dompurify';

/**
 * Sanitize HTML input to prevent XSS attacks
 * Uses DOMPurify for comprehensive sanitization
 */
export function sanitizeHtml(input: string): string {
  return DOMPurify.sanitize(input);
}

/**
 * Sanitize text input by removing HTML tags
 * This is for plain text inputs where any HTML should be stripped
 */
export function sanitizeText(input: string): string {
  return DOMPurify.sanitize(input, { ALLOWED_TAGS: [] });
}

/**
 * Sanitize user input for form fields
 * Provides basic sanitization that allows safe characters
 */
export function sanitizeInput(input: string): string {
  return DOMPurify.sanitize(input, {
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: [],
  }).trim();
}

/**
 * Sanitize rich text content (e.g., from a WYSIWYG editor)
 * Allows basic HTML formatting tags
 */
export function sanitizeRichText(input: string): string {
  return DOMPurify.sanitize(input, {
    ALLOWED_TAGS: [
      'p', 'br', 'strong', 'em', 'u', 'ul', 'ol', 'li',
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'a', 'blockquote', 'code', 'pre'
    ],
    ALLOWED_ATTR: ['href', 'title', 'class'],
    ALLOW_DATA_ATTR: false,
  });
}

/**
 * Sanitize URL to prevent javascript: and other dangerous protocols
 */
export function sanitizeUrl(url: string): string {
  const sanitized = DOMPurify.sanitize(url, { ALLOWED_TAGS: [] });
  
  const parsedUrl = sanitized.toLowerCase();
  const dangerousProtocols = ['javascript:', 'data:', 'vbscript:', 'file:'];
  
  for (const protocol of dangerousProtocols) {
    if (parsedUrl.startsWith(protocol)) {
      return '#';
    }
  }
  
  return sanitized;
}

/**
 * Configure DOMPurify for the application
 * This can be called during app initialization
 */
export function configureDOMPurify() {
  DOMPurify.addHook('uponSanitizeElement', (node, data) => {
    if (data.tagName && data.tagName.startsWith('on')) {
      if (node.parentNode) {
        node.parentNode.removeChild(node);
      }
    }
  });

  DOMPurify.addHook('uponSanitizeAttribute', (node, data) => {
    if (data.attrName && data.attrName.startsWith('on')) {
      node.removeAttribute(data.attrName);
    }
  });
}
