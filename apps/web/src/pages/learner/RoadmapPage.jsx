import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Info } from 'lucide-react';

import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { NODE_STATES } from '../../utils/constants';
import { useAuth } from '../../context/AuthContext';

import AvatarSelector, { LEARNING_AVATARS } from '../../components/dashboard/AvatarSelector';
import CurrentMission from '../../components/dashboard/CurrentMission';
import CompetencyRoadmap from '../../components/dashboard/CompetencyRoadmap';

export default function RoadmapPage() {
  const [roadmap, setRoadmap] = useState(null);
  const [selectedConcept, setSelectedConcept] = useState(null);
  const [explainModal, setExplainModal] = useState(null);
  const [activeAvatar, setActiveAvatar] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const { user } = useAuth();
  const username = user?.fullName?.split(' ')[0] || 'Learner';

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.ROADMAP.BASE)
      .then((res) => {
        setRoadmap(res.data);
        const activeNode = res.data.nodes?.find((n) => n.state === NODE_STATES.IN_PROGRESS) || res.data.nodes?.[0];
        setSelectedConcept(activeNode);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));

    const savedAvatarId = localStorage.getItem('selectedLearningAvatar');
    if (savedAvatarId) {
      const found = LEARNING_AVATARS.find(a => a.id === savedAvatarId);
      if (found) setActiveAvatar(found);
    }
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
        <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading your live adaptive roadmap...</p>
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col xl:flex-row gap-8 items-start w-full justify-center">
        {/* Center Column: Mission Header & Competency Module Roadmap */}
        <div className="flex-1 w-full xl:max-w-[1000px] min-w-0 flex flex-col gap-8">
          <CurrentMission roadmap={roadmap} />

          <div className="w-full">
            <CompetencyRoadmap 
              roadmap={roadmap} 
              selectedConcept={selectedConcept} 
              onSelectConcept={setSelectedConcept} 
            />
          </div>
        </div>

        {/* Right Column: Personal Avatar */}
        <div className="w-full xl:w-[340px] shrink-0 space-y-8 sticky top-32 xl:ml-12">
          {!activeAvatar ? (
            <AvatarSelector onSave={() => {
              const savedAvatarId = localStorage.getItem('selectedLearningAvatar');
              if (savedAvatarId) {
                const found = LEARNING_AVATARS.find(a => a.id === savedAvatarId);
                if (found) setActiveAvatar(found);
              }
            }} />
          ) : (
            <div className="bg-slate-950/80 backdrop-blur-xl border border-white/10 rounded-[32px] shadow-[0_0_40px_rgba(79,70,229,0.15)] relative overflow-hidden flex flex-col items-center justify-center h-[360px] py-6">
              {/* Greeting */}
              <div 
                className="z-10 text-[20px] sm:text-[22px] lg:text-[26px] text-white/90 mb-4"
                style={{ fontFamily: "'Cormorant Garamond', serif", letterSpacing: '0.05em' }}
              >
                <span className="italic font-light">Hellooo, {username}!</span>
              </div>

              {/* Subtle glow behind the avatar */}
              <div className="absolute inset-0 flex items-center justify-center mt-8">
                <div className="w-[180px] h-[180px] bg-indigo-500/20 blur-[50px] rounded-full pointer-events-none"></div>
              </div>
              
              <div className="relative w-[180px] h-[180px] flex items-center justify-center transition-transform hover:scale-105 duration-500">
                <img src={activeAvatar.src} alt="Learning avatar" className="relative w-full h-full object-contain pointer-events-none drop-shadow-[0_10px_20px_rgba(0,0,0,0.5)]" />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Explainability Modal Drawer */}
      {explainModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="fixed inset-0 bg-slate-900/60 backdrop-blur-md" onClick={() => setExplainModal(null)}></div>
          <div className="relative bg-slate-900/90 backdrop-blur-xl border border-white/10 rounded-3xl p-6 max-w-lg w-full shadow-[0_0_50px_rgba(79,70,229,0.2)] z-10 space-y-4">
            <div className="flex items-center gap-2 text-indigo-400 font-bold text-sm">
              <Info className="w-4 h-4" />
              Prerequisite & Reasoning Trail
            </div>
            <h3 className="text-base font-bold text-white">
              Why learn {explainModal.conceptName}?
            </h3>
            <p className="text-sm text-slate-300 leading-relaxed bg-black/40 backdrop-blur-md p-4 rounded-xl border border-white/5">
              {explainModal.reason}
            </p>
          </div>
        </div>
      )}
    </>
  );
}
