// Mirrors Postgres schema: CHECK (state IN ('not_started', 'in_progress', 'weak_evidence', 'competent'))
export const NODE_STATES = {
  NOT_STARTED: 'not_started',
  IN_PROGRESS: 'in_progress',
  WEAK_EVIDENCE: 'weak_evidence',
  COMPETENT: 'competent',
};

// Mirrors Postgres schema: CHECK (role IN ('learner', 'curator', 'admin'))
export const USER_ROLES = {
  LEARNER: 'learner',
  CURATOR: 'curator',
  ADMIN: 'admin',
};

export const NODE_CONFIG = {
  [NODE_STATES.COMPETENT]: {
    label: 'Competent',
    badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200 ring-1 ring-emerald-500/20',
    dotClass: 'bg-emerald-500',
    colorHex: '#10B981',
  },
  [NODE_STATES.IN_PROGRESS]: {
    label: 'In Progress',
    badgeClass: 'bg-indigo-900/40 backdrop-blur-sm text-indigo-400 border-blue-200 ring-1 ring-blue-500/20',
    dotClass: 'bg-indigo-600',
    colorHex: '#2563EB',
  },
  [NODE_STATES.WEAK_EVIDENCE]: {
    label: 'Weak Evidence',
    badgeClass: 'bg-amber-50 text-amber-700 border-amber-200 ring-1 ring-amber-500/20',
    dotClass: 'bg-amber-500',
    colorHex: '#F59E0B',
  },
  [NODE_STATES.NOT_STARTED]: {
    label: 'Locked',
    badgeClass: 'bg-black/30 backdrop-blur-md text-slate-400 border-white/10',
    dotClass: 'bg-slate-400',
    colorHex: '#94A3B8',
  },
};
