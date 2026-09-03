import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import {
  Video,
  FileText,
  Wrench,
  CheckCircle,
  ArrowRight,
  ExternalLink,
  ArrowLeft,
  Sparkles,
  BookOpen,
  Lock,
  Layers,
} from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';
import ModuleGeminiChatbot from '../../components/learner/ModuleGeminiChatbot';

const DOMAIN_UUID_MAP = {
  'f9ec7df2-79d6-52b1-9786-2be23e1738ee': 'ai_data_scientist',
  'ffc78d4c-4ed7-5f25-a82c-52c38181bafd': 'ai_engineer',
  '3e5fea96-c7e9-5817-a396-daacbbafeb7b': 'backend_engineer',
  'c46c25c9-44c3-581e-b903-d217a9c8c03c': 'data_analyst',
  '4a9bcbcf-0459-531f-856d-b20c270c376b': 'full_stack',
  '66a1c5de-4d1a-5303-bce1-5439aad10da2': 'devops_sre',
  '9190cc21-9768-5b94-850e-3a28f66c4055': 'frontend_engineer',
  '5e728fa9-b31e-5170-9dce-dec79630dd21': 'full_stack',
  '3f02aad3-d57b-5129-9cf2-b914ed7e313e': 'machine_learning',
  '2e0c2a7f-e1e1-534c-8501-4074196f0915': 'mobile_engineer',
  '714b074c-43b9-5ecc-861d-f1f3bab7663f': 'product_manager',
  '1296678f-0912-5d9e-8d1a-5d331d2b1cd7': 'software_architect',
};

const normalizeRoleSlug = (role) => {
  if (!role) return '';
  const clean = role.trim().toLowerCase();
  if (DOMAIN_UUID_MAP[clean]) return DOMAIN_UUID_MAP[clean];
  if (clean === 'data_engineer') return 'full_stack';
  return clean;
};

