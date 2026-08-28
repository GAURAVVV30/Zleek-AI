import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Users, FileText, ArrowLeft, ShieldAlert } from 'lucide-react';
import Header from './Header';
import { ErrorBoundary } from '../common/ErrorBoundary';

export default function AdminLayout() {
  const navItems = [
    { label: 'User & Role Management', path: '/admin/users', icon: Users },
    { label: 'Audit Logs', path: '/admin/audit', icon: FileText },
  ];

  return (
    <div className="min-h-screen bg-[#f8fafc] flex flex-col">
      <Header />
      <div className="flex-1 flex max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 gap-8">
        <aside className="hidden md:flex flex-col w-64 shrink-0 gap-6">
          <div className="bg-white border border-slate-200/80 rounded-2xl p-3 shadow-sm">
            <div className="px-3 py-2 border-b border-slate-100 mb-2 flex items-center gap-2 text-slate-900 font-bold text-xs">
              <ShieldAlert className="w-4 h-4 text-slate-800" />
              Platform Admin
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
                          ? 'bg-slate-100 text-slate-900 shadow-sm'
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

          <NavLink
            to="/roadmap"
            className="flex items-center gap-2 text-xs font-semibold text-slate-500 hover:text-slate-900 px-3 py-2"
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
