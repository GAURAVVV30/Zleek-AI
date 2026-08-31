import React, { createContext, useContext, useState, useEffect } from 'react';
import { apiClient } from '../services/apiClient';
import { ENDPOINTS } from '../utils/endpoints';
import { USER_ROLES } from '../utils/constants';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      setUser(null);
      setIsLoading(false);
      return;
    }
    // Initial fetch of current user
    apiClient
      .get(ENDPOINTS.AUTH.ME)
      .then((res) => {
        if (res.data) {
          setUser(res.data);
        } else {
          setUser(null);
        }
        setIsLoading(false);
      })
      .catch(() => {
        setUser(null);
        setIsLoading(false);
      });
  }, []);

  const login = async (email, password) => {
    setIsLoading(true);
    try {
      const res = await apiClient.post(ENDPOINTS.AUTH.LOGIN, { email, password });
      if (res.data?.user) {
        setUser(res.data.user);
        localStorage.setItem('access_token', res.data.accessToken);
      }
      return res;
    } finally {
      setIsLoading(false);
    }
  };

  const signup = async (email, password, fullName) => {
    setIsLoading(true);
    try {
      const res = await apiClient.post(ENDPOINTS.AUTH.SIGNUP, { email, password, fullName });
      if (res.data?.user) {
        setUser(res.data.user);
        localStorage.setItem('access_token', res.data.accessToken);
      }
      return res;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await apiClient.post(ENDPOINTS.AUTH.LOGOUT);
    } finally {
      localStorage.removeItem('access_token');
      setUser(null);
    }
  };

  // Demo role switcher for instant judge/evaluator inspection
  const switchRole = (newRole) => {
    setUser((prev) => ({ ...prev, role: newRole }));
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        signup,
        logout,
        switchRole,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
