import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Compass,
  ArrowRight,
  Target,
  ClipboardList,
  GraduationCap,
  Star,
  RefreshCw,
  ShieldCheck,
  Zap,
  Award,
  Bot,
  Brain,
  Sparkles,
} from 'lucide-react';

export default function LandingPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-[#edf4fe] text-slate-900 overflow-x-hidden font-sans">
      {/* 1. Top Navigation Bar - Full Width with Container */}
      <header className="w-full bg-[#edf4fe] border-b border-blue-100/50 sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-8 py-4 flex items-center justify-between">
          {/* Brand */}
          <div className="flex items-center gap-3 cursor-pointer" onClick={() => navigate('/')}>
            <div className="w-10 h-10 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white font-bold shadow-md shadow-blue-500/20">
              <Compass className="w-5 h-5" />
            </div>
            <span className="font-display font-bold text-xl tracking-tight text-slate-900">
              Amplified<span className="text-blue-600">.AI</span>
            </span>
          </div>

          {/* Center Links */}
          <nav className="hidden md:flex items-center gap-8 text-sm font-medium text-slate-600">
            <a href="#features" className="hover:text-blue-600 transition">
              Features
            </a>
            <a href="#how-it-works" className="hover:text-blue-600 transition font-semibold text-slate-900">
              How it works
            </a>
            <a href="#companions" className="hover:text-blue-600 transition">
              3D AI Companions
            </a>
            <a href="#about" className="hover:text-blue-600 transition">
              About
            </a>
          </nav>

          {/* Auth Buttons */}
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate('/login')}
              className="px-5 py-2 rounded-xl text-sm font-semibold text-slate-700 bg-white/80 hover:bg-white border border-slate-200 shadow-sm transition"
            >
              Login
            </button>
            <button
              onClick={() => navigate('/signup')}
              className="px-5 py-2 rounded-xl text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 shadow-md shadow-blue-500/25 transition transform hover:-translate-y-0.5"
            >
              Sign Up
            </button>
          </div>
        </div>
      </header>

      {/* 2. Full-Bleed Seamless Hero Section (Zero Side Spaces, 100% Responsive) */}
      <section className="w-full bg-[#edf4fe] relative overflow-hidden pt-4 pb-12 sm:pt-8 sm:pb-16 lg:pb-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-8 lg:px-8 grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
          {/* Left Column: Heading & CTAs */}
          <div className="lg:col-span-5 z-20 space-y-6 pt-4 lg:pt-0">
            <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-blue-100/60 border border-blue-200/80 text-blue-800 text-xs font-semibold">
              <Sparkles className="w-3.5 h-3.5 text-blue-600" />
              Evidence-Backed Adaptive Intelligence
            </div>

            <h1 className="font-display text-4xl sm:text-5xl lg:text-[3.5rem] font-extrabold text-slate-900 tracking-tight leading-[1.1]">
              Your path to <br />
              real competence.
            </h1>

            <p className="text-base sm:text-lg text-slate-600 leading-relaxed max-w-md font-normal">
              Adaptive learning that focuses on what you can <strong>actually do</strong> — paired with a personalized 3D AI study companion.
            </p>

            <div className="flex flex-wrap items-center gap-4 pt-2">
              <button
                onClick={() => navigate('/signup')}
                className="px-8 py-3.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-bold text-sm shadow-lg shadow-blue-500/25 transition transform hover:-translate-y-0.5 flex items-center gap-2"
              >
                <span>Get Started</span>
                <ArrowRight className="w-4 h-4" />
              </button>
              <a
                href="#how-it-works"
                className="px-6 py-3.5 bg-white hover:bg-slate-50 text-slate-800 border border-slate-200 rounded-xl font-bold text-sm shadow-sm transition hover:border-slate-300"
              >
                Learn More
              </a>
            </div>
          </div>

          {/* Right Column: Full-Bleed Mountain Canvas (Zero Box, Zero Side Margins) */}
          <div className="lg:col-span-7 relative flex items-center justify-center lg:justify-end -mr-4 sm:-mr-8 lg:-mr-12">
            <div className="relative w-full max-w-2xl lg:max-w-none">
              <img
                src="/illustrations/hero-mountain.png"
                alt="Adaptive learning path to mastery"
                className="w-full h-auto object-contain select-none pointer-events-none"
                style={{
                  // Perfectly dissolve the left edge into the pale blue background
                  maskImage: 'linear-gradient(to right, transparent 0%, rgba(0,0,0,0.5) 5%, black 20%)',
                  WebkitMaskImage: 'linear-gradient(to right, transparent 0%, rgba(0,0,0,0.5) 5%, black 20%)',
                }}
                onError={(e) => {
                  e.target.src = '/illustrations/media_1787759050661.jpg';
                }}
              />
            </div>
          </div>
        </div>
      </section>

      {/* 3. Pedagogical 3D AI Companions Showcase Section */}
      <section id="companions" className="w-full bg-gradient-to-b from-[#edf4fe] via-white to-white py-20 border-t border-blue-100/60">
        <div className="max-w-7xl mx-auto px-4 sm:px-8">
          <div className="text-center max-w-3xl mx-auto mb-16 space-y-3">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-purple-50 text-purple-700 text-xs font-bold uppercase tracking-wider">
              <Bot className="w-4 h-4 text-purple-600" />
              Pedagogical AI Learning Personas
            </div>
            <h2 className="font-display text-3xl sm:text-4xl font-extrabold text-slate-900">
              Why 3D AI Study Companions?
            </h2>
            <p className="text-slate-600 text-sm leading-relaxed">
              Our 3D Companions aren’t just avatars — each is an <strong>active cognitive assistant</strong> engineered with specific pedagogical mechanisms to accelerate learning retention, shield focus, and prevent burnout.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Robot */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold text-xl">
                🤖
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Cyber Mecha-01 (Robot)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-amber-50 text-amber-700 rounded-full">Socratic Logic</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Algorithmic Scaffolding & Quiz Guidance</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                When you get stuck on a diagnostic or assessment question, the Robot intervenes with <strong>step-by-step Socratic hints</strong> instead of giving away answers, training first-principles problem-solving.
              </p>
            </div>

            {/* Astronaut */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-xl">
                🧑‍🚀
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Astro Vanguard (Astronaut)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-blue-50 text-blue-700 rounded-full">Deep Focus Shield</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Cognitive Load & Flow States</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Condenses 40-minute long video tutorials into <strong>3 high-impact takeaways</strong> and suppresses non-essential platform noise to keep you in an uninterrupted flow state.
              </p>
            </div>

            {/* Mococo */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-purple-50 text-purple-600 flex items-center justify-center font-bold text-xl">
                🐺
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Abyss Assassin (Mococo)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-purple-50 text-purple-700 rounded-full">Architecture Scout</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Production Code & Edge-Cases</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Injects real-world <strong>production pitfalls and architecture patterns</strong> into Capstone Project requirements, preparing you for real-world enterprise engineering interviews.
              </p>
            </div>

            {/* Fox */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-orange-50 text-orange-600 flex items-center justify-center font-bold text-xl">
                🦊
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Cyber Kitsune (Fox)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-orange-50 text-orange-700 rounded-full">Fast-Track Navigator</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Dynamic Prerequisite Skipping</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Analyzes your diagnostic responses to <strong>fast-track roadmap milestones</strong> by up to 40%, automatically bypassing basic syntax drills if you demonstrate existing mastery.
              </p>
            </div>

            {/* Horse */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-xl">
                🐴
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Steed of Chronos (Horse)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-emerald-50 text-emerald-700 rounded-full">Spaced Repetition</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Memory Retention & Burnout Guard</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Calculates your Ebbinghaus forgetting curve and triggers <strong>2-minute refresher micro-quizzes</strong> 3 days and 7 days after milestone completion to lock in permanent retention.
              </p>
            </div>

            {/* Brain Stem */}
            <div className="bg-white border border-slate-200/80 rounded-3xl p-6 shadow-sm hover:shadow-md transition space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-fuchsia-50 text-fuchsia-600 flex items-center justify-center font-bold text-xl">
                🧠
              </div>
              <div>
                <div className="flex items-center justify-between">
                  <h3 className="font-bold text-base text-slate-900">Cortex Core (Brain Stem)</h3>
                  <span className="text-[10px] font-bold px-2 py-0.5 bg-fuchsia-50 text-fuchsia-700 rounded-full">First-Principles</span>
                </div>
                <p className="text-xs font-semibold text-blue-600 mt-0.5">Use Case: Deep Mental Model Synthesis</p>
              </div>
              <p className="text-xs text-slate-600 leading-relaxed">
                Connects disparate milestones (e.g. Pandas DataFrames ➔ Linear Algebra Matrices ➔ ML Loss Functions) with <strong>root architectural mental models</strong>.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 4. The 4-Step Adaptive Loop Section */}
      <section id="how-it-works" className="max-w-7xl mx-auto px-4 sm:px-8 py-20 border-t border-slate-200/60 bg-white">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-50 text-blue-700 text-xs font-semibold mb-3">
            <RefreshCw className="w-3.5 h-3.5" />
            The Competency Engine
          </div>
          <h2 className="font-display text-3xl font-extrabold text-slate-900">
            How Amplified.AI Adapts To You
          </h2>
          <p className="text-slate-600 text-sm mt-3 leading-relaxed">
            Most platforms measure progress by what you click. We measure progress by what you can prove, adjusting your roadmap in real-time.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Step 1 */}
          <div className="bg-slate-50/70 border border-slate-200/80 rounded-2xl p-6 shadow-sm hover:shadow-md transition">
            <div className="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-lg mb-4">
              <Target className="w-6 h-6" />
            </div>
            <span className="text-xs font-bold text-blue-600 uppercase tracking-wider">Step 01</span>
            <h3 className="text-base font-bold text-slate-900 mt-1 mb-2">Define Goal & Bond Avatar</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              State your target role and pair with a 3D cognitive AI companion tailored to your learning style.
            </p>
          </div>

          {/* Step 2 */}
          <div className="bg-slate-50/70 border border-slate-200/80 rounded-2xl p-6 shadow-sm hover:shadow-md transition">
            <div className="w-12 h-12 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold text-lg mb-4">
              <ClipboardList className="w-6 h-6" />
            </div>
            <span className="text-xs font-bold text-indigo-600 uppercase tracking-wider">Step 02</span>
            <h3 className="text-base font-bold text-slate-900 mt-1 mb-2">Baseline Diagnostic</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Identify existing strengths and gaps so you never waste time re-learning concepts you already know.
            </p>
          </div>

          {/* Step 3 */}
          <div className="bg-slate-50/70 border border-slate-200/80 rounded-2xl p-6 shadow-sm hover:shadow-md transition">
            <div className="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-lg mb-4">
              <Star className="w-6 h-6" />
            </div>
            <span className="text-xs font-bold text-blue-600 uppercase tracking-wider">Step 03</span>
            <h3 className="text-base font-bold text-slate-900 mt-1 mb-2">Prove Understanding</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Engage with best-of-web curated resources, then unlock the next node by passing quizzes or applied projects.
            </p>
          </div>

          {/* Step 4 */}
          <div className="bg-slate-50/70 border border-slate-200/80 rounded-2xl p-6 shadow-sm hover:shadow-md transition">
            <div className="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-lg mb-4">
              <GraduationCap className="w-6 h-6" />
            </div>
            <span className="text-xs font-bold text-emerald-600 uppercase tracking-wider">Step 04</span>
            <h3 className="text-base font-bold text-slate-900 mt-1 mb-2">Verified Mastery</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Every milestone achieved is backed by immutable evidence, building a portfolio of proven competence.
            </p>
          </div>
        </div>
      </section>

      {/* 5. Product Principles / Features Section */}
      <section id="features" className="max-w-7xl mx-auto px-4 sm:px-8 py-20 bg-slate-50/80 rounded-3xl border border-slate-200/60 my-10">
        <div className="max-w-3xl mx-auto text-center mb-16">
          <h2 className="font-display text-3xl font-extrabold text-slate-900">
            Why Amplified.AI Is Different
          </h2>
          <p className="text-slate-600 text-sm mt-3">
            Built on strict pedagogical principles: expert curation, evidence over self-report, and continuous path adaptation.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="p-6 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-3">
            <div className="w-10 h-10 rounded-xl bg-blue-100 text-blue-600 flex items-center justify-center">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <h3 className="font-bold text-base text-slate-900">Expert Knowledge Graphs</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Curricula originate from vetted taxonomies and verified prerequisite relationships — never uncontrolled LLM hallucinations.
            </p>
          </div>

          <div className="p-6 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-3">
            <div className="w-10 h-10 rounded-xl bg-amber-100 text-amber-600 flex items-center justify-center">
              <Zap className="w-5 h-5" />
            </div>
            <h3 className="font-bold text-base text-slate-900">Adaptive Remediation</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Struggling on a specific topic? Targeted remediation resources are seamlessly inserted into your path before moving forward.
            </p>
          </div>

          <div className="p-6 bg-white rounded-2xl border border-slate-200/80 shadow-sm space-y-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-100 text-emerald-600 flex items-center justify-center">
              <Award className="w-5 h-5" />
            </div>
            <h3 className="font-bold text-base text-slate-900">Transparent Explainability</h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              Every concept and resource comes with clear "Why am I learning this?" reasoning tied directly to your career goal.
            </p>
          </div>
        </div>
      </section>

      {/* 6. Bottom CTA Banner */}
      <section className="max-w-7xl mx-auto px-4 sm:px-8 py-16 text-center">
        <div className="bg-gradient-to-r from-blue-600 via-indigo-600 to-blue-700 rounded-3xl p-10 sm:p-14 text-white shadow-elevated relative overflow-hidden">
          <div className="relative z-10 max-w-2xl mx-auto space-y-5">
            <h2 className="font-display text-3xl sm:text-4xl font-extrabold tracking-tight">
              Ready to climb your path to mastery?
            </h2>
            <p className="text-blue-100 text-sm sm:text-base leading-relaxed">
              Define your goal in 30 seconds, choose your 3D study companion, and let our adaptive engine build your personalized journey.
            </p>
            <button
              onClick={() => navigate('/signup')}
              className="px-8 py-4 bg-white text-blue-700 hover:bg-blue-50 font-bold rounded-xl text-sm shadow-lg transition transform hover:-translate-y-0.5"
            >
              Start Free Assessment
            </button>
          </div>
        </div>
      </section>

      {/* 7. Footer */}
      <footer className="w-full bg-white py-8 border-t border-slate-200 text-center text-xs text-slate-500">
        <p>© 2026 Amplified.AI Learning Platform. Built for verified competence.</p>
      </footer>
    </div>
  );
}
