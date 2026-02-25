"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { User, AuthSession } from './types';
import { logger } from './logger';

type AuthContextType = {
  user: User | null;
  session: AuthSession | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
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

const AUTH_STORAGE_KEY = 'edduhub_auth';
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const DEFAULT_SESSION_TTL_MS = 24 * 60 * 60 * 1000;
const VALID_USER_ROLES = new Set<User['role']>(['student', 'faculty', 'admin', 'super_admin', 'parent']);

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
  return (data && firstString(data.token)) || '';
}

function extractExpiresAt(payload: unknown): string {
  const data = asRecord(payload);
  return (
    firstString(data?.expiresAt, data?.expires_at) ||
    new Date(Date.now() + DEFAULT_SESSION_TTL_MS).toISOString()
  );
}

function extractMessage(payload: unknown, fallback: string): string {
  const data = asRecord(payload);
  return firstString(data?.message, data?.error, fallback) || fallback;
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
      const storedAuth = localStorage.getItem(AUTH_STORAGE_KEY);
      let storedSession: AuthSession | null = null;

      if (storedAuth) {
        try {
          const parsed = asRecord(JSON.parse(storedAuth));
          const parsedUser = normalizeUser(parsed?.user || null);
          const parsedExpiresAt = firstString(parsed?.expiresAt, parsed?.expires_at);
          const parsedToken = firstString(parsed?.token) || '';

          if (parsedUser && parsedExpiresAt) {
            storedSession = {
              token: parsedToken,
              user: parsedUser,
              expiresAt: parsedExpiresAt,
            };
          } else {
            logger.warn('Stored auth payload is missing required fields', { key: AUTH_STORAGE_KEY });
            localStorage.removeItem(AUTH_STORAGE_KEY);
          }
        } catch (error) {
          logger.error('Failed to parse stored auth during bootstrap', error as Error, { key: AUTH_STORAGE_KEY });
          localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      }

      const storedToken = storedSession?.token || '';

      try {
        // Resolve current identity from the auth callback endpoint.
        const headers: Record<string, string> = { 'Content-Type': 'application/json' };
        if (storedToken) {
          headers.Authorization = `Bearer ${storedToken}`;
        }

        const resp = await fetch(`${API_BASE}/auth/callback`, {
          method: 'GET',
          credentials: 'include',
          headers,
          signal: abortController.signal,
        });

        // Check if request was aborted
        if (abortController.signal.aborted) return;

        if (resp.ok) {
          const result = await resp.json();
          const userData = normalizeUser(unwrapData(result), storedSession?.user || null);
          if (userData) {
            setUser(userData);
            setSession({
              token: storedToken,
              user: userData,
              expiresAt:
                storedSession?.expiresAt ||
                new Date(Date.now() + DEFAULT_SESSION_TTL_MS).toISOString(),
            });
            setIsLoading(false);
            return;
          }

          logger.warn('Auth callback response missing user fields; falling back to stored session');
        } else if (storedToken && (resp.status === 401 || resp.status === 403)) {
          logger.warn('Stored token is no longer valid; clearing local session cache', {
            status: resp.status,
          });
          storedSession = null;
          localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') return;
        logger.error('Bootstrap auth check failed', error as Error);
      }

      // Check if component was unmounted during fetch
      if (abortController.signal.aborted) return;

      // Fallback: Load session from localStorage
      if (storedSession) {
        try {
          const authData: AuthSession = storedSession;
          if (new Date(authData.expiresAt) > new Date()) {
            setSession(authData);
            setUser(authData.user);
          } else {
            localStorage.removeItem(AUTH_STORAGE_KEY);
          }
        } catch (error) {
          logger.error('Failed to parse stored auth', error as Error, { key: AUTH_STORAGE_KEY });
          localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      }
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
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(authSession));
  };

  const clearSession = () => {
    setSession(null);
    setUser(null);
    localStorage.removeItem(AUTH_STORAGE_KEY);
  };

  const login = async (email: string, password: string) => {
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
        user: mappedUser,
        expiresAt: extractExpiresAt(data),
      };
      if (!authSession.token) {
        throw new Error('Login response missing auth token');
      }

      saveSession(authSession);
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
        throw new Error(
          extractMessage(responseData, 'Registration successful. Please sign in with your credentials.')
        );
      }

      const authSession: AuthSession = {
        token,
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
        headers: session?.token ? { 'Authorization': `Bearer ${session.token}` } : undefined,
      });
    } catch (error) {
      logger.error('Logout failed', error as Error);
    } finally {
      clearSession();
    }
  };

  const refreshSession = async () => {
    try {
      const headers: Record<string, string> = {};
      if (session?.token) headers['Authorization'] = `Bearer ${session.token}`;
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers,
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
            user: mappedUser,
            expiresAt: new Date(Date.now() + DEFAULT_SESSION_TTL_MS).toISOString(),
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
