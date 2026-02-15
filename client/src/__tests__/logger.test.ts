import { logger, log } from '@/lib/logger';

describe('Logger', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(console, 'debug').mockImplementation(() => {});
    jest.spyOn(console, 'info').mockImplementation(() => {});
    jest.spyOn(console, 'warn').mockImplementation(() => {});
    jest.spyOn(console, 'error').mockImplementation(() => {});
    logger.setEnabled(true);
    logger.setMinLevel('debug');
  });

  afterEach(() => {
    logger.clearLogs();
  });

  describe('debug', () => {
    it('logs debug message', () => {
      logger.debug('Debug message', { key: 'value' });
      expect(console.debug).toHaveBeenCalled();
    });
  });

  describe('info', () => {
    it('logs info message', () => {
      logger.info('Info message', { key: 'value' });
      expect(console.info).toHaveBeenCalled();
    });
  });

  describe('warn', () => {
    it('logs warning message', () => {
      logger.warn('Warning message', { key: 'value' });
      expect(console.warn).toHaveBeenCalled();
    });
  });

  describe('error', () => {
    it('logs error message with error object', () => {
      const error = new Error('Test error');
      logger.error('Error occurred', error, { key: 'value' });
      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('getRecentLogs', () => {
    it('returns empty array initially', () => {
      expect(logger.getRecentLogs()).toEqual([]);
    });

    it('returns recent logs', () => {
      logger.info('Message 1');
      logger.info('Message 2');
      const logs = logger.getRecentLogs();
      expect(logs.length).toBe(2);
    });

    it('respects count parameter', () => {
      logger.info('Message 1');
      logger.info('Message 2');
      logger.info('Message 3');
      const logs = logger.getRecentLogs(2);
      expect(logs.length).toBe(2);
    });
  });

  describe('clearLogs', () => {
    it('clears all logs', () => {
      logger.info('Message');
      logger.clearLogs();
      expect(logger.getRecentLogs()).toEqual([]);
    });
  });

  describe('exportLogs', () => {
    it('exports logs as JSON', () => {
      logger.info('Test message');
      const exported = logger.exportLogs();
      expect(exported).toContain('Test message');
    });
  });

  describe('runtime configuration', () => {
    it('setEnabled enables/disables logging', () => {
      logger.setEnabled(false);
      logger.info('Should not log');
      expect(console.info).not.toHaveBeenCalled();

      logger.setEnabled(true);
      logger.info('Should log');
      expect(console.info).toHaveBeenCalled();
    });

    it('setMinLevel filters logs', () => {
      logger.setMinLevel('error');
      logger.debug('Debug');
      logger.info('Info');
      logger.warn('Warn');
      logger.error('Error');

      expect(console.debug).not.toHaveBeenCalled();
      expect(console.info).not.toHaveBeenCalled();
      expect(console.warn).not.toHaveBeenCalled();
      expect(console.error).toHaveBeenCalled();
    });
  });

  describe('log convenience object', () => {
    it('exports debug function', () => {
      log.debug('Debug');
      expect(console.debug).toHaveBeenCalled();
    });

    it('exports info function', () => {
      log.info('Info');
      expect(console.info).toHaveBeenCalled();
    });

    it('exports warn function', () => {
      log.warn('Warn');
      expect(console.warn).toHaveBeenCalled();
    });

    it('exports error function', () => {
      log.error('Error');
      expect(console.error).toHaveBeenCalled();
    });
  });
});
