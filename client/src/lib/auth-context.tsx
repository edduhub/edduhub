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
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const AUTH_STORAGE_KEY = 'edduhub_auth';
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<AuthSession | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const bootstrap = async () => {
      try {
        // Prefer cookie-based session via server profile endpoint
        const resp = await fetch(`${API_BASE}/api/profile`, {
          method: 'GET',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        });
        if (resp.ok) {
          const result = await resp.json();
          const data = result.data || result;
          // Expect data to be a full profile or user object
          const userData: User = {
            id: data.user?.id ?? data.id,
            email: data.user?.email ?? data.email,
            firstName: data.user?.firstName ?? data.firstName,
            lastName: data.user?.lastName ?? data.lastName,
            role: data.user?.role ?? data.role,
            collegeId: data.user?.collegeId ?? data.collegeId,
            collegeName: data.user?.collegeName ?? data.collegeName,
            verified: data.user?.verified ?? data.verified ?? false,
            avatar: data.user?.avatar ?? data.avatar,
          } as User;
          setUser(userData);
          setSession({ token: '', user: userData, expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString() });
          setIsLoading(false);
          return;
        }
      } catch {
        // ignore and fallback to local storage
      }

      // Fallback: Load session from localStorage
      const storedAuth = localStorage.getItem(AUTH_STORAGE_KEY);
      if (storedAuth) {
        try {
          const authData: AuthSession = JSON.parse(storedAuth);
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
        } catch {}
        throw new Error(msg);
      }

      const result = await response.json();
      const data = result.data || result;

      // Support transition: token may exist, but cookie session is primary
      const authSession: AuthSession = {
        token: data.token || '',
        user: {
          id: data.user.id,
          email: data.user.email,
          firstName: data.user.firstName,
          lastName: data.user.lastName,
          role: data.user.role,
          collegeId: data.user.collegeId,
          collegeName: data.user.collegeName,
          verified: data.user.verified || false,
        },
        expiresAt: data.expiresAt || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      };

      // Immediately refresh user from profile via cookie session (authoritative)
      try {
        const prof = await fetch(`${API_BASE}/api/profile`, { credentials: 'include' });
        if (prof.ok) {
          const profRes = await prof.json();
          const p = profRes.data || profRes;
          authSession.user = {
            id: p.user?.id ?? p.id,
            email: p.user?.email ?? p.email,
            firstName: p.user?.firstName ?? p.firstName,
            lastName: p.user?.lastName ?? p.lastName,
            role: p.user?.role ?? p.role,
            collegeId: p.user?.collegeId ?? p.collegeId,
            collegeName: p.user?.collegeName ?? p.collegeName,
            verified: p.user?.verified ?? p.verified ?? false,
            avatar: p.user?.avatar ?? p.avatar,
          } as User;
        }
      } catch {}

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
        } catch {}
        throw new Error(msg);
      }

      const result = await response.json();
      const responseData = result.data || result;

      const authSession: AuthSession = {
        token: responseData.token || '',
        user: {
          id: responseData.user.id,
          email: responseData.user.email,
          firstName: responseData.user.firstName,
          lastName: responseData.user.lastName,
          role: responseData.user.role,
          collegeId: responseData.user.collegeId,
          collegeName: responseData.user.collegeName,
          verified: responseData.user.verified || false,
        },
        expiresAt: responseData.expiresAt || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
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
        const data = result.data || result;
        if (data.session_token) {
          const updatedSession = {
            ...session,
            token: data.session_token,
            expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
          } as AuthSession;
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