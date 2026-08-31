import React, { useState, useEffect, useRef } from 'react';
import { Search, X, BookOpen, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function AnimatedSearchBar() {

  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [isHovered, setIsHovered] = useState(false);
  const [isFocused, setIsFocused] = useState(false);
  const navigate = useNavigate();
  const inputRef = useRef(null);
  const containerRef = useRef(null);

  // Derive expanded state from hover or focus
  const isExpanded = isHovered || isFocused;

  useEffect(() => {
    const handleKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsFocused(true);
        setTimeout(() => inputRef.current?.focus(), 100);
      }
      if (e.key === 'Escape' && isExpanded) {
        setIsFocused(false);
        setIsHovered(false);
        inputRef.current?.blur();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isExpanded]);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsFocused(false);
        setIsHovered(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

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

  const handleExpand = () => {
    setIsFocused(true);
    setTimeout(() => inputRef.current?.focus(), 100);
  };

  return (
    <div 
      className="relative z-50 flex items-center" 
      ref={containerRef}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <motion.div
        initial={false}
        animate={{
          width: isExpanded ? '260px' : '42px',
          backgroundColor: isExpanded ? 'rgba(15, 23, 42, 0.65)' : 'rgba(15, 23, 42, 0.2)',
          borderColor: isExpanded ? 'rgba(129, 140, 248, 0.3)' : 'rgba(255, 255, 255, 0.1)',
        }}
        transition={{ type: 'spring', stiffness: 300, damping: 30 }}
        className="h-[42px] rounded-full flex items-center overflow-hidden border backdrop-blur-md shadow-[0_0_15px_rgba(79,70,229,0.15)] relative"
      >
        <button
          type="button"
          onClick={handleExpand}
          className="w-[42px] h-[42px] flex items-center justify-center shrink-0 text-slate-300 hover:text-white transition"
          aria-label="Search concepts and resources"
        >
          <Search className="w-[18px] h-[18px]" />
        </button>

        <input
          ref={inputRef}
          type="text"
          placeholder={isExpanded ? "Search concepts, resources..." : ""}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsFocused(true)}
          className={`flex-1 bg-transparent border-none outline-none text-sm text-white placeholder:text-slate-400 h-full pl-0 pr-10 min-w-0 transition-opacity duration-200 ${isExpanded ? 'opacity-100' : 'opacity-0'}`}
          tabIndex={isExpanded ? 0 : -1}
        />

        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute right-2 flex items-center gap-2"
            >
              {query && (
                <button
                  type="button"
                  onClick={() => setQuery('')}
                  className="p-1 text-slate-400 hover:text-white"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              )}
              {!query && (
                <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-mono font-semibold text-slate-400 bg-black/40 rounded border border-white/10 pointer-events-none">
                  ⌘K
                </kbd>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>

      {/* Dropdown Results */}
      <AnimatePresence>
        {isExpanded && query.trim().length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10 }}
            className="absolute top-12 right-0 w-[320px] sm:w-[360px] bg-slate-900/95 backdrop-blur-xl rounded-2xl shadow-[0_0_40px_rgba(0,0,0,0.5)] border border-white/10 overflow-hidden"
          >
            <div className="max-h-72 overflow-y-auto p-2">
              {query.trim().length === 0 ? null : results.length === 0 ? (
                <div className="p-6 text-center text-xs text-slate-400">
                  No matching concepts or resources found for "{query}".
                </div>
              ) : (
                results.map((item, idx) => (
                  <button
                    key={idx}
                    onClick={() => {
                      setIsFocused(false);
                      setIsHovered(false);
                      navigate(item.link);
                    }}
                    className="w-full p-3 rounded-xl hover:bg-indigo-900/40 transition flex items-center justify-between text-left group"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-lg bg-indigo-900/60 text-indigo-400 flex items-center justify-center shrink-0">
                        <BookOpen className="w-4 h-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-white group-hover:text-indigo-400 transition truncate">
                          {item.title}
                        </p>
                        <p className="text-xs text-slate-400 truncate">{item.match}</p>
                      </div>
                    </div>
                    <ArrowRight className="w-4 h-4 text-slate-300 group-hover:text-indigo-400 group-hover:translate-x-1 transition shrink-0 ml-2" />
                  </button>
                ))
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
