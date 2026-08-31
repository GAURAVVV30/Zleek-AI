import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Compass, ArrowLeft } from 'lucide-react';

export default function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-black/20 backdrop-blur-sm flex flex-col items-center justify-center p-6 text-center">
      <div className="w-16 h-16 rounded-3xl bg-indigo-900/40 backdrop-blur-sm text-indigo-400 flex items-center justify-center mb-4">
        <Compass className="w-8 h-8" />
      </div>
      <h1 className="font-display text-4xl font-extrabold text-white mb-2">404</h1>
      <p className="text-sm text-slate-300 mb-6 max-w-sm">
        The learning pathway or page you are looking for does not exist.
      </p>
      <button
        onClick={() => navigate('/')}
        className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl text-xs shadow-[0_0_15px_rgba(79,70,229,0.2)] transition flex items-center gap-2"
      >
        <ArrowLeft className="w-4 h-4" /> Return to Home
      </button>
    </div>
  );
}
