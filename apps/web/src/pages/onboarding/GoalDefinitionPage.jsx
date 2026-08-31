import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Target, Sparkles, ArrowRight, Check } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';
import { useAuth } from '../../context/AuthContext';
export default function GoalDefinitionPage() {
  const [goalText, setGoalText] = useState('I want to become a Data Scientist and build real-world ML projects.');
  const [gender, setGender] = useState(null);
  const [selectedAvatar, setSelectedAvatar] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { user } = useAuth();
  const username = user?.fullName?.split(' ')[0] || 'Learner';

  const learningAvatars = [
    { id: "memo-7", gender: "female", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_7.png" },
    { id: "memo-10", gender: "female", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_10.png" },
    { id: "memo-29", gender: "female", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_29.png" },
    { id: "memo-34", gender: "male", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_34.png" },
    { id: "memo-22", gender: "male", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_22.png" },
    { id: "memo-24", gender: "male", src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_24.png" },
  ];
  const navigate = useNavigate();
  const { addToast } = useToast();

  const [domains, setDomains] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fallbackDomains = [
    { id: 'ai_engineer', name: 'AI Engineer' },
    { id: 'backend_engineer', name: 'Backend Engineer' },
    { id: 'data_engineer', name: 'Data Engineer' },
    { id: 'frontend_engineer', name: 'Frontend Engineer' },
    { id: 'full_stack', name: 'Full Stack' },
    { id: 'machine_learning', name: 'Machine Learning' },
    { id: 'mobile_engineer', name: 'Mobile Engineer' },
    { id: 'product_manager', name: 'Product Manager' },
    { id: 'data_analyst', name: 'Data Analyst' },
    { id: 'devops_sre', name: 'DevOps/SRE' },
    { id: 'software_architecture', name: 'Software Architect' },
    { id: 'ai_data_scientist', name: 'AI/Data Scientist' }
  ];

  useEffect(() => {
    const fetchDomains = async () => {
      try {
        setLoading(true);
        const res = await apiClient.get(ENDPOINTS.DOMAINS);
        if (res?.success && Array.isArray(res.data)) {
          setDomains(res.data);
          setError(null);
        } else {
          throw new Error('Invalid domains response format');
        }
      } catch (err) {
        console.error('Failed to fetch domains:', err);
        setError('Failed to load career roles. Using local fallback.');
      } finally {
        setLoading(false);
      }
    };
    fetchDomains();
  }, []);

  const handleContinue = async () => {
    if (!goalText.trim()) {
      addToast('Please enter a goal statement.', 'warning');
      return;
    }
    if (!gender) {
      addToast('Please select your gender to personalize your avatar.', 'warning');
      return;
    }
    if (!selectedAvatar) {
      addToast('Please select an avatar to continue.', 'warning');
      return;
    }

    setIsSubmitting(true);
    try {
      await apiClient.post(ENDPOINTS.GOALS.BASE, {
        goalText,
      });
      localStorage.setItem("onboardingGender", gender);
      localStorage.setItem("selectedLearningAvatar", selectedAvatar);
      addToast(`Goal mapped successfully!`, 'success');
      navigate('/onboarding/preferences');
    } catch (err) {
      addToast(err?.message || 'Goal mapping failed', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-8 space-y-8">
      {/* 1. Goal Input Card */}
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
        <div className="flex items-center justify-between pb-4 border-b border-white/5">
          <span className="text-xs font-bold text-indigo-400 uppercase tracking-wider">
            Step 1 of 3 · Goal Definition & Companion Bond
          </span>
          <span className="text-xs text-slate-400">Onboarding</span>
        </div>

        <div className="space-y-1">
          <h1 className="font-display text-2xl sm:text-3xl font-extrabold text-white">
            What do you want to achieve?
          </h1>
          <p className="text-xs sm:text-sm text-slate-400 leading-relaxed">
            Enter your target career role, and select a 3D AI Companion with unique pedagogical superpowers.
          </p>
        </div>

        {/* Goal Text Area */}
        <div>
          <label className="block text-xs font-semibold text-white mb-2">
            Target Objective / Career Milestone
          </label>
          <div className="relative">
            <textarea
              rows={3}
              value={goalText}
              onChange={(e) => setGoalText(e.target.value)}
              className="w-full p-4 text-sm bg-white/5 backdrop-blur-md border border-white/10 rounded-2xl focus:ring-2 focus:ring-blue-600 focus:border-transparent outline-none leading-relaxed text-white placeholder:text-white/40"
              placeholder="e.g. I want to become a Senior Data Scientist proficient in Pandas and Machine Learning..."
            />
            <Sparkles className="w-5 h-5 text-blue-500 absolute right-3 bottom-3 pointer-events-none" />
          </div>
        </div>

        {/* Domain Chips */}
        <div>
          <label className="block text-xs font-semibold text-white mb-2">
            Suggested Domains
          </label>
          {loading && (
            <div className="text-xs text-slate-400 mb-2 animate-pulse">Loading career roles...</div>
          )}
          {error && (
            <div className="text-xs text-indigo-400 mb-2">{error}</div>
          )}
          <div className="flex flex-wrap gap-2">
            {(domains && domains.length > 0 ? domains : fallbackDomains).map((domain) => (
              <button
                key={domain.id}
                type="button"
                onClick={() => setGoalText(`I want to master ${domain.name} from the ground up.`)}
                className="px-3.5 py-1.5 rounded-full text-xs font-medium bg-black/30 backdrop-blur-md hover:bg-indigo-900/40 backdrop-blur-sm hover:text-indigo-400 border border-white/10 hover:border-blue-200 transition animate-in fade-in duration-300"
              >
                {domain.name}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* 2. Select Gender for Avatar Personalization */}
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:px-8 sm:py-6 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-5 max-w-xl mx-auto w-full">
        <div className="space-y-1 text-center">
          <h2 className="font-display text-lg sm:text-xl font-extrabold text-white">
            Personalize Your Avatar
          </h2>
          <p className="text-[11px] sm:text-xs text-slate-400 leading-relaxed">
            Select your gender to generate personalized learning avatar options.
          </p>
        </div>
        
        <div className="grid grid-cols-2 gap-3 max-w-[280px] mx-auto">
          <button
            onClick={() => { setGender('female'); setSelectedAvatar(null); }}
            className={`py-2 px-4 rounded-xl border flex items-center justify-center transition-all ${
              gender === 'female' 
                ? 'bg-indigo-900/40 border-indigo-400 shadow-[0_0_20px_rgba(79,70,229,0.3)] scale-[1.02]'
                : 'bg-black/30 border-white/10 hover:border-white/30 hover:bg-slate-900/60'
            }`}
          >
            <span className="text-sm font-bold text-white">Female</span>
          </button>
          
          <button
            onClick={() => { setGender('male'); setSelectedAvatar(null); }}
            className={`py-2 px-4 rounded-xl border flex items-center justify-center transition-all ${
              gender === 'male' 
                ? 'bg-indigo-900/40 border-indigo-400 shadow-[0_0_20px_rgba(79,70,229,0.3)] scale-[1.02]'
                : 'bg-black/30 border-white/10 hover:border-white/30 hover:bg-slate-900/60'
            }`}
          >
            <span className="text-sm font-bold text-white">Male</span>
          </button>
        </div>

        {gender && (
          <div className="pt-4 border-t border-white/10 space-y-3 animate-in fade-in slide-in-from-top-2 duration-300">
            {!selectedAvatar ? (
              <>
                <h3 className="text-xs font-bold text-slate-300 text-center uppercase tracking-wider">Select Your Avatar</h3>
                <div className="grid grid-cols-3 gap-3 max-w-sm mx-auto">
                  {learningAvatars
                    .filter((a) => a.gender === gender)
                    .map((avatar) => {
                      const isSelected = selectedAvatar === avatar.id;
                      return (
                        <button
                          key={avatar.id}
                          onClick={() => setSelectedAvatar(avatar.id)}
                          className={`relative aspect-[4/5] rounded-xl border transition-all duration-300 flex items-center justify-center p-2 outline-none focus-visible:ring-4 focus-visible:ring-indigo-500/50 ${
                            isSelected
                              ? 'bg-indigo-900/40 border-indigo-400 shadow-[0_0_20px_rgba(79,70,229,0.3)] scale-[1.03]'
                              : 'bg-black/30 border-white/10 hover:border-white/30 hover:bg-slate-900/60'
                          }`}
                        >
                          <img src={avatar.src} alt="Avatar option" className="w-[90%] h-[90%] object-contain drop-shadow-[0_10px_20px_rgba(0,0,0,0.5)] pointer-events-none" />
                          {isSelected && (
                            <div className="absolute bottom-2 right-2 bg-indigo-500 rounded-full p-1 shadow-lg border border-indigo-400 animate-in fade-in zoom-in duration-200">
                              <Check className="w-3.5 h-3.5 text-white" />
                            </div>
                          )}
                        </button>
                      );
                    })}
                </div>
              </>
            ) : (
              <div className="flex flex-col items-center justify-center space-y-4 py-4 animate-in fade-in zoom-in duration-500">
                <span 
                  style={{ fontFamily: "'Cormorant Garamond', serif", letterSpacing: '0.05em' }} 
                  className="text-2xl sm:text-3xl text-white/90 italic font-light"
                >
                  Hellooo, {username}!
                </span>
                
                <div className="relative w-32 h-32 flex items-center justify-center group">
                  {/* Subtle glow */}
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="w-24 h-24 bg-indigo-500/30 blur-[30px] rounded-full pointer-events-none transition-all duration-500 group-hover:bg-indigo-400/40 group-hover:blur-[40px]"></div>
                  </div>
                  <img 
                    src={learningAvatars.find(a => a.id === selectedAvatar)?.src} 
                    alt="Selected learning avatar" 
                    className="relative w-full h-full object-contain pointer-events-none drop-shadow-[0_10px_20px_rgba(0,0,0,0.5)] transition-transform duration-500 group-hover:scale-105" 
                  />
                </div>
                
                <button 
                  onClick={() => setSelectedAvatar(null)}
                  className="text-xs text-slate-400 hover:text-white transition underline underline-offset-4 mt-4"
                >
                  Choose a different avatar
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* 3. Bottom Action CTA */}
      <div className="flex justify-end pt-2">
        <button
          onClick={handleContinue}
          disabled={isSubmitting}
          className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-sm shadow-elevated hover:shadow-glow transition flex items-center gap-2"
        >
          {isSubmitting ? 'Saving...' : 'Next'}
          <ArrowRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
