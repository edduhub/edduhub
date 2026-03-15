"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { User, AuthSession } from './types';
import { logger } from './logger';

type AuthContextType = {
  user: User | null;
  session: AuthSession | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<AuthSession>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
};

type RegisterData = {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  role: string;
  collegeId: string;
  collegeName: string;
  rollNo: string;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const DEFAULT_SESSION_TTL_MS = 24 * 60 * 60 * 1000;
const VALID_USER_ROLES = new Set<User['role']>(['student', 'faculty', 'admin', 'super_admin', 'parent']);
const STORAGE_KEY = 'edduhub_auth_session';

function saveSessionToStorage(session: AuthSession): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
    localStorage.setItem('auth_token', session.token);
  } catch (error) {
    logger.error('Failed to save session to localStorage', error as Error);
  }
}

function loadSessionFromStorage(): AuthSession | null {
  if (typeof window === 'undefined') return null;
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return null;
    const session = JSON.parse(stored) as AuthSession;
    // Check if session is expired
    if (session.expiresAt && new Date(session.expiresAt) < new Date()) {
      clearSessionFromStorage();
      return null;
    }
    return session;
  } catch (error) {
    logger.error('Failed to load session from localStorage', error as Error);
    return null;
  }
}

function clearSessionFromStorage(): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem('auth_token');
  } catch (error) {
    logger.error('Failed to clear session from localStorage', error as Error);
  }
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    return null;
  }
  return value as Record<string, unknown>;
}

function firstString(...values: unknown[]): string | undefined {
  for (const value of values) {
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (trimmed.length > 0) return trimmed;
      continue;
    }
    if (typeof value === 'number') {
      return String(value);
    }
  }
  return undefined;
}

function getBoolean(...values: unknown[]): boolean | undefined {
  for (const value of values) {
    if (typeof value === 'boolean') {
      return value;
    }
  }
  return undefined;
}

function unwrapData(payload: unknown): unknown {
  const root = asRecord(payload);
  if (!root) return payload;
  return Object.prototype.hasOwnProperty.call(root, 'data') ? root.data : payload;
}

function extractToken(payload: unknown): string {
  const data = asRecord(payload);
  return (data && firstString(data.token, data.accessToken, data.access_token)) || '';
}

function extractRefreshToken(payload: unknown): string | undefined {
  const data = asRecord(payload);
  return data ? firstString(data.refreshToken, data.refresh_token) : undefined;
}

function extractExpiresAt(payload: unknown): string {
  const data = asRecord(payload);
  const expiresIn = data?.expiresIn;
  if (typeof expiresIn === 'number' && expiresIn > 0) {
    return new Date(Date.now() + expiresIn * 1000).toISOString();
  }
  return (
    firstString(data?.expiresAt, data?.expires_at) ||
    new Date(Date.now() + DEFAULT_SESSION_TTL_MS).toISOString()
  );
}



