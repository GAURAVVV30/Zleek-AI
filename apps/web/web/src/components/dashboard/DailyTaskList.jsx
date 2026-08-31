import React from 'react';
import { CheckCircle2, Circle, ArrowRight, PlayCircle } from 'lucide-react';
import { format, isToday } from 'date-fns';

export default function DailyTaskList({ selectedDate, tasks, onToggleTask }) {
  const completedTasks = tasks.filter(t => t.completed).length;
  const totalTasks = tasks.length;
  const progressPercent = totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0;
  
  const remainingMinutes = tasks
    .filter(t => !t.completed)
    .reduce((acc, curr) => acc + curr.duration, 0);

  const isTodayDate = isToday(selectedDate);
  const dayName = isTodayDate ? "TODAY" : format(selectedDate, 'EEEE').toUpperCase();

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Today's Mission Card */}
      <div className="bg-gradient-to-br from-indigo-900/40 to-slate-900/50 backdrop-blur-xl border border-indigo-500/20 rounded-[20px] p-6 shadow-[0_0_30px_rgba(79,70,229,0.1)] relative overflow-hidden">
        {/* Glow effect */}
        <div className="absolute top-0 right-0 -mr-16 -mt-16 w-32 h-32 bg-indigo-500/30 blur-[50px] rounded-full pointer-events-none"></div>
        
        <div className="flex flex-col md:flex-row gap-6 items-start md:items-center justify-between relative z-10">
          <div className="space-y-2">
            <span className="text-[10px] font-bold text-indigo-400 uppercase tracking-widest block">
              {dayName}'S MISSION
            </span>
            <h2 className="text-xl lg:text-2xl font-display font-extrabold text-white">
              Python Basics
            </h2>
            <p className="text-xs text-slate-300 max-w-sm leading-relaxed">
              Build a strong foundation in Python before moving into data analysis.
            </p>
          </div>

          <div className="flex flex-col gap-3 min-w-[200px] w-full md:w-auto">
            <div className="flex items-center justify-between text-xs font-semibold">
              <span className="text-slate-300">Progress</span>
              <span className="text-indigo-300">{completedTasks} / {totalTasks} tasks</span>
            </div>
            {/* Progress Bar */}
            <div className="w-full h-2 bg-black/40 rounded-full overflow-hidden border border-white/5">
              <div 
                className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 rounded-full transition-all duration-500"
                style={{ width: `${progressPercent}%` }}
              ></div>
            </div>
            
            <div className="flex items-center justify-between text-xs mt-1">
              <span className="text-slate-400">Estimated remaining:</span>
              <span className="text-white font-bold">{remainingMinutes} min</span>
            </div>
          </div>
        </div>
      </div>

      {/* Task List */}
      <div className="flex flex-col gap-3">
        {tasks.map(task => (
          <div 
            key={task.id}
            className={`group flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl border backdrop-blur-md transition-all ${
              task.completed 
                ? 'bg-emerald-900/10 border-emerald-500/20' 
                : 'bg-black/40 border-white/10 hover:border-white/20 hover:bg-slate-900/60'
            }`}
          >
            <div className="flex items-start gap-4">
              <button 
                onClick={() => onToggleTask(task.id)}
                className="shrink-0 mt-0.5 focus:outline-none"
              >
                {task.completed ? (
                  <CheckCircle2 className="w-5 h-5 text-emerald-400 transition-transform hover:scale-110" />
                ) : (
                  <Circle className="w-5 h-5 text-slate-500 hover:text-indigo-400 transition-colors" />
                )}
              </button>
              
              <div className="flex flex-col">
                <span className={`text-sm font-semibold transition-colors ${task.completed ? 'text-slate-300 line-through decoration-slate-500/50' : 'text-white group-hover:text-indigo-200'}`}>
                  {task.title}
                </span>
                <span className="text-xs text-slate-400 mt-1">
                  {task.category} &nbsp;·&nbsp; {task.duration} min
                </span>
              </div>
            </div>

            {!task.completed && (
              <button className="shrink-0 flex items-center gap-2 text-xs font-bold text-indigo-400 hover:text-white transition-colors bg-indigo-500/10 hover:bg-indigo-500/20 px-3 py-1.5 rounded-lg border border-indigo-500/20 w-fit sm:w-auto">
                <PlayCircle className="w-4 h-4" />
                Start
              </button>
            )}
          </div>
        ))}
        {tasks.length === 0 && (
          <div className="text-center py-8 text-sm text-slate-500">
            No tasks scheduled for this day.
          </div>
        )}
      </div>
      
      {/* Primary CTA */}
      {tasks.length > 0 && progressPercent < 100 && (
        <button className="w-full sm:w-auto self-center mt-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold py-3 px-8 rounded-xl flex items-center gap-2 transition-colors shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)]">
          Continue Learning
          <ArrowRight className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}
