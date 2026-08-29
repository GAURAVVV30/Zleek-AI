import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Compass } from 'lucide-react';
import { ErrorBoundary } from '../common/ErrorBoundary';

export default function AuthLayout() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-[#edf4fe] via-[#f8faff] to-[#ffffff] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center mb-6">
        <NavLink to="/" className="inline-flex items-center gap-2.5">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white font-bold shadow-md shadow-blue-500/20">
            <Compass className="w-6 h-6" />
          </div>
          <span className="font-display font-bold text-2xl text-slate-900 tracking-tight">
            Amplified<span className="text-blue-600">.AI</span>
          </span>
        </NavLink>
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md px-4 sm:px-0">
        <div className="bg-white/90 backdrop-blur-md py-8 px-6 sm:px-10 shadow-elevated border border-slate-200/80 rounded-3xl">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </div>
      </div>
    </div>
  );
}
