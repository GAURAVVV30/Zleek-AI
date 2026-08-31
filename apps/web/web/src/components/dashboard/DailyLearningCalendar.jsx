import React from 'react';
import { format, isSameDay } from 'date-fns';
import { Check } from 'lucide-react';

export default function DailyLearningCalendar({ weekData, selectedDate, onSelectDate }) {
  // weekData is an array of objects: { date: Date, tasks: [] }
  return (
    <div className="flex flex-col gap-4 w-full">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-bold text-white uppercase tracking-widest">Daily Learning Plan</h2>
      </div>
      <div className="flex overflow-x-auto snap-x snap-mandatory gap-3 lg:gap-4 pb-2 [&::-webkit-scrollbar]:hidden w-full items-stretch">
        {weekData.map((dayData, index) => {
          const isSelected = isSameDay(dayData.date, selectedDate);
          const totalTasks = dayData.tasks.length;
          const completedTasks = dayData.tasks.filter(t => t.completed).length;
          const isAllCompleted = totalTasks > 0 && completedTasks === totalTasks;

          return (
            <div
              key={index}
              onClick={() => onSelectDate(dayData.date)}
              className={`relative snap-center shrink-0 flex-1 min-w-[90px] rounded-2xl flex flex-col items-center justify-center py-4 px-2 cursor-pointer transition-all border backdrop-blur-md ${
                isSelected
                  ? 'bg-indigo-900/40 border-indigo-400/50 shadow-[0_0_20px_rgba(79,70,229,0.25)] ring-1 ring-indigo-500/30'
                  : 'bg-black/40 border-white/10 hover:bg-slate-800/50 hover:border-white/20'
              }`}
            >
              <span className={`text-[10px] lg:text-xs font-bold uppercase tracking-widest mb-1 ${isSelected ? 'text-indigo-300' : 'text-slate-400'}`}>
                {format(dayData.date, 'EEE')}
              </span>
              <span className={`text-sm lg:text-base font-bold mb-3 ${isSelected ? 'text-white' : 'text-slate-300'}`}>
                {format(dayData.date, 'MMM d')}
              </span>
              
              {/* Progress fraction */}
              <div className={`flex items-center justify-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full ${
                isAllCompleted 
                  ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' 
                  : (isSelected ? 'bg-indigo-500/20 text-indigo-200 border border-indigo-500/30' : 'bg-white/5 text-slate-400 border border-white/10')
              }`}>
                {isAllCompleted && <Check className="w-3 h-3" />}
                <span>
                  {totalTasks > 0 ? `${completedTasks}/${totalTasks}` : '-'}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
