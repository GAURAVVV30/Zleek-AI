import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Award, CheckCircle2, AlertTriangle, ArrowRight, Sparkles } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function BaselineResultsPage() {
  const [results, setResults] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const navigate = useNavigate();
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.DIAGNOSTIC.RESULTS('diag_sess_789'))
      .then((res) => {
        setResults(res.data);
        setIsLoading(false);
      })
      .catch(() => {
        setIsLoading(false);
      });
  }, []);

  const handleGenerateRoadmap = async () => {
    setIsGenerating(true);
    try {
      await apiClient.post(ENDPOINTS.ROADMAP.REGENERATE);
      addToast('Your personalized learning roadmap has been generated!', 'success');
      navigate('/roadmap');
    } catch (err) {
      addToast('Failed to generate roadmap', 'error');
    } finally {
      setIsGenerating(false);
    }
  };

  if (isLoading || !results) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-500">Synthesizing concept gaps & prerequisite graph...</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-12">
      <div className="bg-white border border-slate-200/80 rounded-3xl p-8 shadow-card space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-slate-100">
          <div className="flex items-center gap-2">
            <Award className="w-5 h-5 text-blue-600" />
            <span className="text-xs font-bold text-slate-900 uppercase tracking-wider">
              Diagnostic Summary & Baseline
            </span>
          </div>
          <span className="px-3 py-1 bg-blue-50 text-blue-700 text-xs font-bold rounded-full">
            Level: {results.assessedLevel}
          </span>
        </div>

        <div>
          <h1 className="font-display text-2xl font-bold text-slate-900">
            Here's where you're starting from
          </h1>
          <p className="text-xs sm:text-sm text-slate-600 mt-1 leading-relaxed">
            We verified your strong grasp of core fundamentals. The AI has tailored your roadmap to prioritize your identified skill gaps.
          </p>
        </div>

        {/* Concept Coverage Bars */}
        <div className="space-y-4 bg-slate-50/60 p-6 rounded-2xl border border-slate-100">
          <h3 className="text-xs font-bold text-slate-700 uppercase tracking-wider">
            Assessed Concept Coverage
          </h3>
          <div className="space-y-3">
            {results.conceptCoverage.map((concept) => (
              <div key={concept.conceptId} className="space-y-1.5">
                <div className="flex justify-between text-xs font-medium">
                  <span className="text-slate-800">{concept.conceptName}</span>
                  <span className="text-slate-500">{concept.coveragePercentage}%</span>
                </div>
                <div className="w-full h-2 bg-slate-200 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${
                      concept.coveragePercentage >= 80
                        ? 'bg-emerald-500'
                        : concept.coveragePercentage >= 50
                        ? 'bg-blue-600'
                        : 'bg-amber-500'
                    }`}
                    style={{ width: `${concept.coveragePercentage}%` }}
                  ></div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Identified Gaps to close */}
        <div>
          <h3 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2.5">
            Key Identified Gaps to Bridge
          </h3>
          <div className="flex flex-wrap gap-2">
            {results.topGaps.map((gap) => (
              <div
                key={gap}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-50 border border-amber-200 text-amber-800 rounded-xl text-xs font-medium"
              >
                <AlertTriangle className="w-3.5 h-3.5 text-amber-600" />
                <span>{gap}</span>
              </div>
            ))}
          </div>
        </div>

        {/* CTA */}
        <div className="pt-4 border-t border-slate-100 flex items-center justify-end">
          <button
            onClick={handleGenerateRoadmap}
            disabled={isGenerating}
            className="px-8 py-3.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-bold text-xs shadow-elevated hover:shadow-glow transition flex items-center gap-2"
          >
            <Sparkles className="w-4 h-4" />
            <span>{isGenerating ? 'Building path...' : 'Generate My Roadmap'}</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
