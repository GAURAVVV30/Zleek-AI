import React, { useState, useEffect } from 'react';
import { Check, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export const LEARNING_AVATARS = [
  {
    id: "memo-7",
    gender: "female",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_7.png",
  },
  {
    id: "memo-10",
    gender: "female",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_10.png",
  },
  {
    id: "memo-29",
    gender: "female",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_29.png",
  },
  {
    id: "memo-34",
    gender: "male",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_34.png",
  },
  {
    id: "memo-22",
    gender: "male",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_22.png",
  },
  {
    id: "memo-24",
    gender: "male",
    src: "https://cdn.jsdelivr.net/gh/alohe/avatars/png/memo_24.png",
  },
];

export default function AvatarSelector({ onSave }) {
  const [selectedAvatar, setSelectedAvatar] = useState(null);
  const [gender, setGender] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const savedGender = localStorage.getItem('onboardingGender');
    setGender(savedGender);

    const savedAvatar = localStorage.getItem('selectedLearningAvatar');
    if (savedAvatar && savedGender) {
      const avatarObj = LEARNING_AVATARS.find(a => a.id === savedAvatar);
      if (avatarObj && avatarObj.gender === savedGender) {
        setSelectedAvatar(savedAvatar);
      } else {
        localStorage.removeItem('selectedLearningAvatar');
      }
    }
  }, []);

  if (!gender) {
    return (
      <div className="bg-black/20 backdrop-blur-xl border border-white/10 rounded-3xl p-6 lg:p-10 flex flex-col items-center text-center gap-6 shadow-[0_0_20px_rgba(79,70,229,0.15)]">
        <h2 className="font-display text-2xl font-bold text-white">Choose Your Learning Avatar</h2>
        <p className="text-slate-400 text-sm">Complete your profile setup to choose an avatar.</p>
        <button 
          onClick={() => navigate('/onboarding/goal')}
          className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold flex items-center gap-2 transition"
        >
          Continue Setup <ArrowRight className="w-4 h-4" />
        </button>
      </div>
    );
  }

  const availableAvatars = LEARNING_AVATARS.filter(a => a.gender === gender);

  const handleSave = () => {
    if (selectedAvatar) {
      localStorage.setItem('selectedLearningAvatar', selectedAvatar);
      if (onSave) onSave();
    }
  };

  return (
    <div className="bg-black/20 backdrop-blur-xl border border-white/10 rounded-3xl p-6 lg:p-10 flex flex-col gap-8 shadow-[0_0_20px_rgba(79,70,229,0.15)]">
      <div className="text-center md:text-left">
        <h2 className="font-display text-2xl font-bold text-white mb-2">Choose Your Learning Avatar</h2>
        <p className="text-slate-400 text-sm">Pick a character that represents you on your learning journey.</p>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 lg:gap-8 max-w-3xl mx-auto md:mx-0">
        {availableAvatars.map((avatar) => {
          const isSelected = selectedAvatar === avatar.id;
          return (
            <button
              key={avatar.id}
              onClick={() => setSelectedAvatar(avatar.id)}
              aria-label={`Select learning avatar`}
              className={`relative aspect-square rounded-2xl border transition-all duration-300 flex items-center justify-center p-2 outline-none focus-visible:ring-4 focus-visible:ring-indigo-500/50 ${
                isSelected
                  ? 'bg-indigo-900/40 border-indigo-400 shadow-[0_0_20px_rgba(79,70,229,0.3)] scale-[1.03]'
                  : 'bg-black/30 border-white/10 hover:border-white/30 hover:scale-[1.03] hover:shadow-[0_0_15px_rgba(79,70,229,0.15)] hover:bg-slate-900/60'
              }`}
            >
              <img src={avatar.src} alt="Learning avatar" className="w-[85%] h-[85%] object-contain pointer-events-none drop-shadow-lg" />
              
              {isSelected && (
                <div className="absolute bottom-3 right-3 bg-indigo-500 rounded-full p-1 shadow-lg border border-indigo-400 animate-in fade-in zoom-in duration-200">
                  <Check className="w-4 h-4 text-white" />
                </div>
              )}
            </button>
          );
        })}
      </div>

      <div className="flex justify-end pt-4 border-t border-white/10">
        <button
          onClick={handleSave}
          disabled={!selectedAvatar}
          className={`flex items-center gap-2 px-8 py-3 rounded-xl font-bold transition-all ${
            selectedAvatar
              ? 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-[0_0_15px_rgba(79,70,229,0.4)] hover:shadow-[0_0_25px_rgba(79,70,229,0.5)]'
              : 'bg-slate-800 text-slate-500 cursor-not-allowed border border-slate-700'
          }`}
        >
          Save Avatar &rarr;
        </button>
      </div>
    </div>
  );
}
