import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Clock, Video, BookOpen, Layers, ArrowRight, ArrowLeft } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function PreferencesPage() {
  const [weeklyHours, setWeeklyHours] = useState('5_10');
  const [preferredFormat, setPreferredFormat] = useState(['video', 'article']);
  const [experienceLevel, setExperienceLevel] = useState('intermediate');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const { addToast } = useToast();

  const toggleFormat = (fmt) => {
    if (preferredFormat.includes(fmt)) {
      if (preferredFormat.length > 1) {
        setPreferredFormat(preferredFormat.filter((f) => f !== fmt));
      }
    } else {
      setPreferredFormat([...preferredFormat, fmt]);
    }
  };

  const handleContinue = async () => {
    setIsSubmitting(true);
    try {
      await apiClient.patch(ENDPOINTS.PROFILE.PREFERENCES, {
        weeklyHours,
        preferredFormat,
        experienceLevel,
      });
      addToast('Preferences saved! Ready for baseline diagnostic.', 'success');
      navigate('/diagnostic');
    } catch (err) {
      addToast('Could not save preferences', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-12">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-white/5">
          <span className="text-xs font-bold text-indigo-400 uppercase tracking-wider">Step 2 of 3 · Preferences</span>
          <span className="text-xs text-slate-400">Onboarding</span>
        </div>

        <div>
          <h1 className="font-display text-2xl font-bold text-white">
            Customize your learning experience
          </h1>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            We adapt resource duration and technical depth to your schedule and experience.
          </p>
        </div>

        {/* 1. Weekly Time Commitment */}
        <div>
          <label className="block text-xs font-semibold text-white mb-3 flex items-center gap-2">
            <Clock className="w-4 h-4 text-indigo-400" />
            Weekly Time Commitment
          </label>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {[
              { id: 'lt_5', label: '< 5 hours' },
              { id: '5_10', label: '5-10 hours' },
              { id: '10_20', label: '10-20 hours' },
              { id: 'gt_20', label: '20+ hours' },
            ].map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => setWeeklyHours(item.id)}
                className={`py-3 px-4 rounded-xl text-xs font-medium border text-center transition ${
                  weeklyHours === item.id
                    ? 'border-blue-600 bg-indigo-900/40 backdrop-blur-sm/70 text-indigo-400 font-bold shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                    : 'border-white/10 text-white hover:bg-black/30 backdrop-blur-md'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* 2. Preferred Learning Formats */}
        <div>
          <label className="block text-xs font-semibold text-white mb-3 flex items-center gap-2">
            <Video className="w-4 h-4 text-indigo-400" />
            Preferred Learning Format
          </label>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {[
              { id: 'video', label: 'Videos & Demos', icon: Video },
              { id: 'article', label: 'In-Depth Articles', icon: BookOpen },
              { id: 'interactive', label: 'Hands-on Labs', icon: Layers },
            ].map((item) => {
              const Icon = item.icon;
              const isSelected = preferredFormat.includes(item.id);
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => toggleFormat(item.id)}
                  className={`p-3.5 rounded-xl border flex items-center gap-2.5 text-left text-xs transition ${
                    isSelected
                      ? 'border-blue-600 bg-indigo-900/40 backdrop-blur-sm/70 text-indigo-400 font-bold'
                      : 'border-white/10 text-white hover:bg-black/30 backdrop-blur-md'
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isSelected ? 'text-indigo-400' : 'text-slate-400'}`} />
                  <span>{item.label}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* 3. Prior Experience */}
        <div>
          <label className="block text-xs font-semibold text-white mb-3">
            Prior Experience Level
          </label>
          <div className="grid grid-cols-3 gap-3">
            {[
              { id: 'beginner', label: 'Beginner', desc: 'Starting fresh' },
              { id: 'intermediate', label: 'Intermediate', desc: 'Some syntax/tools' },
              { id: 'advanced', label: 'Advanced', desc: 'Practical experience' },
            ].map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => setExperienceLevel(item.id)}
                className={`p-3 rounded-xl border text-center transition ${
                  experienceLevel === item.id
                    ? 'border-blue-600 bg-indigo-900/40 backdrop-blur-sm/70 text-indigo-400 font-bold'
                    : 'border-white/10 text-white hover:bg-black/30 backdrop-blur-md'
                }`}
              >
                <p className="text-xs font-semibold">{item.label}</p>
                <p className="text-[10px] text-slate-400 mt-0.5">{item.desc}</p>
              </button>
            ))}
          </div>
        </div>

        {/* Navigation CTAs */}
        <div className="pt-4 flex items-center justify-between">
          <button
            type="button"
            onClick={() => navigate('/onboarding/goal')}
            className="px-4 py-2.5 text-xs font-semibold text-slate-300 hover:text-white flex items-center gap-1.5"
          >
            <ArrowLeft className="w-4 h-4" /> Back
          </button>
          <button
            onClick={handleContinue}
            disabled={isSubmitting}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-semibold text-xs shadow-[0_0_15px_rgba(79,70,229,0.2)] shadow-blue-500/20 transition flex items-center gap-2"
          >
            {isSubmitting ? 'Saving...' : 'Start Diagnostic'}
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
