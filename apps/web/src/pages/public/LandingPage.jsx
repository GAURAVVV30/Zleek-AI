import React from 'react';
import { RandomLetterSwap } from '../../components/ui/random-letter-swap';
import { useNavigate } from 'react-router-dom';
import { motion, useScroll, useTransform } from 'framer-motion';
import {
  Compass,
  ArrowRight,
  Target,
  ClipboardList,
  GraduationCap,
  Star,
  ShieldCheck,
  Zap,
  Award,
} from 'lucide-react';

const FADE_UP = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.8, ease: [0.16, 1, 0.3, 1] } }
};

const STAGGER = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.1 } }
};

export default function LandingPage() {
  const navigate = useNavigate();

  return (
    <div className="w-full overflow-x-hidden">
      
      {/* Animated Subtle Background */}
      <motion.div 
        className="fixed inset-0 z-0 pointer-events-none"
      >
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-indigo-900/15 via-[#050505]/0 to-[#050505]/0"></div>
      </motion.div>

      {/* Header */}
      <motion.header 
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        className="w-full bg-black/40 backdrop-blur-md border-b border-white/5 sticky top-0 z-50"
      >
        <div className="max-w-7xl mx-auto px-6 md:px-12 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3 cursor-pointer group" onClick={() => navigate('/')}>
            <div className="w-8 h-8 rounded-xl bg-white/10 flex items-center justify-center text-white font-bold transition-transform group-hover:scale-105">
              <Compass className="w-4 h-4" />
            </div>
            <span className="font-display font-bold text-lg tracking-tight text-white">
              Amplified<span className="text-indigo-400">.AI</span>
            </span>
          </div>

          <nav className="hidden md:flex items-center gap-10 text-sm font-medium text-slate-400">
            <RandomLetterSwap href="#features" className="hover:text-white transition-colors inline-block w-[64px]" label="Features" />
            <RandomLetterSwap href="#how-it-works" className="hover:text-white transition-colors inline-block w-[88px]" label="How it works" />
            <RandomLetterSwap href="#about" className="hover:text-white transition-colors inline-block w-[40px]" label="About" />
          </nav>

          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/login')}
              className="text-sm font-semibold text-slate-300 hover:text-white transition-colors"
            >
              Login
            </button>
            <button
              onClick={() => navigate('/signup')}
              className="px-5 py-2 rounded-full text-sm font-semibold text-[#050505] bg-white hover:bg-slate-200 transition-transform transform hover:-translate-y-0.5"
            >
              Sign Up
            </button>
          </div>
        </div>
      </motion.header>

      <main className="relative z-10">
        {/* HERO SECTION */}
        <section className="w-full min-h-[90vh] flex flex-col justify-center py-20">
          <div className="container mx-auto px-6 md:px-12 max-w-7xl">
            <motion.div 
              initial="hidden" 
              animate="visible" 
              variants={STAGGER}
              className="max-w-4xl space-y-8"
            >
              <motion.div variants={FADE_UP} className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-white/10 text-slate-300 text-xs font-semibold tracking-wide uppercase bg-white/5">
                <span className="w-2 h-2 rounded-full bg-indigo-500 animate-pulse"></span>
                Evidence-Backed Adaptive Intelligence
              </motion.div>

              <motion.h1 variants={FADE_UP} className="font-display text-5xl md:text-7xl lg:text-[5.5rem] font-bold text-white tracking-tighter leading-[1.05]">
                AI that builds your <br className="hidden md:block"/> personalized learning path.
              </motion.h1>

              <motion.p variants={FADE_UP} className="text-lg md:text-xl text-slate-400 leading-relaxed max-w-2xl font-light">
                Adaptive learning that focuses on what you can actually do. We measure progress by proven competence, not just clicks.
              </motion.p>

              <motion.div variants={FADE_UP} className="flex flex-wrap items-center gap-4 pt-4">
                <button
                  onClick={() => navigate('/signup')}
                  className="px-8 py-4 bg-white text-[#050505] rounded-full font-semibold text-sm hover:scale-105 transition-transform flex items-center gap-2"
                >
                  <span>Start Your Diagnostic</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
                <a
                  href="#how-it-works"
                  className="px-8 py-4 bg-transparent border border-white/20 hover:border-white/40 text-white rounded-full font-semibold text-sm transition-colors"
                >
                  Discover How
                </a>
              </motion.div>
            </motion.div>
          </div>
        </section>

        {/* HOW IT WORKS SECTION */}
        <section id="how-it-works" className="w-full py-32">
          <div className="container mx-auto px-6 md:px-12 max-w-5xl">
            <motion.div 
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, margin: "-100px" }}
              variants={FADE_UP}
              className="mb-24"
            >
              <h2 className="font-display text-4xl md:text-5xl font-bold tracking-tight mb-4">The Competency Engine</h2>
              <p className="text-slate-400 text-lg max-w-xl">A completely dynamic learning loop adjusting to your proven skills in real-time.</p>
            </motion.div>

            <div className="space-y-32">
              {[
                { step: '01', title: 'Define Goal & Bond', desc: 'State your target role and pair with a 3D cognitive AI companion tailored to your learning style.', icon: Target },
                { step: '02', title: 'Baseline Diagnostic', desc: 'Identify existing strengths and gaps so you never waste time re-learning concepts you already know.', icon: ClipboardList },
                { step: '03', title: 'Prove Understanding', desc: 'Engage with best-of-web curated resources, then unlock the next node by passing quizzes or applied projects.', icon: Star },
                { step: '04', title: 'Verified Mastery', desc: 'Every milestone achieved is backed by immutable evidence, building a portfolio of proven competence.', icon: GraduationCap },
              ].map((item, i) => (
                <motion.div 
                  key={i}
                  initial={{ opacity: 0, y: 100 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, margin: "-20%" }}
                  transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
                  className="flex flex-col md:flex-row gap-8 md:gap-16 items-start"
                >
                  <div className="text-6xl md:text-8xl font-bold text-white/5 font-display shrink-0 tracking-tighter">
                    {item.step}
                  </div>
                  <div className="space-y-4 pt-2 md:pt-6">
                    <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center text-white mb-6">
                      <item.icon className="w-5 h-5" />
                    </div>
                    <h3 className="text-2xl md:text-3xl font-bold tracking-tight">{item.title}</h3>
                    <p className="text-slate-400 text-lg leading-relaxed max-w-lg">{item.desc}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </section>

        {/* FEATURES SECTION */}
        <section id="features" className="w-full py-32">
          <div className="container mx-auto px-6 md:px-12 max-w-7xl">
            <motion.div 
              initial="hidden"
              whileInView="visible"
              viewport={{ once: true, margin: "-100px" }}
              variants={FADE_UP}
              className="mb-24 md:text-center"
            >
              <h2 className="font-display text-4xl md:text-5xl font-bold tracking-tight mb-4">Why Amplified.AI Is Different</h2>
              <p className="text-slate-400 text-lg md:mx-auto max-w-xl">Built on strict pedagogical principles: expert curation, evidence over self-report, and continuous path adaptation.</p>
            </motion.div>

            <div className="grid grid-cols-1 md:grid-cols-12 gap-6 md:gap-12">
              <motion.div 
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.8 }}
                className="md:col-span-8 bg-black/40 backdrop-blur-md rounded-3xl p-8 md:p-12 border border-white/5 group hover:border-white/10 transition-colors"
              >
                <ShieldCheck className="w-8 h-8 text-white mb-8" />
                <h3 className="text-3xl font-bold mb-4 tracking-tight">Expert Knowledge Graphs</h3>
                <p className="text-slate-400 text-lg leading-relaxed max-w-xl">
                  Curricula originate from vetted taxonomies and verified prerequisite relationships — never uncontrolled LLM hallucinations. Precision learning paths backed by real industry standards.
                </p>
              </motion.div>

              <motion.div 
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.8, delay: 0.1 }}
                className="md:col-span-4 bg-black/40 backdrop-blur-md rounded-3xl p-8 md:p-12 border border-white/5 group hover:border-white/10 transition-colors flex flex-col justify-between"
              >
                <Zap className="w-8 h-8 text-white mb-8" />
                <div>
                  <h3 className="text-2xl font-bold mb-4 tracking-tight">Adaptive Remediation</h3>
                  <p className="text-slate-400 text-base leading-relaxed">
                    Struggling on a specific topic? Targeted remediation resources are seamlessly inserted into your path before moving forward.
                  </p>
                </div>
              </motion.div>

              <motion.div 
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.8, delay: 0.2 }}
                className="md:col-span-12 bg-black/40 backdrop-blur-md rounded-3xl p-8 md:p-16 border border-white/5 group hover:border-white/10 transition-colors flex flex-col md:flex-row gap-8 md:gap-16 items-center"
              >
                <div className="shrink-0">
                  <Award className="w-12 h-12 text-white" />
                </div>
                <div>
                  <h3 className="text-3xl font-bold mb-4 tracking-tight">Transparent Explainability</h3>
                  <p className="text-slate-400 text-lg leading-relaxed max-w-3xl">
                    Every concept and resource comes with clear "Why am I learning this?" reasoning tied directly to your career goal. No more blindly following courses; understand the exact utility of what you are mastering.
                  </p>
                </div>
              </motion.div>
            </div>
          </div>
        </section>

        {/* ABOUT SECTION */}
        <section id="about" className="w-full py-32">
          <div className="container mx-auto px-6 md:px-12 max-w-4xl text-center">
            <motion.div 
              initial={{ opacity: 0, y: 40 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 1 }}
            >
              <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight mb-8">
                We believe learning should be proven, not just watched.
              </h2>
              <p className="text-slate-400 text-lg md:text-xl leading-relaxed">
                Amplified.AI was built to bridge the gap between passive video consumption and active competence. Our mission is to provide the most direct, evidence-backed route to professional mastery.
              </p>
            </motion.div>
          </div>
        </section>

        {/* FINAL CTA */}
        <section className="w-full py-24">
          <div className="container mx-auto px-6 md:px-12 max-w-4xl">
            <motion.div 
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
              className="bg-black/50 backdrop-blur-xl rounded-[2rem] p-10 md:p-16 border border-white/10 text-center relative overflow-hidden"
            >
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent pointer-events-none"></div>
              <div className="relative z-10">
                <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight mb-6">
                  Build your path. <br/> Close your gaps.
                </h2>
                <button
                  onClick={() => navigate('/signup')}
                  className="px-8 py-4 bg-white text-[#050505] hover:scale-105 transition-transform rounded-full font-bold text-sm shadow-2xl shadow-white/10 flex items-center gap-3 mx-auto"
                >
                  <span>Start Your Diagnostic</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </motion.div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="w-full py-12 border-t border-white/5 text-center">
        <p className="text-slate-500 text-sm">© 2026 Amplified.AI Learning Platform. Built for verified competence.</p>
      </footer>
    </div>
  );
}
