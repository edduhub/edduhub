import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,

  // Set tracesSampleRate to 1.0 to capture 100%
  // of transactions for performance monitoring.
  tracesSampleRate: 1.0,

  // Environment
  environment: process.env.NODE_ENV,

  // beforeSend filter for server-side
  beforeSend(event) {
    // Don't send events in development
    if (process.env.NODE_ENV === 'development') {
      return null;
    }

    // Filter out sensitive data
    if (event.request?.data) {
      const data = { ...event.request.data } as Record<string, unknown>;
      const sensitiveFields = ['password', 'token', 'apiKey', 'secret', 'authorization'];
      
      sensitiveFields.forEach(field => {
        delete data[field];
      });
      
      event.request.data = data;
    }

    return event;
  },
});
