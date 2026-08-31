import { NODE_STATES } from '../utils/constants';

export const mockDomains = [
  {
    id: 'data-science',
    name: 'Data Science & Analytics',
    description: 'Master Python, Pandas, Statistical Analysis, and Machine Learning workflows.',
    popularGoals: [
      'Become a Data Scientist and build real-world ML projects',
      'Data Analytics with Python & SQL',
      'AI & Machine Learning Engineer'
    ],
  },
  {
    id: 'machine-learning',
    name: 'Machine Learning Engineering',
    description: 'Deep Learning, NLP, Computer Vision, and MLOps deployment.',
    popularGoals: ['ML Engineer', 'Computer Vision Specialist', 'LLM Application Developer'],
  },
  {
    id: 'backend-go',
    name: 'Backend Engineering with Go',
    description: 'High concurrency microservices, PostgreSQL, Docker, and Kubernetes.',
    popularGoals: ['Senior Go Backend Developer', 'Cloud Native Systems Architect'],
  }
];

export const mockCurrentUser = {
  id: 'usr_gokul_123',
  email: 'gokul@example.com',
  fullName: 'Gokulnaath N',
  role: 'learner',
  avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150&auto=format&fit=crop&q=80',
  timezone: 'Asia/Kolkata (GMT +5:30)',
  theme: 'system',
};

export const mockDiagnosticQuestions = [
  {
    questionId: 'diag_q1',
    questionNumber: 1,
    totalQuestions: 5,
    conceptId: 'c_python',
    conceptName: 'Python Basics',
    prompt: 'Which of the following is true about Python mutable vs immutable data types?',
    options: [
      { id: 'opt_1', text: 'Lists and dictionaries are mutable, while tuples and strings are immutable' },
      { id: 'opt_2', text: 'All sequence types in Python are mutable' },
      { id: 'opt_3', text: 'Strings can be modified in-place using item assignment' },
      { id: 'opt_4', text: 'Tuples can have new items appended after creation' }
    ]
  },
  {
    questionId: 'diag_q2',
    questionNumber: 2,
    totalQuestions: 5,
    conceptId: 'c_pandas',
    conceptName: 'Data Analysis with Pandas',
    prompt: 'Which method is used to handle missing (NaN) values in a Pandas DataFrame?',
    options: [
      { id: 'opt_1', text: 'dropna() or fillna()' },
      { id: 'opt_2', text: 'remove_null()' },
      { id: 'opt_3', text: 'clean_missing()' },
      { id: 'opt_4', text: 'filter_na()' }
    ]
  },
  {
    questionId: 'diag_q3',
    questionNumber: 3,
    totalQuestions: 5,
    conceptId: 'c_stats',
    conceptName: 'Hypothesis Testing',
    prompt: 'What does a p-value less than 0.05 typically indicate in statistical hypothesis testing?',
    options: [
      { id: 'opt_1', text: 'Reject the null hypothesis with statistical significance' },
      { id: 'opt_2', text: 'Accept the null hypothesis as true' },
      { id: 'opt_3', text: 'The sample size is too small' },
      { id: 'opt_4', text: 'The variance is equal to zero' }
    ]
  },
  {
    questionId: 'diag_q4',
    questionNumber: 4,
    totalQuestions: 5,
    conceptId: 'c_ml',
    conceptName: 'Bias-Variance Tradeoff',
    prompt: 'Which of the following is true about the bias-variance tradeoff in machine learning?',
    options: [
      { id: 'opt_1', text: 'High variance typically leads to overfitting on the training data' },
      { id: 'opt_2', text: 'High bias causes the model to memorize noise in the training set' },
      { id: 'opt_3', text: 'Both bias and variance always decrease as model complexity increases' },
      { id: 'opt_4', text: 'Bias and variance are completely independent of model capacity' }
    ]
  },
  {
    questionId: 'diag_q5',
    questionNumber: 5,
    totalQuestions: 5,
    conceptId: 'c_clustering',
    conceptName: 'Unsupervised Clustering',
    prompt: 'What is the primary metric minimized in K-Means clustering optimization?',
    options: [
      { id: 'opt_1', text: 'Within-Cluster Sum of Squares (Inertia)' },
      { id: 'opt_2', text: 'Cross-Entropy Loss' },
      { id: 'opt_3', text: 'Mean Absolute Error' },
      { id: 'opt_4', text: 'Gini Impurity' }
    ]
  }
];

