import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Layers, CheckCircle2, PlayCircle, AlertTriangle, Lock } from 'lucide-react';
import { NODE_STATES, NODE_CONFIG } from '../../utils/constants';

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

export default function CompetencyRoadmap({ roadmap, selectedConcept, onSelectConcept }) {
  const navigate = useNavigate();
  if (!roadmap || !roadmap.nodes) return null;

  const handleModuleClick = (node, index) => {
    if (onSelectConcept) onSelectConcept(node);
    const rawRole = roadmap?.domain || roadmap?.domain_id || node?.domain || node?.domain_id || localStorage.getItem('userActiveRole');
    let activeRole = normalizeRoleSlug(rawRole);
    if (!activeRole) activeRole = 'full_stack';
    const moduleQuery = node?.id || `${index + 1}`;
    navigate(`/learn?role=${encodeURIComponent(activeRole)}&module=${encodeURIComponent(moduleQuery)}`);
  };

  return (
    <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-[20px] p-6 lg:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] w-full">
      <div className="flex items-center justify-between pb-4 border-b border-white/5 mb-6">
        <h2 className="text-sm font-bold text-white flex items-center gap-2">
          <Layers className="w-4 h-4 text-indigo-400" />
          Verified Competency Roadmap
        </h2>
        <span className="text-xs text-slate-400">{roadmap.nodes.length} Milestones</span>
      </div>

      <div className="relative grid grid-cols-5 gap-4 lg:gap-6 w-full items-stretch min-w-[600px] overflow-x-auto pb-4 [&::-webkit-scrollbar]:hidden">
        {roadmap.nodes.map((node, index) => {
          const isSelected = selectedConcept?.id === node.id;
          const config = NODE_CONFIG[node.state] || NODE_CONFIG[NODE_STATES.NOT_STARTED];

          return (
            <div key={node.id} className="relative w-full">
              {/* Connector Line to Next Node */}
              {index < roadmap.nodes.length - 1 && (
                <div className="absolute top-[36px] left-[36px] w-[calc(100%+16px)] lg:w-[calc(100%+24px)] h-[3px] bg-slate-600/50 z-0" />
              )}

              <div
                onClick={() => handleModuleClick(node, index)}
                className={`relative z-10 h-full flex flex-col justify-start p-5 gap-4 rounded-3xl border transition cursor-pointer group ${
                  isSelected
                    ? 'border-indigo-400 bg-indigo-900/40 backdrop-blur-sm/50 shadow-[0_0_20px_rgba(79,70,229,0.25)] ring-2 ring-indigo-500/30'
                    : 'border-white/10 bg-black/40 backdrop-blur-xl hover:border-slate-500 hover:bg-slate-800/60'
                }`}
              >
                <div className="flex flex-col gap-3 w-full">
                  {/* Status Indicator Icon & Badge */}
                  <div className="flex items-start justify-between w-full">
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
                        node.state === NODE_STATES.COMPETENT
                          ? 'bg-emerald-500 text-white'
                          : node.state === NODE_STATES.IN_PROGRESS
                          ? 'bg-indigo-600 text-white ring-4 ring-indigo-500/30'
                          : node.state === NODE_STATES.WEAK_EVIDENCE
                          ? 'bg-amber-500 text-white'
                          : 'bg-slate-800 text-slate-400 border border-slate-600'
                      }`}
                    >
                      {node.state === NODE_STATES.COMPETENT && <CheckCircle2 className="w-5 h-5" />}
                      {node.state === NODE_STATES.IN_PROGRESS && <PlayCircle className="w-5 h-5" />}
                      {node.state === NODE_STATES.WEAK_EVIDENCE && <AlertTriangle className="w-5 h-5" />}
                      {node.state === NODE_STATES.NOT_STARTED && <Lock className="w-4 h-4" />}
                    </div>
                    <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border shrink-0 ${config.badgeClass}`}>
                      {config.label}
                    </span>
                  </div>
                  
                  <h3 className="text-xs lg:text-sm font-bold text-white group-hover:text-indigo-400 transition line-clamp-3 leading-snug">
                    {node.title}
                  </h3>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
