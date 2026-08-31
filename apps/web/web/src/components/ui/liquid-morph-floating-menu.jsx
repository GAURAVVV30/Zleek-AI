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

  // Spring physics for liquid-like morphing
  const springConfig = { type: "spring", stiffness: 350, damping: 25, mass: 0.8 };

  return (
    <div 
      className="fixed z-50 left-4 md:left-6 top-20 md:top-24"
      ref={menuRef}
    >
      <motion.div
        layout
        transition={springConfig}
        className={`bg-black/40 backdrop-blur-xl border border-indigo-500/20 shadow-[0_0_30px_rgba(79,70,229,0.25)] overflow-hidden flex flex-col justify-start relative
          ${isOpen ? "rounded-3xl w-[260px] md:w-[280px]" : "rounded-full w-12 h-12 md:w-14 md:h-14"}`
        }
        style={{ originX: 0, originY: 0 }}
      >
        <AnimatePresence mode="popLayout" initial={false}>
          {!isOpen ? (
            <motion.button
              key="trigger"
              layout
              initial={{ opacity: 0, scale: 0.5 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.5 }}
              transition={{ duration: 0.2 }}
              onClick={() => setIsOpen(true)}
              className="absolute inset-0 w-full h-full flex items-center justify-center text-indigo-100 hover:text-indigo-400 hover:bg-white/5 transition-colors cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 rounded-full"
              aria-label="Open navigation"
            >
              <Menu className="w-5 h-5 md:w-6 md:h-6" />
            </motion.button>
          ) : (
            <motion.div
              key="menu"
              layout
              initial={{ opacity: 0, filter: "blur(4px)" }}
              animate={{ opacity: 1, filter: "blur(0px)" }}
              exit={{ opacity: 0, filter: "blur(4px)" }}
              transition={{ duration: 0.3, delay: 0.05 }}
              className="p-4 flex flex-col gap-2 w-full h-full"
            >
              <div className="flex justify-between items-center mb-2 px-2">
                <span className="text-[10px] font-bold text-indigo-300 uppercase tracking-widest flex items-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-pulse"></span>
                  Navigation
                </span>
                <button 
                  onClick={(e) => { e.stopPropagation(); setIsOpen(false); }}
                  className="text-slate-400 hover:text-white p-1 rounded-full hover:bg-white/10 transition outline-none focus-visible:ring-2 focus-visible:ring-indigo-400"
                  aria-label="Close navigation"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              <div className="flex flex-col gap-1.5">
                {navItems.map((item, i) => {
                  const Icon = item.icon;
                  return (
                    <motion.div
                      key={item.path}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.1 + (i * 0.05), ...springConfig }}
                    >
                      <NavLink
                        to={item.path}
                        onClick={() => setIsOpen(false)}
                        className={({ isActive }) =>
                          `flex items-center gap-3 px-3.5 py-3 rounded-xl text-sm font-semibold transition-all duration-300 group outline-none focus-visible:ring-2 focus-visible:ring-indigo-400 ${
                            isActive
                              ? "bg-indigo-900/50 text-indigo-300 shadow-[0_0_15px_rgba(79,70,229,0.3)] border border-indigo-500/40 translate-x-1"
                              : "text-slate-300 hover:text-white hover:bg-white/5 border border-transparent hover:border-white/10 hover:translate-x-1"
                          }`
                        }
                      >
                        {({ isActive }) => (
                          <>
                            <div className={`p-1.5 rounded-lg transition-colors ${isActive ? 'bg-indigo-500/20' : 'bg-transparent group-hover:bg-white/10'}`}>
                              <Icon className="w-4 h-4 group-hover:scale-110 transition-transform" />
                            </div>
                            <span>{item.label}</span>
                          </>
                        )}
                      </NavLink>
                    </motion.div>
                  );
                })}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  );
}
