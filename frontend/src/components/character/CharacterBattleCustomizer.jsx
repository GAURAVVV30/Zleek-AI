import React, { useState } from 'react';
import AvatarViewer3D from './AvatarViewer3D';
import {
  Sparkles,
  Zap,
  Shield,
  Brain,
  Flame,
  Award,
  Crown,
  Check,
  Compass,
  BookOpen,
  HelpCircle,
  Clock,
  Code2,
} from 'lucide-react';

export const CHARACTERS_ROSTER = [
  {
    id: 'robot',
    name: 'Cyber Mecha-01',
    category: 'Robot',
    tier: 'MYTHIC',
    tierColor: 'from-amber-400 to-red-500',
    badgeColor: 'bg-amber-500/20 text-amber-300 border-amber-500/40',
    url: 'https://modelviewer.dev/shared-assets/models/RobotExpressive.glb',
    powerTitle: 'Socratic Logic & Algorithmic Scaffolding',
    powerDesc: 'Precision Quiz Intervention — Intervenes with 3 progressive Socratic hints instead of answer reveals.',
    pedagogicalUseCase: 'Breaks down complex problem statements into executable pseudo-code steps during assessments and audits your code for syntax bottlenecks before grading.',
    learningBenefit: 'Builds independent algorithmic problem-solving capabilities without answer dependency.',
    aura: '#06b6d4',
    stats: {
      processing: 98,
      focus: 92,
      strategy: 96,
    },
  },
  {
    id: 'astronaut',
    name: 'Astro Vanguard',
    category: 'Astronaut',
    tier: 'LEGENDARY',
    tierColor: 'from-blue-400 to-indigo-600',
    badgeColor: 'bg-blue-500/20 text-blue-300 border-blue-500/40',
    url: 'https://modelviewer.dev/shared-assets/models/Astronaut.glb',
    powerTitle: 'Cognitive Load & Distraction Shield',
    powerDesc: 'Deep Flow Guardian — Condenses long tutorials into 3 high-impact takeaways & enforces focus timers.',
    pedagogicalUseCase: 'Automatically extracts essential mental models from long video lectures and suppresses non-essential UI elements to maintain uninterrupted focus.',
    learningBenefit: 'Maximizes working-memory retention during complex conceptual milestones.',
    aura: '#3b82f6',
    stats: {
      processing: 88,
      focus: 99,
      strategy: 90,
    },
  },
  {
    id: 'mococo',
    name: 'Abyss Assassin',
    category: 'Mococo',
    tier: 'MYTHIC',
    tierColor: 'from-purple-400 to-pink-600',
    badgeColor: 'bg-purple-500/20 text-purple-300 border-purple-500/40',
    url: 'https://darshanar190607.github.io/GLB/mococo_abyssgard.glb',
    powerTitle: 'Production Edge-Case & Architecture Hunter',
    powerDesc: 'Real-World Rigor — Injects industry production gotchas and architecture edge-cases into capstones.',
    pedagogicalUseCase: 'Scans your project submissions against enterprise production standards, highlighting scalability traps, missing error handling, and unhandled edge cases.',
    learningBenefit: 'Prepares learners for real-world enterprise engineering & technical interviews.',
    aura: '#ec4899',
    stats: {
      processing: 94,
      focus: 95,
      strategy: 98,
    },
  },
  {
    id: 'fox',
    name: 'Cyber Kitsune',
    category: 'Fox',
    tier: 'ELITE',
    tierColor: 'from-orange-400 to-amber-600',
    badgeColor: 'bg-orange-500/20 text-orange-300 border-orange-500/40',
    url: 'https://modelviewer.dev/shared-assets/models/Fox.glb',
    powerTitle: 'Dynamic Fast-Track & Prerequisite Skipping',
    powerDesc: 'Adaptive Velocity — Skips beginner fluff and dynamically re-routes paths based on diagnostic mastery.',
    pedagogicalUseCase: 'Evaluates your diagnostic answers to fast-track foundational nodes by up to 40%, ensuring you only spend time on genuine skill gaps.',
    learningBenefit: 'Dramatically shortens time-to-competence for learners with prior experience.',
    aura: '#f97316',
    stats: {
      processing: 96,
      focus: 89,
      strategy: 97,
    },
  },
  {
    id: 'horse',
    name: 'Steed of Chronos',
    category: 'Horse',
    tier: 'EPIC',
    tierColor: 'from-emerald-400 to-teal-600',
    badgeColor: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40',
    url: 'https://modelviewer.dev/shared-assets/models/Horse.glb',
    powerTitle: 'Spaced Repetition & Burnout Prevention',
    powerDesc: 'Ebbinghaus Memory Lock — Triggers automated micro-refresher quizzes 3 and 7 days post-completion.',
    pedagogicalUseCase: 'Monitors your study velocity and automatically schedules 2-minute spaced repetition flash drills to permanently lock knowledge into long-term memory.',
    learningBenefit: 'Completely eliminates knowledge decay and prevents study burnout.',
    aura: '#10b981',
    stats: {
      processing: 86,
      focus: 94,
      strategy: 91,
    },
  },
  {
    id: 'brain_stem',
    name: 'Cortex Core',
    category: 'Brain Stem',
    tier: 'ANCIENT',
    tierColor: 'from-fuchsia-400 to-purple-600',
    badgeColor: 'bg-fuchsia-500/20 text-fuchsia-300 border-fuchsia-500/40',
    url: 'https://modelviewer.dev/shared-assets/models/BrainStem.glb',
    powerTitle: 'First-Principles Blueprint & Concept Synthesis',
    powerDesc: 'Under-The-Hood Mastery — Connects disparate milestones (e.g. Pandas ➔ Linear Algebra ➔ ML loss functions).',
    pedagogicalUseCase: 'Forces mental-model understanding by explaining how low-level memory layout, hardware execution, and mathematical formulas dictate higher-level framework APIs.',
    learningBenefit: 'Builds deep, senior-level conceptual mastery that transcends specific libraries.',
    aura: '#a855f7',
    stats: {
      processing: 99,
      focus: 96,
      strategy: 99,
    },
  },
];

