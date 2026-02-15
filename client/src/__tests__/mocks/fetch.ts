export const mockFetch = jest.fn();
global.fetch = mockFetch;

export function createMockFetch() {
  return mockFetch;
}

export function mockFetchSuccess(data: unknown, options?: { status?: number; headers?: Record<string, string> }) {
  mockFetch.mockResolvedValueOnce({
    ok: true,
    status: options?.status ?? 200,
    json: async () => data,
    headers: new Map(Object.entries(options?.headers ?? { 'content-type': 'application/json' })),
  });
}

export function mockFetchError(status = 400, message = 'Error') {
  mockFetch.mockResolvedValueOnce({
    ok: false,
    status,
    statusText: message,
    json: async () => ({ message }),
  });
}
