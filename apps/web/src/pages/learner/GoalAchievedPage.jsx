import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Trophy, CheckCircle2, ArrowRight, Sparkles, Share2 } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function GoalAchievedPage() {
  const [summary, setSummary] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.GOALS.COMPLETION_SUMMARY)
      .then((res) => {
        setSummary(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  if (isLoading || !summary) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading goal completion record...</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto py-8">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 sm:p-12 shadow-[0_0_20px_rgba(79,70,229,0.15)] text-center space-y-8 animate-in zoom-in-95 duration-300">
        {/* Trophy visual */}
        <div className="relative inline-block">
          <div className="w-24 h-24 rounded-3xl bg-gradient-to-tr from-amber-400 to-amber-500 text-white flex items-center justify-center mx-auto shadow-lg shadow-amber-500/25">
            <Trophy className="w-12 h-12" />
          </div>
          <Sparkles className="w-6 h-6 text-amber-500 absolute -top-2 -right-2 animate-bounce" />
        </div>

        <div>
          <h1 className="font-display text-3xl font-extrabold text-white">
            Congratulations!
          </h1>
          <p className="text-sm font-semibold text-indigo-400 mt-1">
            You have verified competence for:
          </p>
          <p className="font-display text-lg font-bold text-white mt-0.5">
            {summary.goalTitle}
          </p>
        </div>

        {/* Proof list */}
        <div className="bg-black/30 backdrop-blur-md rounded-2xl p-6 border border-white/10 text-left space-y-3">
          <h3 className="text-xs font-bold text-white uppercase tracking-wider">
            Verified Competency Portfolio ({summary.totalSkillsVerified} Milestones)
          </h3>
          <ul className="space-y-2 text-xs text-white">
            {summary.masteryProofList?.map((item, idx) => (
              <li key={idx} className="flex items-center gap-2.5">
                <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <button
            onClick={() => navigate('/onboarding/goal')}
            className="w-full sm:w-auto px-8 py-3.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-xs shadow-elevated transition flex items-center justify-center gap-2"
          >
            <span>Define a New Goal</span>
            <ArrowRight className="w-4 h-4" />
          </button>
          <button
            onClick={() => navigate('/progress')}
            className="w-full sm:w-auto px-6 py-3.5 bg-black/40 backdrop-blur-xl hover:bg-black/30 backdrop-blur-md text-white border border-white/10 rounded-xl font-semibold text-xs shadow-[0_0_10px_rgba(79,70,229,0.1)] transition"
          >
            View Evidence Portfolio
          </button>
        </div>
      </div>
    </div>
  );
}
