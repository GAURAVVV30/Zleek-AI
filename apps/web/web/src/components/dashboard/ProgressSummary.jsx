import React from 'react';

export default function ProgressSummary({ progress, summary }) {
  const overallProgress = summary?.overallCompletionPercentage ?? (progress !== undefined ? Number(progress) : 0);
  
  const skillsMastered = summary?.completedConcepts ?? summary?.skillsMastered ?? (
    overallProgress > 0 
      ? Math.round((overallProgress / 100) * (summary?.totalConcepts || 15)) 
      : 0
  );

  const learningHours = summary?.learningHours ?? summary?.hoursSpent ?? (
    skillsMastered > 0 
      ? Math.max(1, Math.round(skillsMastered * 1.5)) 
      : 0
  );

  const projectsCompleted = summary?.projectsCompleted ?? summary?.completedProjects ?? (
    skillsMastered > 0 
      ? Math.floor(skillsMastered / 3) 
      : 0
  );

  const stats = [
    { value: `${overallProgress}%`, label: 'Mission Progress' },
    { value: `${skillsMastered}`, label: 'Skills Mastered' },
    { value: `${learningHours}`, label: 'Learning Hours' },
    { value: `${projectsCompleted}`, label: 'Projects Completed' },
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
