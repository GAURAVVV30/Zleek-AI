import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Target, Sparkles, ArrowRight } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';
import CharacterBattleCustomizer from '../../components/character/CharacterBattleCustomizer';

export default function GoalDefinitionPage() {
  const [goalText, setGoalText] = useState('I want to become a Data Scientist and build real-world ML projects.');
  const [selectedAvatar, setSelectedAvatar] = useState({ id: 'robot', name: 'Cyber Mecha-01' });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const { addToast } = useToast();

  const suggestedDomains = [
    'Data Science',
    'Machine Learning',
    'AI Engineering',
    'Backend Go',
  ];

  const handleContinue = async () => {
    if (!goalText.trim()) {
      addToast('Please enter a goal statement.', 'warning');
      return;
    }

    setIsSubmitting(true);
    try {
      await apiClient.post(ENDPOINTS.GOALS.BASE, {
        goalText,
        avatarId: selectedAvatar.id,
      });
      addToast(`${selectedAvatar.name} bonded to your learning journey!`, 'success');
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
      <div className="bg-white border border-slate-200/80 rounded-3xl p-6 sm:p-8 shadow-card space-y-6">
        <div className="flex items-center justify-between pb-4 border-b border-slate-100">
          <span className="text-xs font-bold text-blue-600 uppercase tracking-wider">
            Step 1 of 3 · Goal Definition & Companion Bond
          </span>
          <span className="text-xs text-slate-400">Onboarding</span>
        </div>

        <div className="space-y-1">
          <h1 className="font-display text-2xl sm:text-3xl font-extrabold text-slate-900">
            What do you want to achieve?
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 leading-relaxed">
            Enter your target career role, and select a 3D AI Companion with unique pedagogical superpowers.
          </p>
        </div>

        {/* Goal Text Area */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 mb-2">
            Target Objective / Career Milestone
          </label>
          <div className="relative">
            <textarea
              rows={3}
              value={goalText}
              onChange={(e) => setGoalText(e.target.value)}
              className="w-full p-4 text-sm border border-slate-200 rounded-2xl focus:ring-2 focus:ring-blue-600 focus:border-transparent outline-none leading-relaxed text-slate-800"
              placeholder="e.g. I want to become a Senior Data Scientist proficient in Pandas and Machine Learning..."
            />
            <Sparkles className="w-5 h-5 text-blue-500 absolute right-3 bottom-3 pointer-events-none" />
          </div>
        </div>

        {/* Domain Chips */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 mb-2">
            Suggested Domains
          </label>
          <div className="flex flex-wrap gap-2">
            {suggestedDomains.map((domain) => (
              <button
                key={domain}
                type="button"
                onClick={() => setGoalText(`I want to master ${domain} from the ground up.`)}
                className="px-3.5 py-1.5 rounded-full text-xs font-medium bg-slate-50 hover:bg-blue-50 hover:text-blue-700 border border-slate-200 hover:border-blue-200 transition"
              >
                {domain}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* 2. Free-Fire Style 3D Character Companion Customizer */}
      <CharacterBattleCustomizer
        initialAvatar="robot"
        onSelectAvatar={(char, color) => setSelectedAvatar({ ...char, color })}
      />

      {/* 3. Bottom Action CTA */}
      <div className="flex justify-end pt-2">
        <button
          onClick={handleContinue}
          disabled={isSubmitting}
          className="px-8 py-4 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-bold text-sm shadow-elevated hover:shadow-glow transition flex items-center gap-2"
        >
          {isSubmitting ? 'Bonding companion...' : 'Continue to Preferences'}
          <ArrowRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
