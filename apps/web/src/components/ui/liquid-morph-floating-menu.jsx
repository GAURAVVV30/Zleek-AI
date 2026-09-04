import React, { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Menu, X } from "lucide-react";
import { NavLink } from "react-router-dom";

export default function LiquidMorphFloatingMenu({ navItems }) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef(null);

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Escape to close
  useEffect(() => {
    const handleEsc = (e) => {
      if (e.key === "Escape") setIsOpen(false);
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, []);

  return (
    <div 
      className="fixed z-50 left-4 md:left-6 top-20 md:top-24"
      ref={menuRef}
    >
      {!isOpen ? (
        <button
          onClick={() => setIsOpen(true)}
          className="w-12 h-12 md:w-14 md:h-14 rounded-full bg-black/60 backdrop-blur-xl border border-indigo-500/30 text-indigo-100 hover:text-white hover:border-indigo-400/60 hover:scale-105 active:scale-95 shadow-[0_0_25px_rgba(79,70,229,0.3)] transition-all duration-150 flex items-center justify-center cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-indigo-400"
          aria-label="Open navigation"
        >
          <Menu className="w-5 h-5 md:w-6 md:h-6" />
        </button>
      ) : (
        <AnimatePresence>
          <motion.div
            key="menu-drawer"
            initial={{ opacity: 0, scale: 0.92, y: -6 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.92, y: -6 }}
            transition={{ duration: 0.15, ease: [0.16, 1, 0.3, 1] }}
            className="bg-black/85 backdrop-blur-xl border border-indigo-500/30 rounded-3xl w-[260px] md:w-[280px] shadow-[0_0_35px_rgba(79,70,229,0.35)] overflow-hidden p-4 flex flex-col gap-2"
          >
            <div className="flex justify-between items-center mb-1 px-2">
              <span className="text-[10px] font-bold text-indigo-300 uppercase tracking-widest flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-pulse"></span>
                Navigation
              </span>
              <button 
                onClick={(e) => { e.stopPropagation(); setIsOpen(false); }}
                className="text-slate-400 hover:text-white p-1.5 rounded-full hover:bg-white/10 transition outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 cursor-pointer"
                aria-label="Close navigation"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="flex flex-col gap-1.5">
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    onClick={() => setIsOpen(false)}
                    className={({ isActive }) =>
                      `flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-semibold transition-all duration-150 group outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 ${
                        isActive
                          ? "bg-indigo-900/60 text-indigo-200 shadow-[0_0_15px_rgba(79,70,229,0.3)] border border-indigo-500/40"
                          : "text-slate-300 hover:text-white hover:bg-white/10 border border-transparent hover:border-white/10"
                      }`
                    }
                  >
                    {({ isActive }) => (
                      <>
                        <div className={`p-1.5 rounded-lg transition-colors ${isActive ? 'bg-indigo-500/20 text-indigo-300' : 'bg-transparent group-hover:bg-white/10 text-slate-400 group-hover:text-white'}`}>
                          <Icon className="w-4 h-4" />
                        </div>
                        <span>{item.label}</span>
                      </>
                    )}
                  </NavLink>
                );
              })}
            </div>
          </motion.div>
        </AnimatePresence>
      )}
    </div>
  );
}
