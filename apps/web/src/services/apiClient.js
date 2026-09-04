import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to attach JWT token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for handling 401s
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      const path = window.location.pathname;
      const isPublicPath = ['/', '/login', '/signup', '/forgot-password'].includes(path);
      localStorage.removeItem('access_token');
      if (!isPublicPath) {
        console.warn('Session expired or unauthorized');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error.response?.data || error);
  }
);
