import { LogLevel, LogEntry } from './types';

/**
 * Logger utility for structured logging throughout the application
 * Automatically disabled in production unless explicitly enabled
 */
class Logger {
  private enabled: boolean;
  private minLevel: LogLevel;
  private logs: LogEntry[] = [];
  private maxLogsInMemory = 100;

  private levelPriority: Record<LogLevel, number> = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
  };

  constructor() {
    this.enabled = process.env.NODE_ENV === 'development' ||
                   process.env.NEXT_PUBLIC_ENABLE_LOGGING === 'true';
    this.minLevel = (process.env.NEXT_PUBLIC_LOG_LEVEL as LogLevel) || 'info';
  }

  private shouldLog(level: LogLevel): boolean {
    if (!this.enabled) return false;
    return this.levelPriority[level] >= this.levelPriority[this.minLevel];
  }

  private createLogEntry(
    level: LogLevel,
    message: string,
    context?: Record<string, any>,
    error?: Error
  ): LogEntry {
    return {
      timestamp: new Date().toISOString(),
      level,
      message,
      context,
      error,
    };
  }

  private addToMemory(entry: LogEntry): void {
    this.logs.push(entry);
    if (this.logs.length > this.maxLogsInMemory) {
      this.logs.shift();
    }
  }

  private formatMessage(entry: LogEntry): string {
    const { timestamp, level, message, context, error } = entry;
    const timeStr = new Date(timestamp).toLocaleTimeString();
    let formatted = `[${timeStr}] [${level.toUpperCase()}] ${message}`;

    if (context && Object.keys(context).length > 0) {
      formatted += ` | Context: ${JSON.stringify(context)}`;
    }

    if (error) {
      formatted += ` | Error: ${error.message}`;
    }

    return formatted;
  }

  private writeLog(entry: LogEntry): void {
    if (!this.shouldLog(entry.level)) return;

    this.addToMemory(entry);

    const formatted = this.formatMessage(entry);

    // Only use console in development or when explicitly enabled
    if (this.enabled) {
      switch (entry.level) {
        case 'debug':
          console.debug(formatted, entry.context, entry.error);
          break;
        case 'info':
          console.info(formatted, entry.context);
          break;
        case 'warn':
          console.warn(formatted, entry.context);
          break;
        case 'error':
          console.error(formatted, entry.context, entry.error);
          break;
      }
    }

    // In production, you could send errors to a monitoring service
    if (process.env.NODE_ENV === 'production' && entry.level === 'error') {
      this.sendToMonitoringService(entry);
    }
  }

  private sendToMonitoringService(entry: LogEntry): void {
    // Integration with Sentry for production error tracking
    try {
      if (typeof window !== 'undefined' && entry.level === 'error') {
        // Client-side: Use @sentry/nextjs
        const Sentry = require('@sentry/nextjs');
        
        if (entry.error) {
          Sentry.captureException(entry.error, {
            level: 'error',
            extra: entry.context,
            tags: {
              component: 'Logger',
            },
          });
        } else {
          Sentry.captureMessage(entry.message, {
            level: 'error',
            extra: entry.context,
          });
        }
      }
    } catch (err) {
      // Silently fail - don't break the app if monitoring fails
      console.error('Failed to send to monitoring service:', err);
    }
  }

  /**
   * Log debug message - lowest priority, only in development
   */
  debug(message: string, context?: Record<string, any>): void {
    const entry = this.createLogEntry('debug', message, context);
    this.writeLog(entry);
  }

  /**
   * Log info message - general information
   */
  info(message: string, context?: Record<string, any>): void {
    const entry = this.createLogEntry('info', message, context);
    this.writeLog(entry);
  }

  /**
   * Log warning message - something unexpected but not breaking
   */
  warn(message: string, context?: Record<string, any>): void {
    const entry = this.createLogEntry('warn', message, context);
    this.writeLog(entry);
  }

  /**
   * Log error message - something broke
   */
  error(message: string, error?: Error, context?: Record<string, any>): void {
    const entry = this.createLogEntry('error', message, context, error);
    this.writeLog(entry);
  }

  /**
   * Get recent logs from memory (useful for debugging)
   */
  getRecentLogs(count?: number): LogEntry[] {
    if (count) {
      return this.logs.slice(-count);
    }
    return [...this.logs];
  }

  /**
   * Clear logs from memory
   */
  clearLogs(): void {
    this.logs = [];
  }

  /**
   * Export logs as JSON
   */
  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }

  /**
   * Enable/disable logging at runtime
   */
  setEnabled(enabled: boolean): void {
    this.enabled = enabled;
  }

  /**
   * Set minimum log level
   */
  setMinLevel(level: LogLevel): void {
    this.minLevel = level;
  }
}

// Export singleton instance
export const logger = new Logger();

// Export convenience functions
export const log = {
  debug: (message: string, context?: Record<string, any>) =>
    logger.debug(message, context),
  info: (message: string, context?: Record<string, any>) =>
    logger.info(message, context),
  warn: (message: string, context?: Record<string, any>) =>
    logger.warn(message, context),
  error: (message: string, error?: Error, context?: Record<string, any>) =>
    logger.error(message, error, context),
};

export default logger;
