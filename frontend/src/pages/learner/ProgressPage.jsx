import React, { useState, useEffect } from 'react';
import { BarChart3, Award, CheckCircle2, AlertTriangle, Clock, ChevronRight } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function ProgressPage() {
  const [activeTab, setActiveTab] = useState('progress');
  const [summary, setSummary] = useState(null);
  const [competencyDetail, setCompetencyDetail] = useState([]);
  const [historyModal, setHistoryModal] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      apiClient.get(ENDPOINTS.PROGRESS.SUMMARY),
      apiClient.get(ENDPOINTS.COMPETENCY.DETAIL),
    ])
      .then(([sumRes, compRes]) => {
        setSummary(sumRes.data);
        setCompetencyDetail(compRes.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  const handleDrillIn = async (conceptId, conceptName) => {
    try {
      const res = await apiClient.get(ENDPOINTS.COMPETENCY.HISTORY(conceptId));
      setHistoryModal({ conceptName, history: res.data });
    } catch (err) {
      console.error(err);
    }
  };

  if (isLoading || !summary) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-500">Loading verified competency records...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header & Tabs */}
      <div className="bg-white border border-slate-200/80 rounded-2xl p-6 shadow-sm flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-xl font-extrabold text-slate-900">
            Progress & Competency Dashboard
          </h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Evidence-grounded breakdown of what you can actually do
          </p>
        </div>

        <div className="flex bg-slate-100 p-1 rounded-xl text-xs font-semibold">
          <button
            onClick={() => setActiveTab('progress')}
            className={`px-4 py-2 rounded-lg transition ${
              activeTab === 'progress' ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-600'
            }`}
          >
            Overview Progress
          </button>
          <button
            onClick={() => setActiveTab('detail')}
            className={`px-4 py-2 rounded-lg transition ${
              activeTab === 'detail' ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-600'
            }`}
          >
            Competency Detail
          </button>
        </div>
      </div>

      {activeTab === 'progress' ? (
        /* Progress Tab: Radial Chart & Mastery Breakdowns */
        <div className="grid grid-cols-1 md:grid-cols-12 gap-6">
          {/* Radial Gauge */}
          <div className="md:col-span-5 bg-white border border-slate-200/80 rounded-3xl p-8 shadow-card text-center flex flex-col items-center justify-center space-y-4">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Overall Goal Competency
            </span>

            <div className="relative w-40 h-40 flex items-center justify-center">
              <svg className="w-full h-full -rotate-90" viewBox="0 0 36 36">
                <path
                  className="text-slate-100"
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
                <span className="font-display text-4xl font-extrabold text-slate-900">
                  {summary.overallCompletionPercentage}%
                </span>
                <span className="text-[10px] font-bold text-emerald-600 uppercase">Complete</span>
              </div>
            </div>

            <p className="text-xs text-slate-500">
              {summary.completedConcepts} of {summary.totalConcepts} milestones evidence-verified.
            </p>
          </div>

          {/* Breakdown Bars */}
          <div className="md:col-span-7 bg-white border border-slate-200/80 rounded-3xl p-6 sm:p-8 shadow-card space-y-4">
            <h3 className="text-xs font-bold text-slate-800 uppercase tracking-wider">
              Concept Mastery Levels
            </h3>
            <div className="space-y-4 pt-2">
              {summary.competencyBreakdown?.map((item, idx) => (
                <div key={idx} className="space-y-1.5">
                  <div className="flex justify-between text-xs font-semibold">
                    <span className="text-slate-800">{item.domain}</span>
                    <span className="text-slate-500">{item.percentage}%</span>
                  </div>
                  <div className="w-full h-2 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        item.percentage >= 70
                          ? 'bg-emerald-500'
                          : item.percentage >= 40
                          ? 'bg-blue-600'
                          : 'bg-amber-500'
                      }`}
                      style={{ width: `${item.percentage}%` }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : (
        /* Competency Detail Tab: Traceable Evidence Table */
        <div className="bg-white border border-slate-200/80 rounded-3xl shadow-card overflow-hidden">
          <div className="p-6 border-b border-slate-100">
            <h3 className="text-sm font-bold text-slate-900">Traceable Competency Records</h3>
            <p className="text-xs text-slate-500 mt-0.5">
              Click on any row to drill into exact assessment dates and scores.
            </p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-50 text-slate-500 uppercase font-bold border-b border-slate-100">
                <tr>
                  <th className="py-3.5 px-6">Concept Name</th>
                  <th className="py-3.5 px-4">Status</th>
                  <th className="py-3.5 px-4">Last Evidence</th>
                  <th className="py-3.5 px-4">Evaluation Type</th>
                  <th className="py-3.5 px-6 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {competencyDetail.map((c) => (
                  <tr
                    key={c.conceptId}
                    onClick={() => handleDrillIn(c.conceptId, c.conceptName)}
                    className="hover:bg-slate-50/80 cursor-pointer transition"
                  >
                    <td className="py-4 px-6 font-bold text-slate-900">{c.conceptName}</td>
                    <td className="py-4 px-4">
                      <span
                        className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase ${
                          c.state === 'competent'
                            ? 'bg-emerald-50 text-emerald-700'
                            : c.state === 'weak_evidence'
                            ? 'bg-amber-50 text-amber-700'
                            : 'bg-blue-50 text-blue-700'
                        }`}
                      >
                        {c.state}
                      </span>
                    </td>
                    <td className="py-4 px-4 text-slate-500">{c.lastEvidenceDate}</td>
                    <td className="py-4 px-4 font-mono text-slate-600 uppercase">{c.evidenceType}</td>
                    <td className="py-4 px-6 text-right text-blue-600 font-semibold flex items-center justify-end gap-1">
                      <span>Drill-in</span>
                      <ChevronRight className="w-3.5 h-3.5" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Drill-in Evidence History Modal */}
      {historyModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="fixed inset-0 bg-slate-900/40 backdrop-blur-sm" onClick={() => setHistoryModal(null)}></div>
          <div className="relative bg-white border border-slate-200 rounded-3xl p-6 max-w-lg w-full shadow-elevated z-10 space-y-4">
            <h3 className="text-base font-bold text-slate-900">
              Evidence Audit: {historyModal.conceptName}
            </h3>

            <div className="space-y-3">
              {historyModal.history?.map((h, i) => (
                <div key={i} className="p-4 bg-slate-50 rounded-2xl border border-slate-200 text-xs space-y-1">
                  <div className="flex justify-between font-bold text-slate-800">
                    <span>Attempt {h.attempt}</span>
                    <span className="text-emerald-600">{h.result} ({h.score}%)</span>
                  </div>
                  <p className="text-slate-500 text-[11px]">{h.date}</p>
                  <p className="text-slate-700 pt-1">{h.details}</p>
                </div>
              ))}
            </div>

            <button
              onClick={() => setHistoryModal(null)}
              className="w-full py-2.5 bg-slate-900 text-white rounded-xl text-xs font-semibold"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
