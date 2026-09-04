import React, { useState, useEffect } from 'react';
import { Bell, Check, Sparkles, CheckCircle, RotateCcw, Info, BookOpen } from 'lucide-react';
import { getNotifications, markAllNotificationsRead } from '../../services/notificationService';

export default function NotificationsPopover() {
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);

  const reloadNotifications = () => {
    setNotifications(getNotifications());
  };

  useEffect(() => {
    reloadNotifications();
    const handleUpdate = () => reloadNotifications();
    window.addEventListener('platform_notifications_updated', handleUpdate);
    return () => {
      window.removeEventListener('platform_notifications_updated', handleUpdate);
    };
  }, []);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const handleMarkAllRead = () => {
    markAllNotificationsRead();
  };

  return (
    <div
      className="relative inline-block"
      onMouseEnter={() => setIsOpen(true)}
      onMouseLeave={() => setIsOpen(false)}
    >
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-slate-300 hover:text-white hover:bg-white/10 rounded-xl transition flex items-center gap-1.5 focus:outline-none"
        title="Hover to view notifications"
      >
        <div className="relative flex items-center justify-center">
          <Bell className={`w-5 h-5 transition-transform duration-300 ${unreadCount > 0 ? 'text-indigo-400 animate-bounce' : 'text-slate-300'}`} />
          {unreadCount > 0 && (
            <span className="absolute -top-1.5 -right-2 px-1.5 py-0.2 text-[10px] font-extrabold bg-rose-500 text-white rounded-full min-w-[18px] h-[18px] flex items-center justify-center shadow-lg border border-slate-900 animate-pulse">
              {unreadCount}
            </span>
          )}
        </div>
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full pt-1 w-80 sm:w-96 z-50 animate-in fade-in zoom-in-95 duration-150">
          <div className="bg-slate-950/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-[0_0_30px_rgba(79,70,229,0.25)] overflow-hidden">
            <div className="p-4 border-b border-white/10 flex items-center justify-between bg-black/40">
              <div className="flex items-center gap-2">
                <span className="font-bold text-white text-xs uppercase tracking-wider">Notifications</span>
                {unreadCount > 0 ? (
                  <span className="px-2 py-0.5 text-[10px] font-extrabold bg-indigo-500/20 border border-indigo-500/30 text-indigo-300 rounded-full">
                    {unreadCount} new
                  </span>
                ) : (
                  <span className="px-2 py-0.5 text-[10px] font-medium bg-slate-800 text-slate-400 rounded-full">
                    Up to date
                  </span>
                )}
              </div>

              {unreadCount > 0 && (
                <button
                  onClick={handleMarkAllRead}
                  className="text-[11px] text-indigo-400 hover:text-indigo-300 font-semibold flex items-center gap-1 transition"
                >
                  <Check className="w-3.5 h-3.5" /> Mark all read
                </button>
              )}
            </div>

            <div className="max-h-80 overflow-y-auto divide-y divide-white/5">
              {notifications.length > 0 ? (
                notifications.map((notif) => (
                  <div
                    key={notif.id}
                    className={`p-4 hover:bg-white/5 transition flex gap-3 text-left ${
                      !notif.read ? 'bg-indigo-950/40 border-l-2 border-indigo-500' : ''
                    }`}
                  >
                    <div className="w-8 h-8 rounded-full bg-black/50 border border-white/10 flex items-center justify-center shrink-0 mt-0.5">
                      {notif.type === 'success' || notif.title?.includes('Completed') ? (
                        <CheckCircle className="w-4 h-4 text-emerald-400" />
                      ) : notif.type === 'warning' || notif.title?.includes('Reset') ? (
                        <RotateCcw className="w-4 h-4 text-rose-400" />
                      ) : notif.type === 'path_update' ? (
                        <Sparkles className="w-4 h-4 text-indigo-400" />
                      ) : (
                        <Info className="w-4 h-4 text-indigo-400" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-bold text-white leading-tight">{notif.title}</p>
                      <p className="text-xs text-slate-300 mt-1 leading-relaxed">{notif.message}</p>
                      <span className="text-[10px] text-slate-500 mt-1.5 block font-mono">{notif.createdAt}</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="p-8 text-center space-y-2">
                  <Bell className="w-8 h-8 text-slate-600 mx-auto" />
                  <p className="text-xs font-semibold text-slate-400">No active notifications</p>
                  <p className="text-[11px] text-slate-500">Module completions and progress resets will appear here.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
