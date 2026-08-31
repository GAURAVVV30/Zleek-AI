import React from 'react';

export default function ProgressSummary({ progress = 64 }) {
  const stats = [
    { value: `${progress}%`, label: 'Mission Progress' },
    { value: '12', label: 'Skills Mastered' },
    { value: '8', label: 'Learning Hours' },
    { value: '4', label: 'Projects Completed' },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 w-full">
      {stats.map((stat, i) => (
        <div key={i} className="bg-black/20 backdrop-blur-xl border border-white/5 rounded-2xl p-4 lg:p-6 flex flex-col items-center justify-center text-center shadow-[0_0_15px_rgba(79,70,229,0.05)]">
          <span className="text-2xl lg:text-3xl font-display font-extrabold text-white mb-1.5">{stat.value}</span>
          <span className="text-[10px] uppercase tracking-widest text-slate-400 font-semibold">{stat.label}</span>
        </div>
      ))}
    </div>
  );
}
