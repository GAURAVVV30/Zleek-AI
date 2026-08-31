export const ENDPOINTS = {
  AUTH: {
    SIGNUP: '/auth/signup',
    LOGIN: '/auth/login',
    REFRESH: '/auth/refresh',
    LOGOUT: '/auth/logout',
    ME: '/auth/me',
    CHANGE_PASSWORD: '/auth/change-password',
  },
  DOMAINS: '/domains',
  GOALS: {
    BASE: '/goals',
    CURRENT: '/goals/current',
    COMPLETION_SUMMARY: '/goals/current/completion-summary',
  },
  PROFILE: {
    SETTINGS: '/profile/settings',
    PREFERENCES: '/profile/preferences',
    ACCOUNT: '/profile/account',
  },
  DIAGNOSTIC: {
    START: '/diagnostics/start',
    ANSWER: (sessionId) => `/diagnostics/${sessionId}/answer`,
    RESULTS: (sessionId) => `/diagnostics/${sessionId}/results`,
  },
  ROADMAP: {
    BASE: '/roadmap',
    WHY_CONCEPT: (id) => `/roadmap/concepts/${id}/why`,
    REGENERATE: '/roadmap/regenerate',
    TASKS: '/roadmap/tasks',
    TOGGLE_TASK: (id) => `/roadmap/tasks/${id}/toggle`,
  },
  CONCEPTS: {
    DETAIL: (id) => `/concepts/${id}`,
    ALTERNATE: (id) => `/concepts/${id}/resources/alternate`,
    WHY_RESOURCE: (id, resId) => `/concepts/${id}/resources/${resId}/why`,
    ENGAGEMENT: (id) => `/concepts/${id}/engagement`,
    ASSESSMENT: (id) => `/concepts/${id}/assessment`,
    SUBMIT_ASSESSMENT: (id) => `/concepts/${id}/assessment/submit`,
    PROJECT: (id) => `/concepts/${id}/project`,
    SUBMIT_PROJECT: (id) => `/concepts/${id}/project/submit`,
    PROJECT_STATUS: (id) => `/concepts/${id}/project/status`,
  },
  RESOURCES: {
    FEEDBACK: (id) => `/resources/${id}/feedback`,
  },
  PROGRESS: {
    SUMMARY: '/progress/summary',
  },
  COMPETENCY: {
    DETAIL: '/competency/detail',
    HISTORY: (conceptId) => `/competency/${conceptId}/history`,
  },
  CURATOR: {
    STRUCTURES: '/curator/knowledge-structures',
    VALIDATE_STRUCTURE: '/curator/knowledge-structures/validate',
    RESOURCES: '/curator/resources',
    RESOURCE_DETAIL: (id) => `/curator/resources/${id}`,
    RESOURCE_FEEDBACK_SIGNALS: (id) => `/curator/resources/${id}/feedback-signals`,
  },
  ADMIN: {
    USERS: '/admin/users',
    USER_DETAIL: (id) => `/admin/users/${id}`,
    AUDIT_LOG: '/admin/audit-log',
  },
  STORAGE: {
    UPLOAD_URL: '/storage/upload-url',
  },
  SEARCH: '/search',
  TELEMETRY: {
    EVENTS: '/telemetry/events',
  },
  HEALTH: '/health',
};
