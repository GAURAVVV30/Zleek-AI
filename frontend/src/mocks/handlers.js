import { http, HttpResponse, delay } from 'msw';
import {
  mockDomains,
  mockCurrentUser,
  mockDiagnosticQuestions,
  mockBaselineResults,
  initialRoadmapNodes,
  mockConceptDetail,
  mockAssessmentQuiz,
  mockProject,
  mockNotifications,
} from './mockData';
import { NODE_STATES } from '../utils/constants';

let currentRoadmap = [...initialRoadmapNodes];
let currentUser = { ...mockCurrentUser };
let activeGoal = {
  id: 'goal_ds_01',
  title: 'Become a Data Scientist and build real-world ML projects',
  domainId: 'data-science',
  domainName: 'Data Science & Analytics',
  status: 'active',
};

export const handlers = [
  // 1. Auth Handlers
  http.post('/api/v1/auth/signup', async ({ request }) => {
    await delay(300);
    const body = await request.json();
    currentUser = {
      ...currentUser,
      email: body.email || 'learner@example.com',
      fullName: body.fullName || 'New Learner',
      role: 'learner',
    };
    return HttpResponse.json({
      success: true,
      data: {
        accessToken: 'mock_jwt_token_123',
        refreshToken: 'mock_refresh_token_123',
        user: currentUser,
      },
    });
  }),

  http.post('/api/v1/auth/login', async ({ request }) => {
    await delay(300);
    const body = await request.json();
    return HttpResponse.json({
      success: true,
      data: {
        accessToken: 'mock_jwt_token_123',
        refreshToken: 'mock_refresh_token_123',
        user: currentUser,
      },
    });
  }),

  http.get('/api/v1/auth/me', async () => {
    await delay(100);
    return HttpResponse.json({
      success: true,
      data: currentUser,
    });
  }),

  http.post('/api/v1/auth/logout', async () => {
    return HttpResponse.json({ success: true });
  }),

  // 2. Domains & Goals
  http.get('/api/v1/domains', async () => {
    await delay(150);
    return HttpResponse.json({
      success: true,
      data: mockDomains,
    });
  }),

  http.post('/api/v1/goals', async ({ request }) => {
    await delay(400);
    const body = await request.json();
    activeGoal = {
      id: 'goal_user_defined',
      title: body.goalText || 'Data Science Career Path',
      domainId: 'data-science',
      domainName: 'Data Science & Analytics',
      status: 'active',
    };
    return HttpResponse.json({
      success: true,
      data: {
        goalId: activeGoal.id,
        domainId: activeGoal.domainId,
        domainName: activeGoal.domainName,
        confidence: 0.94,
        isSupported: true,
      },
    });
  }),

  http.get('/api/v1/goals/current', async () => {
    return HttpResponse.json({
      success: true,
      data: activeGoal,
    });
  }),

  http.patch('/api/v1/profile/preferences', async () => {
    await delay(200);
    return HttpResponse.json({ success: true, message: 'Preferences saved.' });
  }),

  // 3. Diagnostic Assessment
  http.post('/api/v1/diagnostics/start', async () => {
    await delay(250);
    return HttpResponse.json({
      success: true,
      data: {
        sessionId: 'diag_sess_789',
        firstQuestion: mockDiagnosticQuestions[0],
        totalQuestions: mockDiagnosticQuestions.length,
      },
    });
  }),

  http.post('/api/v1/diagnostics/:sessionId/answer', async ({ params, request }) => {
    await delay(200);
    const body = await request.json();
    const currentIdx = mockDiagnosticQuestions.findIndex((q) => q.questionId === body.questionId);
    const nextQuestion = mockDiagnosticQuestions[currentIdx + 1] || null;

    return HttpResponse.json({
      success: true,
      data: {
        isComplete: !nextQuestion,
        nextQuestion,
      },
    });
  }),

  http.get('/api/v1/diagnostics/:sessionId/results', async () => {
    await delay(300);
    return HttpResponse.json({
      success: true,
      data: mockBaselineResults,
    });
  }),

  // 4. Roadmap
  http.get('/api/v1/roadmap', async () => {
    await delay(200);
    return HttpResponse.json({
      success: true,
      data: {
        goalId: activeGoal.id,
        goalTitle: activeGoal.title,
        progressPercentage: 64,
        currentNodeId: 'c_pandas',
        nodes: currentRoadmap,
      },
    });
  }),

  http.get('/api/v1/roadmap/concepts/:id/why', async ({ params }) => {
    await delay(150);
    return HttpResponse.json({
      success: true,
      data: {
        conceptId: params.id,
        conceptName: 'Data Analysis with Pandas',
        reason: 'Pandas is a foundational prerequisite for Statistics and Machine Learning. Your diagnostic showed a 35% gap in exploratory data transformations.',
        prerequisitesMet: ['Python Basics'],
        unlocksConcepts: ['Statistics & Probability', 'Machine Learning Models'],
      },
    });
  }),

  http.post('/api/v1/roadmap/regenerate', async () => {
    await delay(400);
    return HttpResponse.json({
      success: true,
      message: 'Roadmap generated successfully based on diagnostic baseline.',
      data: { nodes: currentRoadmap },
    });
  }),

  // 5. Learning Workspace
  http.get('/api/v1/concepts/:id', async ({ params }) => {
    await delay(200);
    return HttpResponse.json({
      success: true,
      data: mockConceptDetail,
    });
  }),

  http.post('/api/v1/concepts/:id/engagement', async () => {
    await delay(100);
    return HttpResponse.json({
      success: true,
      message: 'Engagement recorded. Assessment unlocked.',
    });
  }),

  http.get('/api/v1/concepts/:id/resources/:resId/why', async () => {
    return HttpResponse.json({
      success: true,
      data: {
        reason: 'Curated by PyData experts with 98% positive verification ratings across 1,200 learners.',
      },
    });
  }),

  http.get('/api/v1/concepts/:id/resources/alternate', async () => {
    return HttpResponse.json({
      success: true,
      data: mockConceptDetail.alternateResources,
    });
  }),

  http.post('/api/v1/resources/:id/feedback', async () => {
    return HttpResponse.json({ success: true, message: 'Thank you for your feedback!' });
  }),

  // 6. Assessment & Quiz
  http.get('/api/v1/concepts/:id/assessment', async () => {
    await delay(200);
    return HttpResponse.json({
      success: true,
      data: mockAssessmentQuiz,
    });
  }),

  http.post('/api/v1/concepts/:id/assessment/submit', async ({ params }) => {
    await delay(400);
    // Invalidate and advance roadmap node
    currentRoadmap = currentRoadmap.map((node) => {
      if (node.id === params.id) {
        return { ...node, state: NODE_STATES.COMPETENT };
      }
      if (node.id === 'c_stats') {
        return { ...node, state: NODE_STATES.IN_PROGRESS, unlockRequirement: undefined };
      }
      return node;
    });

    return HttpResponse.json({
      success: true,
      data: {
        passed: true,
        scorePercentage: 100,
        newCompetencyState: NODE_STATES.COMPETENT,
        feedback: 'Evidence verified! You demonstrated full understanding of Pandas aggregation, filtering, and missing data handling.',
        remediationTriggered: false,
      },
    });
  }),

  // 7. Project & Storage
  http.get('/api/v1/concepts/:id/project', async () => {
    return HttpResponse.json({
      success: true,
      data: mockProject,
    });
  }),

  http.post('/api/v1/storage/upload-url', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json({
      success: true,
      data: {
        uploadUrl: 'https://mock-s3-bucket.amazonaws.com/uploads/' + encodeURIComponent(body.filename || 'project.zip'),
        fileKey: `projects/${Date.now()}_${body.filename}`,
        expiresInSeconds: 3600,
      },
    });
  }),

  http.post('/api/v1/concepts/:id/project/submit', async ({ request }) => {
    await delay(400);
    return HttpResponse.json({
      success: true,
      data: {
        submissionId: 'sub_' + Date.now(),
        status: 'pending_review',
        message: 'Project uploaded successfully and queued for verification.',
      },
    });
  }),

  // 8. Progress & Summary
  http.get('/api/v1/progress/summary', async () => {
    await delay(200);
    return HttpResponse.json({
      success: true,
      data: {
        overallCompletionPercentage: 64,
        totalConcepts: 5,
        completedConcepts: 2,
        activeRemediations: 0,
        competencyBreakdown: [
          { domain: 'Python Basics', percentage: 90, status: 'Competent' },
          { domain: 'Data Analysis (Pandas)', percentage: 75, status: 'Competent' },
          { domain: 'Statistics & Probability', percentage: 40, status: 'Weak Evidence' },
          { domain: 'Machine Learning Models', percentage: 20, status: 'In Progress' },
          { domain: 'SQL & Data Extraction', percentage: 60, status: 'In Progress' },
        ],
      },
    });
  }),

  http.get('/api/v1/competency/detail', async () => {
    return HttpResponse.json({
      success: true,
      data: [
        { conceptId: 'c_python', conceptName: 'Python Basics', state: 'competent', lastEvidenceDate: '2026-08-24', evidenceType: 'quiz', score: 95 },
        { conceptId: 'c_pandas', conceptName: 'Data Analysis with Pandas', state: 'competent', lastEvidenceDate: '2026-08-26', evidenceType: 'quiz', score: 100 },
        { conceptId: 'c_stats', conceptName: 'Statistics & Hypothesis Testing', state: 'weak_evidence', lastEvidenceDate: '2026-08-23', evidenceType: 'diagnostic', score: 40 },
        { conceptId: 'c_ml', conceptName: 'Machine Learning Fundamentals', state: 'in_progress', lastEvidenceDate: '2026-08-20', evidenceType: 'diagnostic', score: 45 },
      ],
    });
  }),

  http.get('/api/v1/competency/:id/history', async ({ params }) => {
    return HttpResponse.json({
      success: true,
      data: [
        { attempt: 1, date: '2026-08-26 18:30', score: 100, result: 'Pass', details: 'Quiz passed with 3/3 questions correct.' }
      ],
    });
  }),

  http.get('/api/v1/goals/current/completion-summary', async () => {
    return HttpResponse.json({
      success: true,
      data: {
        goalTitle: activeGoal.title,
        completionDate: '2026-08-26',
        totalSkillsVerified: 5,
        masteryProofList: [
          'Python Advanced Idioms & Vectorization',
          'Data Analysis & Cleaning with Pandas',
          'Statistical Inference & Hypothesis Testing',
          'Supervised Regression & Classification',
          'Applied K-Means Customer Segmentation Capstone'
        ],
      },
    });
  }),

  // 9. Profile & Settings
  http.patch('/api/v1/profile/settings', async ({ request }) => {
    const body = await request.json();
    currentUser = { ...currentUser, ...body };
    return HttpResponse.json({ success: true, data: currentUser });
  }),

  http.post('/api/v1/auth/change-password', async () => {
    return HttpResponse.json({ success: true, message: 'Password updated.' });
  }),

  // 10. Curator & Admin Handlers
  http.get('/api/v1/curator/knowledge-structures', async () => {
    return HttpResponse.json({
      success: true,
      data: [
        {
          id: 'ks_ds',
          domain: 'Data Science',
          version: 2,
          status: 'published',
          concepts: [
            { id: '1', name: 'Python Basics', parent: null },
            { id: '2', name: 'Data Analysis', parent: '1', children: ['Exploratory Data Analysis', 'Data Cleaning'] },
            { id: '3', name: 'Statistics & Probability', parent: '2' },
            { id: '4', name: 'Machine Learning Models', parent: '3' },
          ],
        },
      ],
    });
  }),

  http.post('/api/v1/curator/knowledge-structures/validate', async () => {
    return HttpResponse.json({
      success: true,
      valid: true,
      message: 'Knowledge structure DAG validated with 0 circular dependencies.',
    });
  }),

  http.get('/api/v1/curator/resources', async () => {
    return HttpResponse.json({
      success: true,
      data: [
        { id: 'r1', title: 'Pandas Tutorial for Beginners', type: 'video', duration: '15 min', status: 'pending', curator: 'Gokulnaath' },
        { id: 'r2', title: 'Data Analysis Handbook', type: 'article', duration: '20 min', status: 'approved', curator: 'Darshan' },
        { id: 'r3', title: 'Advanced Pandas Tricks', type: 'video', duration: '25 min', status: 'pending', curator: 'Deepa' },
      ],
    });
  }),

  http.get('/api/v1/admin/users', async () => {
    return HttpResponse.json({
      success: true,
      data: [
        { id: 'u1', email: 'gokul@example.com', role: 'learner', status: 'active', createdAt: '2026-08-20' },
        { id: 'u2', email: 'curator1@example.com', role: 'curator', status: 'active', createdAt: '2026-08-15' },
        { id: 'u3', email: 'admin@example.com', role: 'admin', status: 'active', createdAt: '2026-08-01' },
      ],
    });
  }),

  http.get('/api/v1/admin/audit-log', async () => {
    return HttpResponse.json({
      success: true,
      data: [
        { id: 'a1', actor: 'admin@example.com', action: 'User role updated to Curator', target: 'curator1@example.com', timestamp: 'Aug 26, 10:30 AM' },
        { id: 'a2', actor: 'curator1@example.com', action: 'Resource approved: Pandas Handbook', target: 'res_pandas_101', timestamp: 'Aug 26, 09:15 AM' },
        { id: 'a3', actor: 'gokul@example.com', action: 'Goal defined: Data Science Career', target: 'goal_ds_01', timestamp: 'Aug 24, 08:45 AM' },
      ],
    });
  }),

  // 11. Cross-Cutting (Notifications, Search, Telemetry)
  http.get('/api/v1/notifications', async () => {
    return HttpResponse.json({ success: true, data: mockNotifications });
  }),

  http.patch('/api/v1/notifications/:id/read', async () => {
    return HttpResponse.json({ success: true });
  }),

  http.get('/api/v1/search', async ({ request }) => {
    const url = new URL(request.url);
    const q = (url.searchParams.get('q') || '').toLowerCase();
    const results = [
      { type: 'concept', title: 'Data Analysis with Pandas', link: '/learn/c_pandas', match: 'Core tabular manipulation library' },
      { type: 'concept', title: 'Machine Learning Models', link: '/learn/c_ml', match: 'Supervised classification and regression' },
      { type: 'resource', title: 'Pandas Cookbook Guide', link: '/learn/c_pandas', match: 'Syntax cheat sheet' },
    ].filter((item) => item.title.toLowerCase().includes(q) || item.match.toLowerCase().includes(q));

    return HttpResponse.json({ success: true, data: results });
  }),

  http.post('/api/v1/telemetry/events', async () => {
    return HttpResponse.json({ success: true, received: true });
  }),

  http.get('/api/v1/health', async () => {
    return HttpResponse.json({ status: 'healthy', timestamp: new Date().toISOString() });
  }),
];
