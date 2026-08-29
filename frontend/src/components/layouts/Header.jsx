import React, { useState } from 'react';
import { Search, Compass, Shield, User, LogOut, ChevronDown } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import NotificationsPopover from '../cross-cutting/NotificationsPopover';
import GlobalSearchModal from '../cross-cutting/GlobalSearchModal';
import { USER_ROLES } from '../../utils/constants';

export default function Header() {
  const { user, logout, switchRole } = useAuth();
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const navigate = useNavigate();

  return (
    <>
      <header className="sticky top-0 z-30 h-16 bg-white/80 backdrop-blur-md border-b border-slate-200/80 px-4 sm:px-8 flex items-center justify-between">
        {/* Left: Branding */}
        <div className="flex items-center gap-3 cursor-pointer" onClick={() => navigate('/roadmap')}>
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white font-bold shadow-md shadow-blue-500/20">
            <Compass className="w-5 h-5" />
          </div>
          <span className="font-display font-bold text-lg text-slate-900 tracking-tight">
            Amplified<span className="text-blue-600">.AI</span>
          </span>
        </div>

        {/* Center: Quick Search Bar Trigger */}
        <div className="hidden md:flex items-center">
          <button
            onClick={() => setIsSearchOpen(true)}
            className="flex items-center gap-3 px-4 py-2 bg-slate-100/80 hover:bg-slate-100 rounded-full text-xs text-slate-500 transition w-64 border border-slate-200/60"
          >
            <Search className="w-4 h-4 text-slate-400" />
            <span>Search concepts, resources...</span>
            <kbd className="ml-auto text-[10px] bg-white px-1.5 py-0.5 rounded border border-slate-200 text-slate-400 font-mono">
              ⌘K
            </kbd>
          </button>
        </div>

        {/* Right: Role Switcher Demo Bar, Notifications, User Menu */}
        <div className="flex items-center gap-2 sm:gap-4">
          {/* Quick Evaluator Role Toggle */}
          <div className="hidden lg:flex items-center gap-1 bg-slate-100 p-1 rounded-xl text-xs">
            <button
              onClick={() => {
                switchRole(USER_ROLES.LEARNER);
                navigate('/roadmap');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${
                user.role === USER_ROLES.LEARNER
                  ? 'bg-white text-blue-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Learner
            </button>
            <button
              onClick={() => {
                switchRole(USER_ROLES.CURATOR);
                navigate('/curator/structures');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${
                user.role === USER_ROLES.CURATOR
                  ? 'bg-white text-purple-600 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Curator
            </button>
            <button
              onClick={() => {
                switchRole(USER_ROLES.ADMIN);
                navigate('/admin/users');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${
                user.role === USER_ROLES.ADMIN
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              Admin
            </button>
          </div>

          <NotificationsPopover />

          {/* User Profile Dropdown */}
          <div className="relative">
            <button
              onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
              className="flex items-center gap-2 p-1.5 rounded-xl hover:bg-slate-100 transition"
            >
              <img
                src={user.avatarUrl}
                alt={user.fullName}
                className="w-8 h-8 rounded-lg object-cover ring-1 ring-slate-200"
              />
              <span className="hidden sm:inline text-xs font-semibold text-slate-700">
                {user.fullName}
              </span>
              <ChevronDown className="w-3.5 h-3.5 text-slate-400" />
            </button>

            {isUserMenuOpen && (
              <>
                <div className="fixed inset-0 z-20" onClick={() => setIsUserMenuOpen(false)}></div>
                <div className="absolute right-0 mt-2 w-56 bg-white border border-slate-200 rounded-2xl shadow-elevated z-30 p-2 text-xs">
                  <div className="px-3 py-2 border-b border-slate-100">
                    <p className="font-semibold text-slate-900">{user.fullName}</p>
                    <p className="text-slate-500 text-[11px] truncate">{user.email}</p>
                    <span className="inline-block mt-1 px-2 py-0.5 bg-blue-50 text-blue-700 font-semibold rounded text-[10px] uppercase">
                      {user.role}
                    </span>
                  </div>

                  <div className="py-1">
                    <button
                      onClick={() => {
                        setIsUserMenuOpen(false);
                        navigate('/settings');
                      }}
                      className="w-full px-3 py-2 rounded-lg text-slate-700 hover:bg-slate-50 flex items-center gap-2 text-left"
                    >
                      <User className="w-4 h-4 text-slate-400" /> Settings & Preferences
                    </button>
                    {user.role !== USER_ROLES.LEARNER && (
                      <button
                        onClick={() => {
                          setIsUserMenuOpen(false);
                          navigate(user.role === USER_ROLES.ADMIN ? '/admin/users' : '/curator/structures');
                        }}
                        className="w-full px-3 py-2 rounded-lg text-purple-700 hover:bg-purple-50 flex items-center gap-2 text-left font-medium"
                      >
                        <Shield className="w-4 h-4 text-purple-600" /> Management Console
                      </button>
                    )}
                  </div>

                  <div className="pt-1 border-t border-slate-100">
                    <button
                      onClick={() => {
                        setIsUserMenuOpen(false);
                        logout();
                        navigate('/login');
                      }}
                      className="w-full px-3 py-2 rounded-lg text-red-600 hover:bg-red-50 flex items-center gap-2 text-left"
                    >
                      <LogOut className="w-4 h-4 text-red-500" /> Log out
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </header>

      <GlobalSearchModal isOpen={isSearchOpen} onClose={() => setIsSearchOpen(false)} />
    </>
  );
}
