import React from 'react';

export default function CurrentMission({ roadmap }) {
  if (!roadmap) return null;

  return (
    <div className="relative z-10 bg-slate-900/30 backdrop-blur-xl border border-white/10 rounded-[20px] px-6 lg:px-10 py-6 shadow-[0_0_30px_rgba(79,70,229,0.15)] flex flex-col md:flex-row items-center gap-6 md:gap-8 min-h-[120px] w-full">
      {/* Left: Circular Progress Ring */}
      <div className="relative flex items-center justify-center shrink-0 w-[64px] h-[64px]">
        <svg className="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
          {/* Background Track */}
          <path
            className="text-white/10"
            strokeWidth="3"
            stroke="currentColor"
            fill="none"
            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
          />
          {/* Progress Fill */}
          <path
            className="text-indigo-400 drop-shadow-[0_0_8px_rgba(129,140,248,0.6)]"
            strokeDasharray={`${roadmap.progressPercentage}, 100`}
            strokeWidth="3"
            strokeLinecap="round"
            stroke="currentColor"
            fill="none"
            d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-[13px] font-bold text-white">{roadmap.progressPercentage}%</span>
        </div>
      </div>

      {/* Center: Mission Text */}
      <div className="flex-1 min-w-0 text-center md:text-left flex flex-col justify-center">
        <span className="text-[11px] font-medium text-slate-400 tracking-widest uppercase mb-1.5 block">
          Current Mission:
        </span>
        <h1 className="font-display text-xl lg:text-[26px] font-extrabold text-white tracking-tight truncate leading-tight">
          {(() => {
            const raw = roadmap.goalTitle || '';
            const lower = raw.toLowerCase();
            if (lower.includes('data scientist')) return 'Become a Data Scientist';
            if (lower.includes('data analyst')) return 'Become a Data Analyst';
            if (lower.includes('ml engineer') || lower.includes('machine learning engineer')) return 'Become a Machine Learning Engineer';
            
            const match = raw.match(/become an? (.*?)(?:\s+and\s+|\.|$)/i);
            if (match && match[1]) {
              const role = match[1].replace(/\b\w/g, l => l.toUpperCase());
              return `Become a ${role}`;
            }
            
            if (raw.length < 30) return `Become a ${raw}`;
            return raw;
          })()}
        </h1>
      </div>

      {/* Vertical Divider */}
      <div className="hidden md:block w-px h-14 bg-white/10 shrink-0 mx-2"></div>

      {/* Right: Time Remaining */}
      <div className="shrink-0 flex flex-col justify-center text-center md:text-left">
        <span className="text-[11px] font-medium text-slate-400 tracking-widest uppercase mb-1.5 block">
          Time Remaining:
        </span>
        <span className="text-xl lg:text-2xl font-bold text-white leading-tight">
          3 Weeks
        </span>
      </div>
    </div>
  );
}
