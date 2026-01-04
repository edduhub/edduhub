import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  
  // Set tracesSampleRate to 1.0 to capture 100%
  // of transactions for performance monitoring.
  tracesSampleRate: 1.0,

  // Set `tracePropagationTargets` to control for which URLs distributed tracing should be enabled
  tracePropagationTargets: ["localhost", /^https:\/\/yourdomain\.com/],

  // Capture Replay on 10% of all sessions,
  // plus for 100% of sessions with an error
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,

  // Environment
  environment: process.env.NODE_ENV,

  // beforeSend filter to filter out sensitive data
  beforeSend(event, hint) {
    // Don't send events in development
    if (process.env.NODE_ENV === 'development') {
      return null;
    }

    // Filter out sensitive data from request bodies
    if (event.request?.data) {
      const data = { ...event.request.data };
      delete data.password;
      delete data.passwordConfirm;
      delete data.currentPassword;
      delete data.newPassword;
      event.request.data = data;
    }

    return event;
  },

  // Filter out sensitive query params
  beforeBreadcrumb(breadcrumb) {
    if (breadcrumb.category === 'http') {
      const url = new URL(breadcrumb.data?.url || '');
      const sensitiveParams = ['password', 'token', 'apiKey', 'secret'];
      
      sensitiveParams.forEach(param => {
        url.searchParams.delete(param);
      });
      
      breadcrumb.data = {
        ...breadcrumb.data,
        url: url.toString(),
      };
    }
    return breadcrumb;
  },

  // Enable debugging in development
  debug: process.env.NODE_ENV === 'development',

  // Integrations
  integrations: [
    Sentry.replayIntegration({
      // Additional Replay configuration goes in here, for example:
      maskAllText: true,
      blockAllMedia: true,
    }),
  ],
});