export const mockBaselineResults = {
  assessedLevel: 'Intermediate',
  overallScorePercentage: 64,
  conceptCoverage: [
    { conceptId: 'c_python', conceptName: 'Python Basics', coveragePercentage: 90, status: 'competent' },
    { conceptId: 'c_pandas', conceptName: 'Data Analysis with Pandas', coveragePercentage: 65, status: 'in_progress' },
    { conceptId: 'c_stats', conceptName: 'Statistics & Hypothesis Testing', coveragePercentage: 40, status: 'gap' },
    { conceptId: 'c_ml', conceptName: 'Machine Learning Fundamentals', coveragePercentage: 45, status: 'gap' },
    { conceptId: 'c_clustering', conceptName: 'Capstone Project (Customer Segmentation)', coveragePercentage: 20, status: 'gap' },
  ],
  topGaps: ['Hypothesis Testing', 'Bias-Variance Tradeoff', 'K-Means Clustering'],
};

export const initialRoadmapNodes = [
  {
    id: 'c_python',
    title: 'Python Basics',
    description: 'Core syntax, data structures, functions, and list comprehensions.',
    domain: 'Data Science',
    state: NODE_STATES.COMPETENT,
    order: 1,
    estimatedMinutes: 45,
    isRemediation: false,
  },
  {
    id: 'c_pandas',
    title: 'Data Analysis with Pandas',
    description: 'Understand and apply data manipulation, cleaning, and exploration techniques with Pandas.',
    domain: 'Data Science',
    state: NODE_STATES.IN_PROGRESS,
    order: 2,
    estimatedMinutes: 60,
    isRemediation: false,
    nextSubConcept: 'Exploratory Data Analysis',
  },
  {
    id: 'c_stats',
    title: 'Statistics & Probability',
    description: 'Descriptive stats, distributions, hypothesis testing, and confidence intervals.',
    domain: 'Data Science',
    state: NODE_STATES.NOT_STARTED,
    order: 3,
    unlockRequirement: 'Unlocks after: Data Analysis with Pandas',
    estimatedMinutes: 90,
    isRemediation: false,
  },
  {
    id: 'c_ml',
    title: 'Machine Learning Models',
    description: 'Supervised regression, classification, model evaluation, and regularisation.',
    domain: 'Machine Learning',
    state: NODE_STATES.NOT_STARTED,
    order: 4,
    unlockRequirement: 'Unlocks after: Statistics & Probability',
    estimatedMinutes: 120,
    isRemediation: false,
  },
  {
    id: 'c_clustering',
    title: 'Capstone Project: Customer Segmentation',
    description: 'Build an end-to-end K-Means clustering notebook on real customer transaction data.',
    domain: 'Data Science',
    state: NODE_STATES.NOT_STARTED,
    order: 5,
    unlockRequirement: 'Unlocks after: Machine Learning Models',
    estimatedMinutes: 180,
    isRemediation: false,
  },
];

export const mockConceptDetail = {
  id: 'c_pandas',
  title: 'Data Analysis',
  breadcrumb: ['Data Science', 'Data Analysis', 'Exploratory Data Analysis'],
  whyItMatters: 'Pandas is the industry standard for tabular data processing. Mastering data frame filtering, grouping, and transformations is required before building ML pipelines.',
  primaryResource: {
    id: 'res_pandas_101',
    title: 'Data Analysis with Pandas (Complete Guide)',
    type: 'video',
    durationMinutes: 20,
    provider: 'YouTube / PyData Expert Series',
    sourceUrl: 'https://www.youtube-nocookie.com/embed/dcqPhpY7tWk',
    provenance: {
      author: 'Core Python & Data Science Guild',
      vettedBy: 'Curator Gokulnaath',
      vettedDate: '2026-08-20',
    },
    whyThisResource: 'This video is rated in the top 1% for pedagogical clarity and matches your preferred video format.'
  },
  alternateResources: [
    {
      id: 'res_pandas_alt_1',
      title: 'Pandas Cookbook & Interactive Cheat-Sheet',
      type: 'article',
      durationMinutes: 15,
      provider: 'Official Documentation'
    },
    {
      id: 'res_pandas_alt_2',
      title: 'Exploratory Data Analysis Lab Notebook',
      type: 'interactive',
      durationMinutes: 25,
      provider: 'Kaggle Learn'
    }
  ]
};

