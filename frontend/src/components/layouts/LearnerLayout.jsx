import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Map, BarChart3, Settings, Trophy, HelpCircle } from 'lucide-react';
import Header from './Header';
import { ErrorBoundary } from '../common/ErrorBoundary';

export default function LearnerLayout() {
  const navItems = [
    { label: 'Personalized Roadmap', path: '/roadmap', icon: Map },
    { label: 'Progress & Competency', path: '/progress', icon: BarChart3 },
    { label: 'Goal Achieved', path: '/goal-achieved', icon: Trophy },
    { label: 'Profile & Settings', path: '/settings', icon: Settings },
  ];

  return (
    <div className="min-h-screen bg-[#f8faff] flex flex-col">
      <Header />
      <div className="flex-1 flex max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 gap-8">
        {/* Persistent Left Sidebar */}
        <aside className="hidden md:flex flex-col w-64 shrink-0 gap-6">
          <div className="bg-white border border-slate-200/80 rounded-2xl p-3 shadow-sm">
            <nav className="space-y-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    className={({ isActive }) =>
                      `flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-xs font-semibold transition ${
                        isActive
                          ? 'bg-blue-50 text-blue-600 shadow-sm'
                          : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                      }`
                    }
                  >
                    <Icon className="w-4 h-4" />
                    <span>{item.label}</span>
                  </NavLink>
                );
              })}
            </nav>
          </div>

          {/* Quick Help & Explainability Widget */}
          <div className="bg-gradient-to-br from-blue-50 to-indigo-50/50 border border-blue-100 rounded-2xl p-4 text-xs">
            <div className="flex items-center gap-2 text-blue-800 font-bold mb-1">
              <HelpCircle className="w-4 h-4 text-blue-600" />
              Evidence-Gated Learning
            </div>
            <p className="text-slate-600 leading-relaxed text-[11px]">
              Progress advances only when understanding is demonstrated through quizzes or projects — never by clicking next.
            </p>
          </div>
        </aside>

        {/* Main Content View with Error Boundary */}
        <main className="flex-1 min-w-0">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </main>
      </div>
    </div>
  );
}
