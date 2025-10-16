"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { User, AuthSession } from './types';

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
    // Load session from localStorage on mount
    const storedAuth = localStorage.getItem(AUTH_STORAGE_KEY);
    if (storedAuth) {
      try {
        const authData: AuthSession = JSON.parse(storedAuth);
        // Check if session is expired
        if (new Date(authData.expiresAt) > new Date()) {
          setSession(authData);
          setUser(authData.user);
        } else {
          localStorage.removeItem(AUTH_STORAGE_KEY);
        }
      } catch (error) {
        console.error('Failed to parse stored auth:', error);
        localStorage.removeItem(AUTH_STORAGE_KEY);
      }
    }
    setIsLoading(false);
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
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || error.error || 'Login failed');
      }

      const result = await response.json();
      // Backend returns {data: {token, user, expiresAt}, message}
      const data = result.data || result;
      
      const authSession: AuthSession = {
        token: data.token,
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
      
      saveSession(authSession);
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  const register = async (data: RegisterData) => {
    try {
      const response = await fetch(`${API_BASE}/auth/register/complete`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || error.error || 'Registration failed');
      }

      const result = await response.json();
      const responseData = result.data || result;
      
      const authSession: AuthSession = {
        token: responseData.token,
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
      console.error('Registration error:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      if (session?.token) {
        await fetch(`${API_BASE}/auth/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${session.token}`,
          },
        });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      clearSession();
    }
  };

  const refreshSession = async () => {
    if (!session?.token) return;

    try {
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${session.token}`,
        },
      });

      if (response.ok) {
        const result = await response.json();
        const data = result.data || result;
        if (data.session_token) {
          // Update token
          const updatedSession = {
            ...session,
            token: data.session_token,
            expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
          };
          saveSession(updatedSession);
        }
      } else {
        clearSession();
      }
    } catch (error) {
      console.error('Session refresh error:', error);
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