export const mockAssessmentQuiz = {
  conceptId: 'c_pandas',
  conceptTitle: 'Data Analysis with Pandas',
  questions: [
    {
      id: 'q_1',
      number: 1,
      prompt: 'Which method is used to handle missing values in a Pandas DataFrame?',
      options: [
        { id: 'opt_a', text: 'dropna()' },
        { id: 'opt_b', text: 'fillna()' },
        { id: 'opt_c', text: 'replace()' },
        { id: 'opt_d', text: 'All of the above' }
      ]
    },
    {
      id: 'q_2',
      number: 2,
      prompt: 'How do you filter rows in a DataFrame `df` where the column "age" is greater than 25?',
      options: [
        { id: 'opt_a', text: 'df[df["age"] > 25]' },
        { id: 'opt_b', text: 'df.filter(age > 25)' },
        { id: 'opt_c', text: 'df.where("age" > 25)' },
        { id: 'opt_d', text: 'df.select("age > 25")' }
      ]
    },
    {
      id: 'q_3',
      number: 3,
      prompt: 'Which Pandas function calculates aggregated statistics (mean, count, etc.) across groups?',
      options: [
        { id: 'opt_a', text: 'df.groupby("col").agg()' },
        { id: 'opt_b', text: 'df.split_apply()' },
        { id: 'opt_c', text: 'df.combine("col")' },
        { id: 'opt_d', text: 'df.pivot_simple()' }
      ]
    }
  ]
};

export const mockProject = {
  id: 'c_clustering',
  title: 'Capstone Project: Customer Segmentation',
  description: 'Build an end-to-end unsupervised clustering model on the provided e-commerce customer transaction dataset and present visual insights.',
  requirements: [
    'Perform Exploratory Data Analysis and data cleaning in a Jupyter Notebook (.ipynb)',
    'Use K-Means or Hierarchical Clustering with Elbow / Silhouette evaluation',
    'Provide at least 3 actionable business insights per segment',
    'Include visualizations for cluster distributions and PCA projections'
  ],
  rubric: [
    { criterion: 'Data Preprocessing & Scaling', weight: 25 },
    { criterion: 'Optimal Cluster Selection (Elbow/Silhouette)', weight: 30 },
    { criterion: 'Cluster Interpretation & Business Insights', weight: 45 }
  ],
  previousAttempts: [
    {
      attemptNumber: 1,
      submittedAt: '2026-08-24T14:30:00Z',
      filename: 'customer_segmentation_v1.ipynb',
      fileSize: '2.4 MB',
      status: 'approved',
      feedback: 'Excellent feature scaling and clean cluster visualizations.'
    }
  ]
};

export const mockNotifications = [
  {
    id: 'notif_1',
    title: 'Path Updated',
    message: 'Your path adapted: Data Analysis is ready for evaluation.',
    type: 'path_update',
    read: false,
    createdAt: '10 mins ago',
  },
  {
    id: 'notif_2',
    title: 'Milestone Achieved',
    message: 'Python Basics verified as Competent!',
    type: 'success',
    read: false,
    createdAt: '2 hours ago',
  },
  {
    id: 'notif_3',
    title: 'New Resource Added',
    message: 'Curator added an updated Pandas 2.0 reference guide.',
    type: 'info',
    read: true,
    createdAt: '1 day ago',
  },
];
