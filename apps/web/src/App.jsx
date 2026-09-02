import React from 'react';
import { BrowserRouter, useLocation } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from './lib/queryClient';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import AppRoutes from './routes/AppRoutes';

function GlobalBackground() {
  const location = useLocation();
  const isLandingPage = location.pathname === '/';

  return (
    <div 
      className={`fixed inset-0 pointer-events-none -z-10 ${
        isLandingPage 
          ? 'bg-black/20 backdrop-blur-md' 
          : 'bg-black/25 backdrop-blur-sm'
      }`} 
    />
  );
}

export default function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AuthProvider>
            <ToastProvider>
              <>
                <GlobalBackground />
                <div className="relative min-h-screen w-full text-white font-sans selection:bg-indigo-500/30">
                  <AppRoutes />
                </div>
              </>
            </ToastProvider>
          </AuthProvider>
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
