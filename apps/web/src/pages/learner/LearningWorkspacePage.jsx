import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Video,
  CheckCircle,
  ThumbsUp,
  ThumbsDown,
  Sparkles,
  ArrowRight,
  ExternalLink,
  BookOpen,
  ArrowLeft,
  Bot,
  Zap,
  Shield,
  HelpCircle,
} from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';
// Avatar import removed

export default function LearningWorkspacePage() {
  const { conceptId } = useParams();
  const [concept, setConcept] = useState(null);
  const [isReviewed, setIsReviewed] = useState(false);
  const [showAlternate, setShowAlternate] = useState(false);
  const [feedbackSent, setFeedbackSent] = useState(false);
  const [companionTipOpen, setCompanionTipOpen] = useState(true);
  // Companion removed
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.CONCEPTS.DETAIL(conceptId || 'c_pandas'))
      .then((res) => {
        setConcept(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, [conceptId]);

  const handleMarkReviewed = async () => {
    try {
      await apiClient.post(ENDPOINTS.CONCEPTS.ENGAGEMENT(conceptId || 'c_pandas'), {
        action: 'marked_reviewed',
      });
      setIsReviewed(true);
      addToast('Engagement verified! Assessment gate unlocked.', 'success');
    } catch (err) {
      addToast('Could not record engagement. Please try again.', 'error');
    }
  };

  const handleFeedback = async (rating) => {
    try {
      await apiClient.post(ENDPOINTS.RESOURCES.FEEDBACK(concept?.primaryResource?.id || 'res_1'), {
        rating,
      });
      setFeedbackSent(true);
      addToast('Feedback recorded to improve curator ranking algorithm.', 'info');
    } catch (err) {
      setFeedbackSent(true);
    }
  };

  if (isLoading || !concept) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading learning workspace...</p>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Breadcrumbs & Navigation */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-400">
          <button onClick={() => navigate('/roadmap')} className="hover:text-indigo-400 flex items-center gap-1">
            <ArrowLeft className="w-3.5 h-3.5" /> Roadmap
          </button>
          <span>/</span>
          <span className="text-white">{concept.breadcrumb?.join(' > ')}</span>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-xs font-mono font-bold text-slate-400">Assistant:</span>
          <span className="px-2.5 py-1 bg-slate-900 text-cyan-300 font-mono text-[11px] rounded-lg border border-slate-700 flex items-center gap-1.5">
            <Bot className="w-3.5 h-3.5 text-cyan-400" />
            AI Study Guide
          </span>
        </div>
      </div>

      {/* AI Assistant Banner */}
      {companionTipOpen && (
        <div className="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 border border-slate-700 rounded-3xl p-5 text-white shadow-lg space-y-2 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-cyan-400 font-mono font-bold text-xs">
              <Zap className="w-4 h-4" />
              <span>AI LEARNING ASSISTANT</span>
            </div>
            <button
              onClick={() => setCompanionTipOpen(false)}
              className="text-[11px] text-slate-400 hover:text-white transition"
            >
              Dismiss
            </button>
          </div>
          <p className="text-xs text-slate-200 leading-relaxed font-medium">
            💡 <strong>Study Tip:</strong> "You're making great progress! Review the examples carefully before taking the quiz."
          </p>
        </div>
      )}

      {/* Main Workspace Card */}
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
        {/* Concept Title & Why it matters */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold text-indigo-400 uppercase tracking-wider">
              Core Concept Node
            </span>
          </div>
          <h1 className="font-display text-2xl font-extrabold text-white">
            {concept.title}
          </h1>
          <p className="text-xs sm:text-sm text-slate-300 leading-relaxed max-w-3xl">
            {concept.whyItMatters}
          </p>
        </div>

        {/* Primary Trusted Resource Card */}
        <div className="bg-black/30 backdrop-blur-md border border-white/10 rounded-2xl p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-bold text-white">
              <Video className="w-4 h-4 text-indigo-400" />
              Primary Gold-Standard Resource
            </div>
            <span className="text-[11px] font-semibold text-slate-400">
              {concept.primaryResource?.durationMinutes} min runtime
            </span>
          </div>

          <div className="bg-black/40 backdrop-blur-xl rounded-xl p-4 border border-white/10 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div>
              <h3 className="text-sm font-bold text-white">
                {concept.primaryResource?.title}
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Provider: {concept.primaryResource?.provider} · Vetted by {concept.primaryResource?.provenance?.vettedBy}
              </p>
            </div>
            <a
              href="https://www.youtube.com/watch?v=dcqPhpY7tWk"
              target="_blank"
              rel="noreferrer"
              className="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-semibold flex items-center gap-1.5 shrink-0 transition"
            >
              <span>Open Resource</span>
              <ExternalLink className="w-3.5 h-3.5" />
            </a>
          </div>

          {/* Explainability Framing */}
          <div className="p-3.5 bg-indigo-900/40 backdrop-blur-sm/60 border border-blue-100 rounded-xl text-xs text-blue-900 flex items-start gap-2.5">
            <Sparkles className="w-4 h-4 text-indigo-400 shrink-0 mt-0.5" />
            <p className="leading-relaxed text-[11px]">
              <strong>Why this resource?</strong> {concept.primaryResource?.whyThisResource}
            </p>
          </div>

          {/* Engagement Review & Feedback */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between pt-2 gap-4 border-t border-white/10">
            <button
              onClick={handleMarkReviewed}
              className={`px-5 py-2.5 rounded-xl text-xs font-bold transition flex items-center gap-2 ${
                isReviewed
                  ? 'bg-emerald-50 text-emerald-700 border border-emerald-200 cursor-default'
                  : 'bg-indigo-600 hover:bg-indigo-700 text-white shadow-[0_0_10px_rgba(79,70,229,0.1)]'
              }`}
            >
              <CheckCircle className="w-4 h-4" />
              <span>{isReviewed ? 'Engagement Logged' : 'Mark as Studied / Reviewed'}</span>
            </button>

            {/* Micro Feedback */}
            <div className="flex items-center gap-3 text-xs text-slate-400">
              <span>Was this resource helpful?</span>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => handleFeedback('up')}
                  disabled={feedbackSent}
                  className={`p-2 rounded-lg border hover:bg-slate-100 transition ${
                    feedbackSent ? 'opacity-50 cursor-default' : 'hover:text-indigo-400'
                  }`}
                >
                  <ThumbsUp className="w-3.5 h-3.5" />
                </button>
                <button
                  onClick={() => handleFeedback('down')}
                  disabled={feedbackSent}
                  className={`p-2 rounded-lg border hover:bg-slate-100 transition ${
                    feedbackSent ? 'opacity-50 cursor-default' : 'hover:text-red-600'
                  }`}
                >
                  <ThumbsDown className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Alternate Formats (Articles / Interactive Labs) */}
        <div>
          <button
            onClick={() => setShowAlternate(!showAlternate)}
            className="text-xs font-semibold text-indigo-400 hover:underline flex items-center gap-1"
          >
            <BookOpen className="w-3.5 h-3.5" />
            <span>{showAlternate ? 'Hide Alternate Formats' : 'Prefer an article or interactive lab instead?'}</span>
          </button>

          {showAlternate && (
            <div className="mt-3 p-4 bg-black/30 backdrop-blur-md rounded-2xl border border-white/10 space-y-2 animate-in fade-in duration-200">
              <h4 className="text-xs font-bold text-white">Alternative Vetted Materials:</h4>
              <ul className="space-y-1.5 text-xs text-slate-300">
                <li className="flex items-center justify-between p-2 rounded-lg bg-black/40 backdrop-blur-xl border border-white/5">
                  <span>Interactive Tutorial: 10 Minutes to Pandas (Official Docs)</span>
                  <a
                    href="https://pandas.pydata.org/docs/user_guide/10min.html"
                    target="_blank"
                    rel="noreferrer"
                    className="text-indigo-400 font-semibold flex items-center gap-1"
                  >
                    Read <ExternalLink className="w-3 h-3" />
                  </a>
                </li>
              </ul>
            </div>
          )}
        </div>

        {/* Assessment CTA Gate */}
        <div className="pt-4 flex items-center justify-between border-t border-white/5">
          <button
            type="button"
            onClick={() => navigate('/roadmap')}
            className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white"
          >
            ← Back to Roadmap
          </button>

          <button
            onClick={() => navigate(`/assessment/${conceptId || 'c_pandas'}`)}
            disabled={!isReviewed}
            className={`px-8 py-3.5 rounded-xl font-bold text-xs shadow-elevated transition flex items-center gap-2 ${
              isReviewed
                ? 'bg-emerald-600 hover:bg-emerald-700 text-white shadow-emerald-600/20'
                : 'bg-slate-200 text-slate-400 cursor-not-allowed'
            }`}
          >
            <span>Take Evidence Assessment</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