function normalizeUser(payload: unknown, fallbackUser: User | null = null): User | null {
  const data = asRecord(payload);
  if (!data) return fallbackUser;

  const nestedUser = asRecord(data.user);
  const source = nestedUser || data;

  const traits = asRecord(source.traits) || asRecord(data.traits);
  const name = asRecord(traits?.name);
  const college = asRecord(traits?.college);

  const id = firstString(source.id, data.id, fallbackUser?.id);
  const email = firstString(source.email, data.email, traits?.email, fallbackUser?.email);
  const firstName = firstString(
    source.firstName,
    source.first_name,
    data.firstName,
    data.first_name,
    name?.first,
    name?.firstName,
    fallbackUser?.firstName
  );
  const lastName = firstString(
    source.lastName,
    source.last_name,
    data.lastName,
    data.last_name,
    name?.last,
    name?.lastName,
    fallbackUser?.lastName
  );
  const collegeId = firstString(
    source.collegeId,
    source.college_id,
    data.collegeId,
    data.college_id,
    college?.id,
    fallbackUser?.collegeId
  );
  const collegeName = firstString(
    source.collegeName,
    source.college_name,
    data.collegeName,
    data.college_name,
    college?.name,
    fallbackUser?.collegeName,
    collegeId
  );

  const roleValue = firstString(source.role, data.role, traits?.role, fallbackUser?.role);
  const role: User['role'] =
    roleValue && VALID_USER_ROLES.has(roleValue as User['role'])
      ? (roleValue as User['role'])
      : (fallbackUser?.role || 'student');

  if (!id || !email || !firstName || !lastName || !collegeId || !collegeName) {
    return fallbackUser;
  }

  const verified =
    getBoolean(source.verified, data.verified, traits?.verified, fallbackUser?.verified) ?? false;
  const avatar = firstString(source.avatar, data.avatar, fallbackUser?.avatar);

  return {
    id,
    email,
    firstName,
    lastName,
    role,
    collegeId,
    collegeName,
    verified,
    ...(avatar ? { avatar } : {}),
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<AuthSession | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const abortController = new AbortController();

    const bootstrap = async () => {
      // First, try to restore session from localStorage
      const storedSession = loadSessionFromStorage();
      if (storedSession && storedSession.user) {
        setSession(storedSession);
        setUser(storedSession.user);
        setIsLoading(false);
        return;
      }

      // If no stored session, try to fetch from server
      try {
        const resp = await fetch(`${API_BASE}/auth/session`, {
          method: 'GET',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          signal: abortController.signal,
        });

        // Check if request was aborted
        if (abortController.signal.aborted) return;

        if (resp.ok) {
          const result = await resp.json();
          const payload = unwrapData(result);
          const userData = normalizeUser(payload, null);
          const token = extractToken(payload) || '';
          const expiresAt = extractExpiresAt(payload);
          if (userData) {
            const newSession: AuthSession = {
              token,
              refreshToken: extractRefreshToken(payload),
              user: userData,
              expiresAt,
            };
            setUser(userData);
            setSession(newSession);
            saveSessionToStorage(newSession);
            setIsLoading(false);
            return;
          }

          logger.warn('Auth session response missing user fields; forcing unauthenticated state', {
            status: resp.status,
          });
        }
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') return;
        logger.error('Bootstrap auth check failed', error as Error);
      }

      // Check if component was unmounted during fetch
      if (abortController.signal.aborted) return;

      setIsLoading(false);
    };

    bootstrap();

    // Cleanup: abort any in-flight requests when component unmounts
    return () => {
      abortController.abort();
    };
  }, []);

  const saveSession = (authSession: AuthSession) => {
    setSession(authSession);
    setUser(authSession.user);
    saveSessionToStorage(authSession);
  };

  const clearSession = () => {
    setSession(null);
    setUser(null);
    clearSessionFromStorage();
  };

  const login = async (email: string, password: string): Promise<AuthSession> => {
    try {
      const response = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        let msg = 'Login failed';
        try {
          const error = await response.json();
          msg = error.message || error.error || msg;
        } catch (parseError) {
          logger.error('Failed to parse login error response', parseError as Error);
        }
        throw new Error(msg);
      }

      const result = await response.json();
      const data = unwrapData(result);
      const mappedUser = normalizeUser(data);
      if (!mappedUser) {
        throw new Error('Login response missing user details');
      }

      const authSession: AuthSession = {
        token: extractToken(data),
        refreshToken: extractRefreshToken(data),
        user: mappedUser,
        expiresAt: extractExpiresAt(data),
      };
      if (!authSession.token) {
        throw new Error('Login response missing auth token');
      }

      saveSession(authSession);
      return authSession;
    } catch (error) {
      logger.error('Login failed', error as Error, { email });
      throw error;
    }
  };

  const register = async (data: RegisterData) => {
    try {
      const response = await fetch(`${API_BASE}/auth/register/complete`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        let msg = 'Registration failed';
        try {
          const error = await response.json();
          msg = error.message || error.error || msg;
        } catch (parseError) {
          logger.error('Failed to parse registration error response', parseError as Error);
        }
        throw new Error(msg);
      }

      const result = await response.json();
      const responseData = unwrapData(result);
      const token = extractToken(responseData);
      const mappedUser = normalizeUser(responseData);

      if (!token || !mappedUser) {
        // Backend might not return a token upon registration. In that case, automatically log in.
        await login(data.email, data.password);
        return;
      }

      const authSession: AuthSession = {
        token,
        refreshToken: extractRefreshToken(responseData),
        user: mappedUser,
        expiresAt: extractExpiresAt(responseData),
      };

      saveSession(authSession);
    } catch (error) {
      logger.error('Registration failed', error as Error, { email: data.email });
      throw error;
    }
  };

  const logout = async () => {
    try {
      await fetch(`${API_BASE}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      logger.error('Logout failed', error as Error);
    } finally {
      clearSession();
    }
  };

  const refreshSession = async () => {
    try {
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });

      if (response.ok) {
        const result = await response.json();
        const data = unwrapData(result);
        const refreshedToken = extractToken(data);
        if (refreshedToken && (session || user)) {
          const mappedUser = normalizeUser(data, session?.user || user || null);
          if (!mappedUser) {
            logger.warn('Refresh response missing user fields; keeping existing session user');
            return;
          }

          const currentSession = session || { token: '', user: mappedUser, expiresAt: '' };
          const updatedSession: AuthSession = {
            ...currentSession,
            token: refreshedToken,
            refreshToken: extractRefreshToken(data) || currentSession.refreshToken,
            user: mappedUser,
            expiresAt: extractExpiresAt(data),
          };
          saveSession(updatedSession);
        }
      } else if (response.status === 401) {
        clearSession();
      }
    } catch (error) {
      logger.error('Session refresh failed', error as Error);
      clearSession();
    }
  };

  const value = {
    user,
    session,
    isLoading,
    isAuthenticated: !!user,
    login,
    register,
    logout,
    refreshSession,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
