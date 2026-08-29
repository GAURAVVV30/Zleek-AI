import React, { useState } from 'react';
import { NavLink } from 'react-router-dom';
import { Mail, ArrowLeft } from 'lucide-react';
import { useToast } from '../../context/ToastContext';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const { addToast } = useToast();

  const handleSubmit = (e) => {
    e.preventDefault();
    addToast('Password reset link sent to your email address.', 'info');
  };

  return (
    <div>
      <div className="text-center mb-6">
        <h2 className="font-display text-2xl font-bold text-slate-900">Reset Password</h2>
        <p className="text-xs text-slate-500 mt-1">Enter your registered email to receive reset instructions</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-xs font-semibold text-slate-700 mb-1">Email</label>
          <div className="relative">
            <Mail className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full pl-9 pr-3 py-2 text-xs border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-600 outline-none"
              placeholder="you@example.com"
            />
          </div>
        </div>

        <button
          type="submit"
          className="w-full py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-xl text-xs shadow-md transition"
        >
          Send Reset Link
        </button>
      </form>

      <div className="mt-6 pt-4 border-t border-slate-100 text-center">
        <NavLink to="/login" className="inline-flex items-center gap-1 text-xs text-slate-600 hover:text-blue-600">
          <ArrowLeft className="w-3.5 h-3.5" /> Back to Log In
        </NavLink>
      </div>
    </div>
  );
}
