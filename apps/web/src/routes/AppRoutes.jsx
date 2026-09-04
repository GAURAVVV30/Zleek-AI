import React from 'react';
import { Routes, Route } from 'react-router-dom';

// Layouts
import LearnerLayout from '../components/layouts/LearnerLayout';
import CuratorLayout from '../components/layouts/CuratorLayout';
import AdminLayout from '../components/layouts/AdminLayout';
import AuthLayout from '../components/layouts/AuthLayout';
import ProtectedRoute from '../components/layouts/ProtectedRoute';

// Pages
import LandingPage from '../pages/public/LandingPage';
import SignUpPage from '../pages/auth/SignUpPage';
import LoginPage from '../pages/auth/LoginPage';
import ForgotPasswordPage from '../pages/auth/ForgotPasswordPage';
import GoalDefinitionPage from '../pages/onboarding/GoalDefinitionPage';
import PreferencesPage from '../pages/onboarding/PreferencesPage';
import DiagnosticPage from '../pages/onboarding/DiagnosticPage';
import BaselineResultsPage from '../pages/onboarding/BaselineResultsPage';
import RoadmapPage from '../pages/learner/RoadmapPage';
import LearningWorkspacePage from '../pages/learner/LearningWorkspacePage';
import AssessmentPage from '../pages/learner/AssessmentPage';
import ProjectSubmissionPage from '../pages/learner/ProjectSubmissionPage';
import ProgressPage from '../pages/learner/ProgressPage';
import GoalAchievedPage from '../pages/learner/GoalAchievedPage';
import SettingsPage from '../pages/learner/SettingsPage';
import KnowledgeStructurePage from '../pages/curator/KnowledgeStructurePage';
import ResourceCurationPage from '../pages/curator/ResourceCurationPage';
import UserManagementPage from '../pages/admin/UserManagementPage';
import AuditLogPage from '../pages/admin/AuditLogPage';
import NotFoundPage from '../pages/NotFoundPage';

export default function AppRoutes() {
  return (
    <Routes>
      {/* 01: Landing Page */}
      <Route path="/" element={<LandingPage />} />

      {/* 02 & 03: Authentication */}
      <Route element={<AuthLayout />}>
        <Route path="/signup" element={<SignUpPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      </Route>

      {/* 04 - 07: Onboarding Flow */}
      <Route element={<ProtectedRoute allowedRoles={['learner', 'curator', 'admin']} />}>
        <Route path="/onboarding/goal" element={<GoalDefinitionPage />} />
        <Route path="/onboarding/preferences" element={<PreferencesPage />} />
        <Route path="/diagnostic" element={<DiagnosticPage />} />
        <Route path="/diagnostic/results" element={<BaselineResultsPage />} />
      </Route>

      {/* 08 - 14: Main Learner App */}
      <Route element={<ProtectedRoute allowedRoles={['learner', 'curator', 'admin']} />}>
        <Route element={<LearnerLayout />}>
          <Route path="/roadmap" element={<RoadmapPage />} />
          <Route path="/learn/:conceptId" element={<LearningWorkspacePage />} />
          <Route path="/learn" element={<LearningWorkspacePage />} />
          <Route path="/assessment/:conceptId" element={<AssessmentPage />} />
          <Route path="/project/:conceptId" element={<ProjectSubmissionPage />} />
          <Route path="/progress" element={<ProgressPage />} />
          <Route path="/goal-achieved" element={<GoalAchievedPage />} />
          <Route path="/completion-badge" element={<GoalAchievedPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Route>

      {/* 15 & 16: Curator Console */}
      <Route element={<ProtectedRoute allowedRoles={['curator', 'admin']} />}>
        <Route element={<CuratorLayout />}>
          <Route path="/curator/structures" element={<KnowledgeStructurePage />} />
          <Route path="/curator/resources" element={<ResourceCurationPage />} />
        </Route>
      </Route>

      {/* 17 & 18: Admin Console */}
      <Route element={<ProtectedRoute allowedRoles={['admin']} />}>
        <Route element={<AdminLayout />}>
          <Route path="/admin/users" element={<UserManagementPage />} />
          <Route path="/admin/audit" element={<AuditLogPage />} />
        </Route>
      </Route>

      {/* 404 Catch-all */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