export const SKINS_PALETTE = [
  { id: 'default', name: 'Original', hex: null, class: 'bg-slate-700' },
  { id: 'cyber_cyan', name: 'Cyber Cyan', hex: '#06b6d4', class: 'bg-cyan-500' },
  { id: 'crimson_flare', name: 'Crimson Flare', hex: '#ef4444', class: 'bg-red-500' },
  { id: 'toxic_emerald', name: 'Toxic Emerald', hex: '#10b981', class: 'bg-emerald-500' },
  { id: 'hyper_gold', name: 'Hyper Gold', hex: '#f59e0b', class: 'bg-amber-500' },
  { id: 'neon_violet', name: 'Neon Violet', hex: '#8b5cf6', class: 'bg-purple-500' },
];

export default function CharacterBattleCustomizer({ onSelectAvatar, initialAvatar = 'robot' }) {
  const [selectedCharacter, setSelectedCharacter] = useState(
    CHARACTERS_ROSTER.find((c) => c.id === initialAvatar) || CHARACTERS_ROSTER[0]
  );
  const [customColor, setCustomColor] = useState(null);

  const handleSelect = (char) => {
    setSelectedCharacter(char);
    if (onSelectAvatar) {
      onSelectAvatar(char, customColor);
    }
  };

  const handleColorChange = (hex) => {
    setCustomColor(hex);
    if (onSelectAvatar) {
      onSelectAvatar(selectedCharacter, hex);
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 sm:p-8 text-white shadow-2xl relative overflow-hidden">
      {/* Ambient Cyber Grid Background */}
      <div className="absolute inset-0 bg-[radial-gradient(#1e293b_1px,transparent_1px)] [background-size:16px_16px] opacity-40 pointer-events-none" />

      {/* Header Banner */}
      <div className="relative z-10 flex flex-col sm:flex-row items-start sm:items-center justify-between pb-6 border-b border-slate-800 gap-4">
        <div>
          <div className="flex items-center gap-2">
            <Crown className="w-5 h-5 text-amber-400" />
            <span className="text-xs font-mono font-bold tracking-widest text-amber-400 uppercase">
              Pedagogical 3D Study Companion System
            </span>
          </div>
          <h2 className="font-display text-2xl font-extrabold tracking-tight mt-1 text-white">
            Select Your AI Learning Companion
          </h2>
          <p className="text-xs text-slate-400 mt-0.5 max-w-2xl">
            Each 3D avatar provides a specific cognitive superpower to enhance your diagnostic speed, shield your focus, and accelerate roadmap mastery.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className={`px-3 py-1 rounded-full text-xs font-extrabold uppercase border tracking-wider ${selectedCharacter.badgeColor}`}>
            ★ {selectedCharacter.tier} COMPANION
          </span>
        </div>
      </div>

      {/* Main Split View: 3D Stage on Right, Tactical Customizer on Left */}
      <div className="relative z-10 grid grid-cols-1 lg:grid-cols-12 gap-8 items-center pt-6">
        {/* Left Column: Character Cards, Stats, Pedagogical Use Case, and Skin Customization */}
        <div className="lg:col-span-6 space-y-5">
          {/* Character Roster Grid */}
          <div>
            <label className="block text-xs font-mono text-slate-400 uppercase tracking-wider mb-2.5 flex items-center gap-1.5">
              <Compass className="w-4 h-4 text-cyan-400" /> 1. Select Companion Role
            </label>

            <div className="grid grid-cols-3 sm:grid-cols-3 gap-2.5">
              {CHARACTERS_ROSTER.map((char) => {
                const isSelected = selectedCharacter.id === char.id;
                return (
                  <button
                    key={char.id}
                    type="button"
                    onClick={() => handleSelect(char)}
                    className={`p-3 rounded-2xl border text-left transition transform duration-150 ${
                      isSelected
                        ? 'border-cyan-400 bg-cyan-950/50 shadow-lg shadow-cyan-500/20 scale-[1.02] ring-1 ring-cyan-400'
                        : 'border-slate-800 bg-slate-800/40 hover:bg-slate-800 hover:border-slate-700'
                    }`}
                  >
                    <p className="text-xs font-bold text-white truncate">{char.category}</p>
                    <span className="text-[10px] font-mono text-cyan-400 block mt-0.5">{char.tier}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Detailed Pedagogical Learning Use Case Card */}
          <div className="p-4 bg-slate-800/80 border border-slate-700 rounded-2xl space-y-2 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-cyan-500/10 rounded-full blur-2xl pointer-events-none" />
            <div className="flex items-center gap-2 text-cyan-400 font-bold text-xs">
              <Zap className="w-4 h-4" />
              <span>COGNITIVE ROLE: {selectedCharacter.powerTitle}</span>
            </div>
            <p className="text-xs text-slate-200 leading-relaxed font-medium">
              {selectedCharacter.pedagogicalUseCase}
            </p>
            <div className="pt-2 border-t border-slate-700/60 flex items-start gap-2 text-[11px] text-emerald-400">
              <Check className="w-3.5 h-3.5 mt-0.5 shrink-0" />
              <span><strong>Mastery Outcome:</strong> {selectedCharacter.learningBenefit}</span>
            </div>
          </div>

          {/* Tactical Learning Aptitude Metrics */}
          <div className="space-y-2">
            <label className="block text-xs font-mono text-slate-400 uppercase tracking-wider">
              Aptitude Alignment
            </label>

            <div className="space-y-2 text-xs">
              <div>
                <div className="flex justify-between text-[11px] mb-1 font-mono">
                  <span className="text-slate-300 flex items-center gap-1.5"><Zap className="w-3 h-3 text-cyan-400" /> Neural Processing Speed</span>
                  <span className="text-cyan-400 font-bold">{selectedCharacter.stats.processing}%</span>
                </div>
                <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-cyan-500 to-blue-500 rounded-full transition-all duration-500"
                    style={{ width: `${selectedCharacter.stats.processing}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between text-[11px] mb-1 font-mono">
                  <span className="text-slate-300 flex items-center gap-1.5"><Shield className="w-3 h-3 text-indigo-400" /> Focus Distortion Shield</span>
                  <span className="text-indigo-400 font-bold">{selectedCharacter.stats.focus}%</span>
                </div>
                <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 rounded-full transition-all duration-500"
                    style={{ width: `${selectedCharacter.stats.focus}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between text-[11px] mb-1 font-mono">
                  <span className="text-slate-300 flex items-center gap-1.5"><Brain className="w-3 h-3 text-pink-400" /> Architectural Problem Mastery</span>
                  <span className="text-pink-400 font-bold">{selectedCharacter.stats.strategy}%</span>
                </div>
                <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-pink-500 to-rose-500 rounded-full transition-all duration-500"
                    style={{ width: `${selectedCharacter.stats.strategy}%` }}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Skin Armor Color Palette */}
          <div>
            <label className="block text-xs font-mono text-slate-400 uppercase tracking-wider mb-2 flex items-center gap-1.5">
              <Sparkles className="w-4 h-4 text-amber-400" /> 2. Custom Armor Skin Coating
            </label>

            <div className="flex items-center gap-3">
              {SKINS_PALETTE.map((skin) => {
                const isSelected = customColor === skin.hex;
                return (
                  <button
                    key={skin.id}
                    type="button"
                    title={skin.name}
                    onClick={() => handleColorChange(skin.hex)}
                    className={`w-9 h-9 rounded-xl ${skin.class} border-2 flex items-center justify-center transition transform hover:scale-110 shadow-md ${
                      isSelected ? 'border-white ring-2 ring-cyan-400 scale-110' : 'border-slate-700'
                    }`}
                  >
                    {isSelected && <Check className="w-4 h-4 text-white drop-shadow" />}
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right Column: High-Impact 3D Hologram Battle Stage */}
        <div className="lg:col-span-6 flex flex-col items-center justify-center relative">
          <div className="w-full h-[420px] sm:h-[480px] bg-gradient-to-b from-slate-950/80 via-slate-900/60 to-slate-950/90 rounded-3xl border border-slate-800/80 shadow-[0_0_50px_rgba(6,182,212,0.15)] relative overflow-hidden flex items-center justify-center">
            {/* Hologram Stage Lighting Overlay */}
            <div className="absolute top-4 left-6 z-20">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
                <span className="text-[10px] font-mono font-bold text-emerald-400 tracking-wider">
                  HOLOGRAPHIC COMPANION ACTIVE
                </span>
              </div>
              <h4 className="text-lg font-extrabold text-white mt-0.5">
                {selectedCharacter.name}
              </h4>
            </div>

            {/* 3D WebGL Canvas */}
            <AvatarViewer3D
              character={selectedCharacter}
              customColor={customColor}
              auraColor={selectedCharacter.aura}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
