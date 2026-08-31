import React, { useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { ArrowRight, Lock, Mail, User, Eye, EyeOff, Github, Chrome } from 'lucide-react';

export default function SignUpPage() {
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [acceptedTerms, setAcceptedTerms] = useState(false);
  
  const { signup, isLoading } = useAuth();
  const { addToast } = useToast();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!fullName || !email || !password || !confirmPassword) {
      addToast('Please fill in all required fields.', 'error');
      return;
    }
    if (password !== confirmPassword) {
      addToast('Passwords do not match.', 'error');
      return;
    }
    if (!acceptedTerms) {
      addToast('You must accept the Terms of Service to continue.', 'error');
      return;
    }
    
    try {
      await signup(email, password, fullName);
      addToast('Account created successfully!', 'success');
      navigate('/onboarding/goal');
    } catch (err) {
      addToast(err?.message || 'Registration failed', 'error');
    }
  };

  return (
    <div>
      <div className="text-center mb-6">
        <h2 className="font-display text-2xl font-bold text-white">Create your account</h2>
        <p className="text-xs text-slate-400 mt-1">
          Already have an account?{' '}
          <NavLink to="/login" className="text-indigo-400 font-semibold hover:underline">
            Sign in
          </NavLink>
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-semibold text-white mb-1">Full Name</label>
          <div className="relative">
            <User className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
            <input
              type="text"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              className="w-full pl-9 pr-3 py-2 text-xs bg-black/30 text-white border border-white/10 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none placeholder-slate-500 transition"
              placeholder="Full Name"
            />
          </div>
        </div>

        <div>
          <label className="block text-xs font-semibold text-white mb-1">Email Address</label>
          <div className="relative">
            <Mail className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full pl-9 pr-3 py-2 text-xs bg-black/30 text-white border border-white/10 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none placeholder-slate-500 transition"
              placeholder="Email Address"
            />
          </div>
        </div>

        <div>
          <label className="block text-xs font-semibold text-white mb-1">Password</label>
          <div className="relative">
            <Lock className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
            <input
              type={showPassword ? 'text' : 'password'}
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full pl-9 pr-10 py-2 text-xs bg-black/30 text-white border border-white/10 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none placeholder-slate-500 transition"
              placeholder="••••••••"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-2.5 text-slate-400 hover:text-white transition"
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>
        
        <div>
          <label className="block text-xs font-semibold text-white mb-1">Confirm Password</label>
          <div className="relative">
            <Lock className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
            <input
              type={showConfirmPassword ? 'text' : 'password'}
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full pl-9 pr-10 py-2 text-xs bg-black/30 text-white border border-white/10 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none placeholder-slate-500 transition"
              placeholder="••••••••"
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-3 top-2.5 text-slate-400 hover:text-white transition"
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
            >
              {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            </button>
          </div>
        </div>

        <div className="mt-2">
          <label className="flex items-center gap-2 cursor-pointer group">
            <input 
              type="checkbox" 
              checked={acceptedTerms}
              onChange={(e) => setAcceptedTerms(e.target.checked)}
              className="w-3.5 h-3.5 rounded border-white/10 bg-black/30 text-indigo-500 focus:ring-indigo-500/50 cursor-pointer accent-indigo-500" 
            />
            <span className="text-[11px] font-medium text-slate-300 group-hover:text-white transition">
              I agree to the Terms of Service and Privacy Policy
            </span>
          </label>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="w-full py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-bold rounded-xl text-xs shadow-[0_0_20px_rgba(79,70,229,0.3)] transition flex items-center justify-center gap-2 mt-4"
        >
          {isLoading ? 'Creating account...' : 'Create Account'}
          {!isLoading && <ArrowRight className="w-4 h-4" />}
        </button>
      </form>

      <div className="mt-6 flex items-center gap-3">
        <div className="flex-1 h-px bg-white/10"></div>
        <span className="text-[10px] uppercase font-bold tracking-wider text-slate-500">or</span>
        <div className="flex-1 h-px bg-white/10"></div>
      </div>

      <div className="mt-6 space-y-3">
        <button type="button" className="w-full flex items-center justify-center gap-2 py-2.5 bg-black/30 hover:bg-white/10 border border-white/10 rounded-xl text-xs font-semibold text-white transition">
          <Chrome className="w-4 h-4" />
          Continue with Google
        </button>
        <button type="button" className="w-full flex items-center justify-center gap-2 py-2.5 bg-black/30 hover:bg-white/10 border border-white/10 rounded-xl text-xs font-semibold text-white transition">
          <Github className="w-4 h-4" />
          Continue with GitHub
        </button>
      </div>
    </div>
  );
}
