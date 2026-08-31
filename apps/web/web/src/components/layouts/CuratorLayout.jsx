import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Network, Inbox, ArrowLeft, ShieldCheck } from 'lucide-react';
import Header from './Header';
import { ErrorBoundary } from '../common/ErrorBoundary';

export default function CuratorLayout() {
  const navItems = [
    { label: 'Knowledge Structures', path: '/curator/structures', icon: Network },
    { label: 'Resource Curation Queue', path: '/curator/resources', icon: Inbox },
  ];

  return (
    <div className="min-h-screen bg-transparent flex flex-col">
      <Header />
      <div className="flex-1 flex max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 gap-8">
        <aside className="hidden md:flex flex-col w-64 shrink-0 gap-6">
          <div className="bg-black/40 backdrop-blur-xl border border-purple-200/80 rounded-2xl p-3 shadow-[0_0_10px_rgba(79,70,229,0.1)]">
            <div className="px-3 py-2 border-b border-purple-100 mb-2 flex items-center gap-2 text-purple-900 font-bold text-xs">
              <ShieldCheck className="w-4 h-4 text-purple-600" />
              Curator Console
            </div>
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
                          ? 'bg-purple-900/40 text-purple-400 shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                          : 'text-slate-300 hover:text-white hover:bg-black/30 backdrop-blur-md'
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

          <NavLink
            to="/roadmap"
            className="flex items-center gap-2 text-xs font-semibold text-slate-400 hover:text-white px-3 py-2"
          >
            <ArrowLeft className="w-4 h-4" /> Back to Learner View
          </NavLink>
        </aside>

        <main className="flex-1 min-w-0">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </main>
      </div>
    </div>
  );
}
