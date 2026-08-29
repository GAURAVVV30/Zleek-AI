import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  CheckCircle2,
  PlayCircle,
  Lock,
  AlertTriangle,
  ArrowRight,
  Sparkles,
  Info,
  Layers,
  Crown,
  Zap,
} from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { NODE_STATES, NODE_CONFIG } from '../../utils/constants';
import AvatarViewer3D from '../../components/character/AvatarViewer3D';
import { CHARACTERS_ROSTER } from '../../components/character/CharacterBattleCustomizer';

export default function RoadmapPage() {
  const [roadmap, setRoadmap] = useState(null);
  const [selectedConcept, setSelectedConcept] = useState(null);
  const [explainModal, setExplainModal] = useState(null);
  const [activeAvatar, setActiveAvatar] = useState(CHARACTERS_ROSTER[0]);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.ROADMAP.BASE)
      .then((res) => {
        setRoadmap(res.data);
        const activeNode = res.data.nodes.find((n) => n.state === NODE_STATES.IN_PROGRESS) || res.data.nodes[0];
        setSelectedConcept(activeNode);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  const handleWhyConcept = async (conceptId) => {
    try {
      const res = await apiClient.get(ENDPOINTS.ROADMAP.WHY_CONCEPT(conceptId));
      setExplainModal(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  if (isLoading || !roadmap) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-500">Loading your live adaptive roadmap...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Top Goal Banner with 3D Companion Live Status */}
      <div className="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 border border-slate-700/80 rounded-3xl p-6 sm:p-8 text-white shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative overflow-hidden">
        <div className="space-y-2 relative z-10">
          <div className="flex items-center gap-2">
            <Crown className="w-4 h-4 text-amber-400" />
            <span className="text-[11px] font-mono font-bold text-amber-400 uppercase tracking-widest">
              Active Milestone Objective
            </span>
          </div>
          <h1 className="font-display text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
            {roadmap.goalTitle}
          </h1>
          <p className="text-xs text-slate-300">
            {roadmap.progressPercentage}% of knowledge graph verified through quizzes and projects.
          </p>
        </div>

        <div className="flex items-center gap-4 shrink-0 relative z-10">
          <div className="text-right">
            <span className="text-sm font-bold text-cyan-400 block">{roadmap.progressPercentage}% Competent</span>
            <span className="text-[10px] text-slate-400">Verified Evidence</span>
          </div>
          <div className="w-20 h-2 bg-slate-700 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-cyan-400 to-blue-500 rounded-full"
              style={{ width: `${roadmap.progressPercentage}%` }}
            />
          </div>
        </div>
      </div>

      {/* Main Split View: Vertical Graph on Left, Selected Node & 3D Companion on Right */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column: Vertical Interactive Path */}
        <div className="lg:col-span-7 bg-white border border-slate-200/80 rounded-3xl p-6 sm:p-8 shadow-card">
          <div className="flex items-center justify-between pb-4 border-b border-slate-100 mb-6">
            <h2 className="text-sm font-bold text-slate-900 flex items-center gap-2">
              <Layers className="w-4 h-4 text-blue-600" />
              Verified Competency Roadmap
            </h2>
            <span className="text-xs text-slate-400">{roadmap.nodes.length} Milestones</span>
          </div>

          <div className="relative pl-6 space-y-6 before:absolute before:left-9 before:top-4 before:bottom-4 before:w-0.5 before:bg-slate-200">
            {roadmap.nodes.map((node) => {
              const isSelected = selectedConcept?.id === node.id;
              const config = NODE_CONFIG[node.state] || NODE_CONFIG[NODE_STATES.NOT_STARTED];

              return (
                <div
                  key={node.id}
                  onClick={() => setSelectedConcept(node)}
                  className={`relative flex items-start gap-4 p-4 rounded-2xl border transition cursor-pointer group ${
                    isSelected
                      ? 'border-blue-600 bg-blue-50/50 shadow-md ring-2 ring-blue-500/20'
                      : 'border-slate-200/80 bg-white hover:border-slate-300 hover:bg-slate-50/60'
                  }`}
                >
                  {/* Status Indicator Icon */}
                  <div
                    className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 z-10 ${
                      node.state === NODE_STATES.COMPETENT
                        ? 'bg-emerald-500 text-white'
                        : node.state === NODE_STATES.IN_PROGRESS
                        ? 'bg-blue-600 text-white ring-4 ring-blue-100'
                        : node.state === NODE_STATES.WEAK_EVIDENCE
                        ? 'bg-amber-500 text-white'
                        : 'bg-slate-200 text-slate-500'
                    }`}
                  >
                    {node.state === NODE_STATES.COMPETENT && <CheckCircle2 className="w-4 h-4" />}
                    {node.state === NODE_STATES.IN_PROGRESS && <PlayCircle className="w-4 h-4" />}
                    {node.state === NODE_STATES.WEAK_EVIDENCE && <AlertTriangle className="w-4 h-4" />}
                    {node.state === NODE_STATES.NOT_STARTED && <Lock className="w-3.5 h-3.5" />}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <h3 className="text-sm font-bold text-slate-900 group-hover:text-blue-600 transition truncate">
                        {node.title}
                      </h3>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border ${config.badgeClass}`}>
                        {config.label}
                      </span>
                    </div>

                    <p className="text-xs text-slate-500 mt-1 line-clamp-1">
                      {node.description}
                    </p>

                    {node.unlockRequirement && (
                      <p className="text-[11px] text-slate-400 mt-1 italic">
                        {node.unlockRequirement}
                      </p>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Column: Selected Node Preview & 3D Companion Stage */}
        <div className="lg:col-span-5 space-y-6 sticky top-24">
          {/* Selected Concept Preview */}
          <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-card space-y-4">
            {selectedConcept ? (
              <>
                <div>
                  <span className="text-[11px] font-bold text-blue-600 uppercase tracking-wider block mb-1">
                    Selected Milestone
                  </span>
                  <h3 className="font-display text-lg font-extrabold text-slate-900">
                    {selectedConcept.title}
                  </h3>
                  <p className="text-xs text-slate-600 mt-1 leading-relaxed">
                    {selectedConcept.description}
                  </p>
                </div>

                <div className="p-3.5 bg-slate-50 border border-slate-200/80 rounded-2xl flex items-center justify-between">
                  <span className="text-xs font-bold text-slate-800 flex items-center gap-1.5">
                    <Sparkles className="w-3.5 h-3.5 text-blue-600" />
                    Why this milestone?
                  </span>
                  <button
                    onClick={() => handleWhyConcept(selectedConcept.id)}
                    className="text-xs text-blue-600 hover:underline font-semibold"
                  >
                    Explain logic
                  </button>
                </div>

                <div>
                  {selectedConcept.state === NODE_STATES.NOT_STARTED ? (
                    <button
                      disabled
                      className="w-full py-3 bg-slate-100 text-slate-400 font-semibold rounded-xl text-xs flex items-center justify-center gap-2 cursor-not-allowed"
                    >
                      <Lock className="w-4 h-4" /> Locked (Prerequisites Pending)
                    </button>
                  ) : (
                    <button
                      onClick={() => navigate(`/learn/${selectedConcept.id}`)}
                      className="w-full py-3.5 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl text-xs shadow-elevated hover:shadow-glow transition flex items-center justify-center gap-2 group"
                    >
                      <span>{selectedConcept.state === NODE_STATES.COMPETENT ? 'Review Concept' : 'Continue Learning'}</span>
                      <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition" />
                    </button>
                  )}
                </div>
              </>
            ) : null}
          </div>

          {/* 3D Holographic Companion Stage Card */}
          <div className="bg-slate-950 border border-slate-800 rounded-3xl p-5 text-white shadow-xl relative overflow-hidden">
            <div className="flex items-center justify-between pb-3 border-b border-slate-800">
              <div className="flex items-center gap-2">
                <Zap className="w-4 h-4 text-cyan-400" />
                <span className="text-xs font-mono font-bold text-cyan-400">
                  {activeAvatar.name}
                </span>
              </div>
              <span className="px-2 py-0.5 bg-cyan-500/20 text-cyan-300 rounded text-[10px] font-mono">
                {activeAvatar.powerTitle}
              </span>
            </div>

            <div className="h-60 relative flex items-center justify-center">
              <AvatarViewer3D character={activeAvatar} auraColor={activeAvatar.aura} />
            </div>

            <p className="text-[11px] text-slate-400 text-center italic mt-1">
              "{activeAvatar.powerDesc}"
            </p>
          </div>
        </div>
      </div>

      {/* Explainability Modal Drawer */}
      {explainModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="fixed inset-0 bg-slate-900/40 backdrop-blur-sm" onClick={() => setExplainModal(null)}></div>
          <div className="relative bg-white border border-slate-200 rounded-3xl p-6 max-w-lg w-full shadow-elevated z-10 space-y-4">
            <div className="flex items-center gap-2 text-blue-600 font-bold text-sm">
              <Info className="w-4 h-4" />
              Prerequisite & Reasoning Trail
            </div>
            <h3 className="text-base font-bold text-slate-900">
              Why learn {explainModal.conceptName}?
            </h3>
            <p className="text-xs text-slate-600 leading-relaxed bg-slate-50 p-4 rounded-xl border border-slate-100">
              {explainModal.reason}
            </p>
            <button
              onClick={() => setExplainModal(null)}
              className="w-full py-2.5 bg-slate-900 text-white rounded-xl text-xs font-semibold hover:bg-slate-800 transition"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