export default function LearningWorkspacePage() {
  const { conceptId } = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { addToast } = useToast();

  // Active Role and Module resolution
  const [roleId, setRoleId] = useState('');
  const [moduleData, setModuleData] = useState(null);
  const [activeTab, setActiveTab] = useState('video'); // 'video' | 'documentation' | 'hands_on'
  const [isLoading, setIsLoading] = useState(true);
  const [completedModules, setCompletedModules] = useState([]);
  const [isCompleted, setIsCompleted] = useState(false);

  useEffect(() => {
    const fetchModuleContext = async () => {
      setIsLoading(true);
      let activeRole = normalizeRoleSlug(searchParams.get('role'));

      if (!activeRole) {
        try {
          const roadmapRes = await apiClient.get(ENDPOINTS.ROADMAP.BASE);
          const rData = roadmapRes?.data;
          if (rData?.domain || rData?.domain_id) {
            activeRole = normalizeRoleSlug(rData.domain || rData.domain_id);
          }
        } catch (e) {}
      }

      if (!activeRole) {
        activeRole = normalizeRoleSlug(localStorage.getItem('userActiveRole'));
      }

      if (!activeRole) {
        activeRole = 'full_stack';
      }

      setRoleId(activeRole);
      localStorage.setItem('userActiveRole', activeRole);

      // Module ID parameter
      const mQuery = conceptId || searchParams.get('module') || '1';

      try {
        const res = await apiClient.get(ENDPOINTS.CONCEPTS.GOLD(activeRole, mQuery));
        if (res?.data && res.data.status === 'success') {
          setModuleData(res.data);
        } else if (res?.status === 'success') {
          setModuleData(res);
        } else {
          setModuleData({
            status: 'unavailable',
            role_id: activeRole,
            module_id: mQuery,
            module_number: 1,
            module_name: 'Module Context',
            resources: { documentation: [], video: [], hands_on: [] },
          });
        }
      } catch (err) {
        setModuleData({
          status: 'unavailable',
          role_id: activeRole,
          module_id: mQuery,
          module_number: 1,
          module_name: 'Module Context',
          resources: { documentation: [], video: [], hands_on: [] },
        });
      } finally {
        setIsLoading(false);
      }

      // Load completed modules progress from localStorage
      const savedProgress = localStorage.getItem(`gold_completed_modules_${activeRole}`);
      if (savedProgress) {
        try {
          const parsed = JSON.parse(savedProgress);
          setCompletedModules(parsed);
          setIsCompleted(parsed.includes(mQuery) || parsed.includes(moduleData?.module_id));
        } catch (e) {}
      }
    };

    fetchModuleContext();
  }, [conceptId, searchParams]);

  useEffect(() => {
    if (moduleData?.module_id) {
      setIsCompleted(completedModules.includes(moduleData.module_id) || completedModules.includes(`${moduleData.module_number}`));
    }
  }, [moduleData, completedModules]);

  const handleCompleteModule = async () => {
    if (!moduleData) return;
    const currentId = moduleData.module_id || `${moduleData.module_number}`;
    const nextModules = Array.from(new Set([...completedModules, currentId, `${moduleData.module_number}`]));
    
    setCompletedModules(nextModules);
    setIsCompleted(true);
    localStorage.setItem(`gold_completed_modules_${roleId}`, JSON.stringify(nextModules));

    try {
      await apiClient.post(ENDPOINTS.CONCEPTS.ENGAGEMENT(currentId), { action: 'marked_reviewed' });
    } catch (e) {
      console.warn('Engagement sync error:', e);
    }

    addToast(`Module ${moduleData.module_number} marked as complete! Next module unlocked.`, 'success');
  };

  const handleNextModule = () => {
    if (!moduleData) return;
    const nextNum = (moduleData.module_number || 1) + 1;
    navigate(`/learn?role=${roleId}&module=${nextNum}`);
  };

  if (isLoading) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading Gold Tier learning module...</p>
      </div>
    );
  }

  // Active Category Resources
  const resourcesGroup = moduleData?.resources || { video: [], documentation: [], hands_on: [] };
  const currentCategoryResources =
    activeTab === 'video'
      ? resourcesGroup.video || []
      : activeTab === 'documentation'
      ? resourcesGroup.documentation || []
      : resourcesGroup.hands_on || [];

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Top Breadcrumb Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-400">
          <button onClick={() => navigate('/roadmap')} className="hover:text-indigo-400 flex items-center gap-1">
            <ArrowLeft className="w-3.5 h-3.5" /> Roadmap
          </button>
          <span>/</span>
          <span className="text-indigo-400 font-bold uppercase tracking-wider">{roleId.replace('_', ' ')}</span>
          <span>/</span>
          <span className="text-white">Module {moduleData?.module_number || 1}</span>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-3 py-1 bg-indigo-950/60 border border-indigo-500/30 text-indigo-300 font-mono text-[11px] rounded-full flex items-center gap-1.5">
            <Sparkles className="w-3.5 h-3.5 text-indigo-400" />
            Gold Tier Dataset
          </span>
        </div>
      </div>

      {/* Main Grid: Left = Module & Resources, Right = Gemini Chatbot */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column (8 cols) */}
        <div className="lg:col-span-7 xl:col-span-8 space-y-6">
          {/* Dedicated Module Header Card */}
          <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-[32px] p-6 sm:p-8 shadow-[0_0_30px_rgba(79,70,229,0.15)] space-y-4">
            <div className="flex items-center justify-between">
              <span className="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 font-mono text-xs font-bold rounded-lg">
                MODULE {moduleData?.module_number || 1}
              </span>
              {isCompleted ? (
                <span className="flex items-center gap-1.5 text-xs font-bold text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-3 py-1 rounded-full">
                  <CheckCircle className="w-4 h-4" /> Completed & Unlocked
                </span>
              ) : (
                <span className="flex items-center gap-1.5 text-xs font-semibold text-slate-400 bg-slate-900 border border-slate-700 px-3 py-1 rounded-full">
                  <Layers className="w-3.5 h-3.5 text-indigo-400" /> In Progress
                </span>
              )}
            </div>

            <h1 className="font-display text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              {moduleData?.module_name || 'Learning Module'}
            </h1>
            <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
              Authoritative learning path extracted from the Gold Tier resource dataset. Study the materials below and mark complete to unlock the next module.
            </p>
          </div>

          {/* Resource Category Tab Bar */}
          <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-2 flex items-center gap-2">
            <button
              onClick={() => setActiveTab('video')}
              className={`flex-1 py-3 px-4 rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 ${
                activeTab === 'video'
                  ? 'bg-indigo-600 text-white shadow-[0_0_20px_rgba(79,70,229,0.3)]'
                  : 'bg-transparent text-slate-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <Video className="w-4 h-4" />
              <span>VIDEO ({resourcesGroup.video?.length || 0})</span>
            </button>

            <button
              onClick={() => setActiveTab('documentation')}
              className={`flex-1 py-3 px-4 rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 ${
                activeTab === 'documentation'
                  ? 'bg-indigo-600 text-white shadow-[0_0_20px_rgba(79,70,229,0.3)]'
                  : 'bg-transparent text-slate-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <FileText className="w-4 h-4" />
              <span>DOCUMENTATION ({resourcesGroup.documentation?.length || 0})</span>
            </button>

            <button
              onClick={() => setActiveTab('hands_on')}
              className={`flex-1 py-3 px-4 rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 ${
                activeTab === 'hands_on'
                  ? 'bg-indigo-600 text-white shadow-[0_0_20px_rgba(79,70,229,0.3)]'
                  : 'bg-transparent text-slate-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <Wrench className="w-4 h-4" />
              <span>HANDS-ON ({resourcesGroup.hands_on?.length || 0})</span>
            </button>
          </div>

          {/* Single Active Resource Category Display Area */}
          <div className="space-y-4">
            {currentCategoryResources.length > 0 ? (
              currentCategoryResources.map((res, idx) => (
                <div
                  key={res.id || idx}
                  className="bg-black/30 backdrop-blur-md border border-white/10 hover:border-indigo-500/40 rounded-2xl p-5 transition-all duration-300 space-y-3 group"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-indigo-400 bg-indigo-950/60 px-2 py-0.5 rounded border border-indigo-500/20">
                          {activeTab.replace('_', ' ')}
                        </span>
                        {res.provider && (
                          <span className="text-[10px] text-slate-400 font-semibold">
                            Provider: {res.provider}
                          </span>
                        )}
                      </div>
                      <h3 className="text-base font-bold text-white group-hover:text-indigo-300 transition">
                        {res.title}
                      </h3>
                      {res.description && (
                        <p className="text-xs text-slate-300 leading-relaxed pt-1">
                          {res.description}
                        </p>
                      )}
                    </div>

                    {res.url ? (
                      <a
                        href={res.url}
                        target="_blank"
                        rel="noreferrer"
                        className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl text-xs flex items-center gap-1.5 shrink-0 shadow-md hover:shadow-indigo-500/20 transition"
                      >
                        <span>Open Resource</span>
                        <ExternalLink className="w-3.5 h-3.5" />
                      </a>
                    ) : (
                      <span className="px-3 py-1.5 bg-slate-900 border border-slate-700 text-slate-400 text-xs font-mono rounded-xl shrink-0">
                        Unavailable
                      </span>
                    )}
                  </div>
                </div>
              ))
            ) : (
              <div className="p-10 text-center text-slate-400 bg-black/20 backdrop-blur-md rounded-2xl border border-white/10 space-y-2">
                <p className="text-sm font-semibold">No {activeTab.replace('_', ' ')} resources mapped for this module.</p>
                <p className="text-xs text-slate-500">Switch tabs above to inspect other available resource categories.</p>
              </div>
            )}
          </div>

          {/* Module Completion Flow Actions */}
          <div className="pt-6 border-t border-white/10 flex flex-col sm:flex-row items-center justify-between gap-4">
            <button
              onClick={handleCompleteModule}
              className={`w-full sm:w-auto px-6 py-3.5 rounded-2xl font-bold text-xs transition flex items-center justify-center gap-2 ${
                isCompleted
                  ? 'bg-emerald-600/20 border border-emerald-500/40 text-emerald-300 cursor-default'
                  : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-[0_0_20px_rgba(79,70,229,0.3)]'
              }`}
            >
              <CheckCircle className="w-4 h-4" />
              <span>{isCompleted ? 'Module Completed' : 'Complete Module'}</span>
            </button>

            <button
              onClick={handleNextModule}
              disabled={!isCompleted}
              className={`w-full sm:w-auto px-6 py-3.5 rounded-2xl font-bold text-xs transition flex items-center justify-center gap-2 ${
                isCompleted
                  ? 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg cursor-pointer'
                  : 'bg-slate-900 border border-slate-800 text-slate-500 cursor-not-allowed'
              }`}
            >
              <span>Unlock & Proceed to Next Module</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Right Column (5 cols) — Scoped Gemini Chatbot */}
        <div className="lg:col-span-5 xl:col-span-4 sticky top-24">
          <ModuleGeminiChatbot
            roleId={roleId}
            roleName={roleId.replace('_', ' ')}
            moduleId={moduleData?.module_id || `${moduleData?.module_number}`}
            moduleName={moduleData?.module_name}
            resources={currentCategoryResources}
          />
        </div>
      </div>
    </div>
  );
}
