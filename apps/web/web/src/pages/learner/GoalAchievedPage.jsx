import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Trophy, Lock, ArrowRight, Sparkles, CheckCircle2, RotateCcw } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function GoalAchievedPage() {
  const [badgeData, setBadgeData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    setIsLoading(true);
    apiClient
      .get(ENDPOINTS.PROGRESS.COMPLETION_BADGE)
      .then((res) => {
        const payload = res?.data || res;
        setBadgeData(payload);
        setIsLoading(false);
      })
      .catch((err) => {
        const errorPayload = err?.response?.data?.data || err?.response?.data || {
          eligible: false,
          role: 'Role Track',
          completedModules: 0,
          totalModules: 1,
          message: 'Complete all roadmap modules in your active role to unlock the official Role Completion Badge.',
        };
        setBadgeData(errorPayload);
        setIsLoading(false);
      });
  }, []);

  if (isLoading) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Verifying authoritative backend badge completion status...</p>
      </div>
    );
  }

  const isEligible = badgeData?.eligible === true;
  const roleName = badgeData?.role || badgeData?.badge?.title || 'LEARNER';
  const completedCount = badgeData?.completedModules || 0;
  const totalCount = badgeData?.totalModules || 1;
  const progressPercent = Math.round((completedCount / totalCount) * 100);

  return (
    <div className="max-w-3xl mx-auto py-8 px-4 space-y-8">
      {isEligible ? (
        /* Authorized Badge View */
        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 sm:p-12 shadow-[0_0_40px_rgba(79,70,229,0.2)] text-center space-y-8 animate-in zoom-in-95 duration-300">
          <div className="flex items-center justify-center gap-2 text-indigo-400 font-bold text-xs uppercase tracking-widest bg-indigo-950/60 border border-indigo-500/20 px-4 py-1.5 rounded-full w-fit mx-auto">
            <Sparkles className="w-4 h-4 text-indigo-400" />
            Verified Role Competency Badge
          </div>

          <div>
            <h1 className="font-display text-3xl sm:text-4xl font-extrabold text-white tracking-tight">
              Congratulations!
            </h1>
            <p className="text-sm font-semibold text-slate-300 mt-2">
              You have completed all {totalCount} roadmap modules and earned the verified role badge for:
            </p>
          </div>

          {/* Premium Blue & Silver Shield Badge Container */}
          <div className="relative w-[320px] sm:w-[380px] h-[360px] sm:h-[420px] mx-auto flex items-center justify-center transition-transform hover:scale-105 duration-500 group">
            {/* Background Radial Glow */}
            <div className="absolute inset-0 bg-indigo-500/25 blur-[60px] rounded-full pointer-events-none group-hover:bg-indigo-400/35 transition-all"></div>

            {/* Official Shield Artwork */}
            <img
              src="/assets/badge_shield.jpg"
              alt={`${roleName} Shield Badge`}
              className="relative w-full h-full object-contain pointer-events-none drop-shadow-[0_15px_30px_rgba(0,0,0,0.8)]"
            />

            {/* Dynamic Runtime Role Overlay inside Shield Frame */}
            <div className="absolute inset-0 flex items-center justify-center px-10 pt-4 pb-8 z-10 pointer-events-none">
              <div className="max-w-[210px] sm:max-w-[240px] text-center">
                <span
                  className="font-display font-extrabold text-white text-base sm:text-lg lg:text-xl tracking-wider leading-snug uppercase drop-shadow-[0_4px_12px_rgba(0,0,0,0.9)] bg-gradient-to-b from-white via-slate-100 to-indigo-200 bg-clip-text text-transparent"
                  style={{ letterSpacing: '0.08em' }}
                >
                  {roleName}
                </span>
              </div>
            </div>
          </div>

          {/* Action Navigation */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
            <button
              onClick={() => navigate('/progress')}
              className="w-full sm:w-auto px-8 py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold text-xs shadow-lg transition flex items-center justify-center gap-2"
            >
              <span>View Progress Dashboard</span>
              <ArrowRight className="w-4 h-4" />
            </button>
            <button
              onClick={() => navigate('/roadmap')}
              className="w-full sm:w-auto px-6 py-3.5 bg-slate-900 border border-slate-700 hover:border-slate-500 text-slate-300 hover:text-white rounded-xl font-semibold text-xs transition"
            >
              Back to Roadmap
            </button>
          </div>
        </div>
      ) : (
        /* Incomplete / Locked Badge State */
        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 sm:p-12 shadow-[0_0_30px_rgba(225,29,72,0.15)] text-center space-y-8">
          <div className="w-20 h-20 rounded-3xl bg-rose-500/10 border border-rose-500/30 text-rose-400 flex items-center justify-center mx-auto shadow-inner">
            <Lock className="w-10 h-10" />
          </div>

          <div className="space-y-2">
            <span className="text-xs font-bold text-rose-400 uppercase tracking-widest block">
              Badge Access Restricted
            </span>
            <h2 className="font-display text-2xl sm:text-3xl font-extrabold text-white">
              {roleName} Badge Locked
            </h2>
            <p className="text-xs sm:text-sm text-slate-300 max-w-md mx-auto leading-relaxed">
              {badgeData?.message || `All ${totalCount} roadmap modules in your active role must be completed to unlock the official Role Completion Badge.`}
            </p>
          </div>

          {/* Progress Tracker Card */}
          <div className="max-w-md mx-auto bg-black/40 backdrop-blur-md border border-white/10 rounded-2xl p-6 text-left space-y-3">
            <div className="flex items-center justify-between text-xs font-semibold">
              <span className="text-slate-300">Roadmap Completion</span>
              <span className="text-rose-400 font-bold">{completedCount} / {totalCount} Modules ({progressPercent}%)</span>
            </div>
            <div className="w-full h-2.5 bg-slate-900 rounded-full overflow-hidden border border-white/5">
              <div
                className="h-full bg-gradient-to-r from-rose-500 to-amber-500 rounded-full transition-all duration-500"
                style={{ width: `${progressPercent}%` }}
              ></div>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <button
              onClick={() => navigate('/roadmap')}
              className="w-full sm:w-auto px-8 py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold text-xs shadow-md transition flex items-center justify-center gap-2"
            >
              <span>Return to Roadmap</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
