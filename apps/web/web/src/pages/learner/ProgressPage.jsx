import React, { useState, useEffect } from 'react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { addDays, format } from 'date-fns';
import LearningActivityHeatmap from '../../components/ui/learning-activity-heatmap';
import ProgressSummary from '../../components/dashboard/ProgressSummary';

export default function ProgressPage() {
  const [summary, setSummary] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      apiClient.get(ENDPOINTS.PROGRESS.SUMMARY),
      apiClient.get(ENDPOINTS.ROADMAP.BASE).catch(() => ({ data: null })),
    ])
      .then(([sumRes, roadmapRes]) => {
        const rawSum = sumRes?.data || sumRes;
        const sumData = rawSum?.data || rawSum;

        const rawRoadmap = roadmapRes?.data || roadmapRes;
        const roadmapData = rawRoadmap?.data || rawRoadmap;

        let finalSummary = sumData;

        // Sync progress breakdown directly with active roadmap milestones
        if (roadmapData && Array.isArray(roadmapData.nodes) && roadmapData.nodes.length > 0) {
          const rawRole = roadmapData?.domain || roadmapData?.domain_id || localStorage.getItem('userActiveRole');
          let activeRole = rawRole ? rawRole.trim().toLowerCase() : 'full_stack';

          let savedCompleted = [];
          try {
            const raw = localStorage.getItem(`gold_completed_modules_${activeRole}`);
            if (raw) savedCompleted = JSON.parse(raw);
          } catch (e) {}

          const totalConcepts = roadmapData.nodes.length;
          let completedConcepts = 0;
          let prevAllCompleted = true;

          const competencyBreakdown = roadmapData.nodes.map((node, index) => {
            let isCompleted = false;
            if (prevAllCompleted) {
              isCompleted = 
                node.state === 'competent' || 
                savedCompleted.includes(node.id) ||
                savedCompleted.includes(`${index + 1}`) ||
                savedCompleted.some(saved => 
                  typeof saved === 'string' && node.title && 
                  (saved.toLowerCase().includes(node.title.toLowerCase()) || node.title.toLowerCase().includes(saved.toLowerCase()))
                );
            }

            if (isCompleted) {
              completedConcepts++;
              prevAllCompleted = true;
            } else {
              prevAllCompleted = false;
            }

            return {
              domain: node.title,
              percentage: isCompleted ? 100 : 0,
              status: isCompleted ? 'Competent' : (index > 0 && !prevAllCompleted && index === completedConcepts ? 'Available' : 'Not Started'),
            };
          });

          const calculatedPercentage = Math.round((completedConcepts / totalConcepts) * 100);

          finalSummary = {
            ...sumData,
            totalConcepts,
            completedConcepts,
            overallCompletionPercentage: calculatedPercentage,
            competencyBreakdown,
          };
        }

        setSummary(finalSummary);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  // Use actual 365-day activity matrix returned from the backend API (past 365 days leading up to today)
  const activityData = React.useMemo(() => {
    if (summary?.activityData && summary.activityData.length > 0) {
      return summary.activityData;
    }
    const data = [];
    const today = new Date();
    for (let i = 0; i < 365; i++) {
      const date = addDays(today, -364 + i);
      data.push({
        date: format(date, 'yyyy-MM-dd'),
        count: 0,
      });
    }
    return data;
  }, [summary]);

  if (isLoading || !summary) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading progress overview...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-[0_0_10px_rgba(79,70,229,0.1)] flex items-center justify-between">
        <div>
          <h1 className="font-display text-xl font-extrabold text-white">
            Progress Overview Dashboard
          </h1>
          <p className="text-xs text-slate-400 mt-0.5">
            Real-time track of your roadmap completion & activity performance
          </p>
        </div>
      </div>

      <ProgressSummary progress={summary.overallCompletionPercentage} summary={summary} />

      {/* Progress View: Radial Chart & Mastery Breakdowns */}
      <div className="grid grid-cols-1 md:grid-cols-12 gap-6">
        {/* Radial Gauge */}
        <div className="md:col-span-5 bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] text-center flex flex-col items-center justify-center space-y-4">
          <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
            Overall Goal Competency
          </span>

          <div className="relative w-40 h-40 flex items-center justify-center">
            <svg className="w-full h-full -rotate-90" viewBox="0 0 36 36">
              <path
                className="text-slate-800"
                strokeWidth="3.5"
                stroke="currentColor"
                fill="none"
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              />
              <path
                className="text-emerald-500 transition-all duration-1000 ease-out"
                strokeDasharray={`${summary.overallCompletionPercentage}, 100`}
                strokeWidth="3.5"
                strokeLinecap="round"
                stroke="currentColor"
                fill="none"
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              />
            </svg>
            <div className="absolute flex flex-col items-center">
              <span className="font-display text-4xl font-extrabold text-white">
                {summary.overallCompletionPercentage}%
              </span>
              <span className="text-[10px] font-bold text-emerald-400 uppercase">Complete</span>
            </div>
          </div>

          <p className="text-xs text-slate-400">
            {summary.completedConcepts} of {summary.totalConcepts} milestones completed.
          </p>
        </div>

        {/* Breakdown Bars */}
        <div className="md:col-span-7 bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-4">
          <h3 className="text-xs font-bold text-white uppercase tracking-wider">
            Concept Mastery Levels
          </h3>
          <div className="space-y-4 pt-2">
            {summary.competencyBreakdown?.map((item, idx) => (
              <div key={idx} className="space-y-1.5">
                <div className="flex justify-between text-xs font-semibold">
                  <span className="text-white">{item.domain}</span>
                  <span className="text-slate-400">{item.percentage}%</span>
                </div>
                <div className="w-full h-2 bg-slate-900 rounded-full overflow-hidden border border-white/5">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${
                      item.percentage >= 70
                        ? 'bg-emerald-500'
                        : item.percentage >= 40
                        ? 'bg-indigo-600'
                        : 'bg-slate-700'
                    }`}
                    style={{ width: `${item.percentage}%` }}
                  ></div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Learning Activity Heatmap */}
        <div className="md:col-span-12">
          <LearningActivityHeatmap data={activityData} />
        </div>
      </div>
    </div>
  );
}
