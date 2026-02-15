export const mockSanitize = jest.fn((input: string) => input);

jest.mock('dompurify', () => ({
  __esModule: true,
  default: {
    sanitize: mockSanitize,
    addHook: jest.fn(),
    removeHook: jest.fn(),
  },
}));
