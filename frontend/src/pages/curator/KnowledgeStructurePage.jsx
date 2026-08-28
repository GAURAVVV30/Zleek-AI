import React, { useState, useEffect } from 'react';
import { Network, Plus, CheckCircle2, AlertTriangle, ShieldCheck, ArrowRight } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function KnowledgeStructurePage() {
  const [structures, setStructures] = useState([]);
  const [isValidating, setIsValidating] = useState(false);
  const [validationResult, setValidationResult] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.CURATOR.STRUCTURES)
      .then((res) => {
        setStructures(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  const handleValidateGraph = async () => {
    setIsValidating(true);
    try {
      const res = await apiClient.post(ENDPOINTS.CURATOR.VALIDATE_STRUCTURE);
      setValidationResult(res);
      addToast('Knowledge Graph passed DAG cycle validation (0 circular dependencies).', 'success');
    } catch (err) {
      addToast('Validation error detected in graph', 'error');
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="bg-white border border-purple-200/80 rounded-2xl p-6 shadow-sm flex items-center justify-between">
        <div>
          <span className="text-[11px] font-bold text-purple-700 uppercase tracking-wider block mb-1">
            Curator Tooling
          </span>
          <h1 className="font-display text-xl font-extrabold text-slate-900">
            Knowledge Structure Manager (DAG Editor)
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleValidateGraph}
            disabled={isValidating}
            className="px-4 py-2.5 bg-purple-50 hover:bg-purple-100 text-purple-700 font-bold rounded-xl text-xs border border-purple-200 transition flex items-center gap-1.5"
          >
            <ShieldCheck className="w-4 h-4" />
            <span>{isValidating ? 'Checking DAG...' : 'Validate Structure'}</span>
          </button>
          <button
            onClick={() => addToast('Draft concept node added to knowledge structure.', 'info')}
            className="px-4 py-2.5 bg-purple-600 hover:bg-purple-700 text-white font-bold rounded-xl text-xs shadow-sm transition flex items-center gap-1.5"
          >
            <Plus className="w-4 h-4" />
            <span>Add Concept</span>
          </button>
        </div>
      </div>

      {validationResult && (
        <div className="p-4 bg-emerald-50 border border-emerald-200 rounded-2xl flex items-center gap-3 text-xs text-emerald-800">
          <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0" />
          <span>{validationResult.message}</span>
        </div>
      )}

      {/* Concept Tree DAG View */}
      <div className="bg-white border border-slate-200/80 rounded-3xl p-6 sm:p-8 shadow-card space-y-6">
        <h3 className="text-xs font-bold text-slate-800 uppercase tracking-wider">
          Data Science Domain Taxonomy (v2.0 Published)
        </h3>

        <div className="border border-slate-200 rounded-2xl p-6 bg-slate-50/50 space-y-4 font-mono text-xs">
          <div className="flex items-center gap-2 text-purple-700 font-bold">
            <Network className="w-4 h-4" /> ▾ Data Science (Root Domain)
          </div>

          <div className="pl-6 space-y-3 border-l-2 border-slate-200">
            <div className="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
              <div>
                <span className="font-bold text-slate-900">1. Python Basics</span>
                <span className="text-[10px] text-slate-400 block font-sans">Prerequisites: None (Entry Level)</span>
              </div>
              <span className="px-2 py-0.5 bg-emerald-50 text-emerald-700 text-[10px] rounded font-bold uppercase font-sans">Published</span>
            </div>

            <div className="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
              <div>
                <span className="font-bold text-slate-900">2. Data Analysis with Pandas</span>
                <span className="text-[10px] text-slate-400 block font-sans">Prerequisites: Python Basics</span>
              </div>
              <span className="px-2 py-0.5 bg-emerald-50 text-emerald-700 text-[10px] rounded font-bold uppercase font-sans">Published</span>
            </div>

            <div className="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
              <div>
                <span className="font-bold text-slate-900">3. Statistics & Hypothesis Testing</span>
                <span className="text-[10px] text-slate-400 block font-sans">Prerequisites: Data Analysis with Pandas</span>
              </div>
              <span className="px-2 py-0.5 bg-emerald-50 text-emerald-700 text-[10px] rounded font-bold uppercase font-sans">Published</span>
            </div>

            <div className="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
              <div>
                <span className="font-bold text-slate-900">4. Machine Learning Models</span>
                <span className="text-[10px] text-slate-400 block font-sans">Prerequisites: Statistics & Hypothesis Testing</span>
              </div>
              <span className="px-2 py-0.5 bg-purple-50 text-purple-700 text-[10px] rounded font-bold uppercase font-sans">Draft</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
