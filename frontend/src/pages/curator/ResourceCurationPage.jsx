import React, { useState, useEffect } from 'react';
import { Inbox, Check, X, ThumbsUp, Video, BookOpen, Clock } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function ResourceCurationPage() {
  const [resources, setResources] = useState([]);
  const [activeTab, setActiveTab] = useState('pending');
  const [isLoading, setIsLoading] = useState(true);
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.CURATOR.RESOURCES)
      .then((res) => {
        setResources(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  const handleDecision = (id, decision) => {
    setResources((prev) =>
      prev.map((r) => (r.id === id ? { ...r, status: decision } : r))
    );
    addToast(`Candidate resource marked as ${decision}.`, 'success');
  };

  const filtered = resources.filter((r) => r.status === activeTab);

  return (
    <div className="space-y-6">
      <div className="bg-white border border-purple-200/80 rounded-2xl p-6 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <span className="text-[11px] font-bold text-purple-700 uppercase tracking-wider block mb-1">
            Curation Queue
          </span>
          <h1 className="font-display text-xl font-extrabold text-slate-900">
            Gold-Standard Resource Review
          </h1>
        </div>

        <div className="flex bg-slate-100 p-1 rounded-xl text-xs font-semibold">
          <button
            onClick={() => setActiveTab('pending')}
            className={`px-3 py-1.5 rounded-lg transition ${
              activeTab === 'pending' ? 'bg-white text-purple-700 shadow-sm font-bold' : 'text-slate-600'
            }`}
          >
            Pending Review ({resources.filter((r) => r.status === 'pending').length})
          </button>
          <button
            onClick={() => setActiveTab('approved')}
            className={`px-3 py-1.5 rounded-lg transition ${
              activeTab === 'approved' ? 'bg-white text-purple-700 shadow-sm font-bold' : 'text-slate-600'
            }`}
          >
            Approved
          </button>
        </div>
      </div>

      <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-card space-y-4">
        {filtered.length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-500">
            No resources in this queue status.
          </div>
        ) : (
          filtered.map((item) => (
            <div
              key={item.id}
              className="p-4 bg-slate-50/70 border border-slate-200 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-4"
            >
              <div className="flex items-start gap-3">
                <div className="w-10 h-10 rounded-xl bg-purple-100 text-purple-600 flex items-center justify-center shrink-0">
                  {item.type === 'video' ? <Video className="w-5 h-5" /> : <BookOpen className="w-5 h-5" />}
                </div>
                <div>
                  <h3 className="text-xs sm:text-sm font-bold text-slate-900">{item.title}</h3>
                  <p className="text-[11px] text-slate-500 mt-0.5">
                    {item.duration} · Discovered by {item.curator}
                  </p>
                </div>
              </div>

              {item.status === 'pending' && (
                <div className="flex items-center gap-2 shrink-0">
                  <button
                    onClick={() => handleDecision(item.id, 'approved')}
                    className="px-3.5 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-1"
                  >
                    <Check className="w-3.5 h-3.5" /> Approve
                  </button>
                  <button
                    onClick={() => handleDecision(item.id, 'retired')}
                    className="px-3.5 py-1.5 bg-red-50 hover:bg-red-100 text-red-600 border border-red-200 rounded-xl text-xs font-bold transition flex items-center gap-1"
                  >
                    <X className="w-3.5 h-3.5" /> Reject
                  </button>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
