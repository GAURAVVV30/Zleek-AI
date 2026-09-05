import React, { useState } from 'react';
import { Search, Compass, Shield, User, LogOut, ChevronDown } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import NotificationsPopover from '../cross-cutting/NotificationsPopover';
import AnimatedSearchBar from '../ui/animated-search-bar';
import { USER_ROLES } from '../../utils/constants';

export default function Header() {
  const { user, logout, switchRole } = useAuth();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const navigate = useNavigate();

  return (
    <>
      <header className="sticky top-0 z-30 h-16 bg-black/40 backdrop-blur-xl/80 backdrop-blur-md border-b border-white/10 px-4 sm:px-8 flex items-center justify-between">
        {/* Left: Branding */}
        <div className="flex items-center gap-3 cursor-pointer" onClick={() => navigate('/roadmap')}>
          <div className="w-10 h-10 flex items-center justify-center overflow-hidden mix-blend-screen">
            <img src="/logo-icon.png" alt="Zleek AI Logo" className="w-full h-full object-contain" />
          </div>
          <span className="font-display font-bold text-xl text-white tracking-tight">
            Zleek<span className="text-indigo-400"> AI</span>
          </span>
        </div>

        {/* Spacer to push everything else to the right */}
        <div className="flex-1" />

        {/* Right: Role Switcher Demo Bar, Notifications, User Menu */}
        <div className="flex items-center gap-2 sm:gap-4">
          {/* Quick Evaluator Role Toggle */}
          <div className="hidden lg:flex items-center gap-1 bg-slate-100 p-1 rounded-xl text-xs">
            <button
              onClick={() => {
                switchRole(USER_ROLES.LEARNER);
                navigate('/roadmap');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${user.role === USER_ROLES.LEARNER
                  ? 'bg-black/40 backdrop-blur-xl text-indigo-400 shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                  : 'text-slate-300 hover:text-white'
                }`}
            >
              Learner
            </button>
            <button
              onClick={() => {
                switchRole(USER_ROLES.CURATOR);
                navigate('/curator/structures');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${user.role === USER_ROLES.CURATOR
                  ? 'bg-black/40 backdrop-blur-xl text-purple-600 shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                  : 'text-slate-300 hover:text-white'
                }`}
            >
              Curator
            </button>
            <button
              onClick={() => {
                switchRole(USER_ROLES.ADMIN);
                navigate('/admin/users');
              }}
              className={`px-2.5 py-1 rounded-lg font-medium transition ${user.role === USER_ROLES.ADMIN
                  ? 'bg-black/40 backdrop-blur-xl text-white shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                  : 'text-slate-300 hover:text-white'
                }`}
            >
              Admin
            </button>
          </div>
          
          <AnimatedSearchBar />

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
              <span className="hidden sm:inline text-xs font-semibold text-white">
                {user.fullName}
              </span>
              <ChevronDown className="w-3.5 h-3.5 text-slate-400" />
            </button>

            {isUserMenuOpen && (
              <>
                <div className="fixed inset-0 z-20" onClick={() => setIsUserMenuOpen(false)}></div>
                <div className="absolute right-0 mt-2 w-56 bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl shadow-elevated z-30 p-2 text-xs">
                  <div className="px-3 py-2 border-b border-white/5">
                    <p className="font-semibold text-white">{user.fullName}</p>
                    <p className="text-slate-400 text-[11px] truncate">{user.email}</p>
                    <span className="inline-block mt-1 px-2 py-0.5 bg-indigo-900/40 backdrop-blur-sm text-indigo-400 font-semibold rounded text-[10px] uppercase">
                      {user.role}
                    </span>
                  </div>

                  <div className="py-1">
                    <button
                      onClick={() => {
                        setIsUserMenuOpen(false);
                        navigate('/settings');
                      }}
                      className="w-full px-3 py-2 rounded-lg text-white hover:bg-black/30 backdrop-blur-md flex items-center gap-2 text-left"
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

                  <div className="pt-1 border-t border-white/5">
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
    </>
  );
}
