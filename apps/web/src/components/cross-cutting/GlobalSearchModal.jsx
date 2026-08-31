import React, { useState, useEffect } from 'react';
import { Search, X, BookOpen, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function GlobalSearchModal({ isOpen, onClose }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    if (!isOpen) return;
    const handler = (e) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [isOpen, onClose]);

  useEffect(() => {
    if (query.trim().length > 1) {
      apiClient
        .get(`${ENDPOINTS.SEARCH}?q=${encodeURIComponent(query)}`)
        .then((res) => setResults(res.data || []))
        .catch(() => setResults([]));
    } else {
      setResults([]);
    }
  }, [query]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-20 px-4">
      <div className="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity" onClick={onClose}></div>
      <div className="relative w-full max-w-xl bg-black/40 backdrop-blur-xl rounded-2xl shadow-elevated border border-white/10 overflow-hidden z-10 animate-in fade-in zoom-in-95 duration-150">
        <div className="p-4 border-b border-white/5 flex items-center gap-3">
          <Search className="w-5 h-5 text-slate-400" />
          <input
            type="text"
            placeholder="Search concepts, resources, topics... (e.g. Pandas, ML)"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
            className="flex-1 text-sm bg-transparent border-none outline-none text-white placeholder:text-slate-400"
          />
          {query && (
            <button onClick={() => setQuery('')} className="p-1 text-slate-400 hover:text-slate-300">
              <X className="w-4 h-4" />
            </button>
          )}
          <kbd className="px-2 py-0.5 text-[10px] font-semibold text-slate-400 bg-slate-100 rounded border border-white/10">
            ESC
          </kbd>
        </div>

        <div className="max-h-72 overflow-y-auto p-2">
          {query.trim().length === 0 ? (
            <div className="p-6 text-center text-xs text-slate-400">
              Type keywords to search across your personalized roadmap and curated library.
            </div>
          ) : results.length === 0 ? (
            <div className="p-6 text-center text-xs text-slate-400">
              No matching concepts or resources found for "{query}".
            </div>
          ) : (
            results.map((item, idx) => (
              <button
                key={idx}
                onClick={() => {
                  onClose();
                  navigate(item.link);
                }}
                className="w-full p-3 rounded-xl hover:bg-indigo-900/40 backdrop-blur-sm/60 transition flex items-center justify-between text-left group"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-indigo-900/60 backdrop-blur-md text-indigo-400 flex items-center justify-center">
                    <BookOpen className="w-4 h-4" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white group-hover:text-indigo-400 transition">
                      {item.title}
                    </p>
                    <p className="text-xs text-slate-400">{item.match}</p>
                  </div>
                </div>
                <ArrowRight className="w-4 h-4 text-slate-300 group-hover:text-indigo-400 group-hover:translate-x-1 transition" />
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
