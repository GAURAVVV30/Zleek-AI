import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Compass, ArrowLeft } from 'lucide-react';
import { ErrorBoundary } from '../common/ErrorBoundary';

export default function AuthLayout() {
  return (
    <div className="min-h-screen bg-transparent flex flex-col justify-center py-12 sm:px-6 lg:px-8 relative">
      {/* Back Button */}
      <div className="absolute top-6 left-6 z-20">
        <NavLink 
          to="/" 
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-black/30 backdrop-blur-md border border-white/10 text-white text-xs font-semibold hover:bg-white/10 hover:shadow-[0_0_15px_rgba(255,255,255,0.1)] transition"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </NavLink>
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center mb-6">
        <NavLink to="/" className="inline-flex items-center gap-2.5 relative z-20">
          <div className="w-12 h-12 flex items-center justify-center overflow-hidden mix-blend-screen">
            <img src="/logo-icon.png" alt="Zleek AI Logo" className="w-full h-full object-contain" />
          </div>
          <span className="font-display font-bold text-2xl text-white tracking-tight">
            Zleek<span className="text-indigo-400"> AI</span>
          </span>
        </NavLink>
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md px-4 sm:px-0">
        <div className="bg-black/50 backdrop-blur-md py-8 px-6 sm:px-10 shadow-[0_0_25px_rgba(79,70,229,0.15)] border border-white/10 rounded-3xl">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </div>
      </div>
    </div>
  );
}
