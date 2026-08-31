import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { UploadCloud, CheckCircle2, FileCode, Clock, ArrowRight, ArrowLeft } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { uploadProjectFile } from '../../services/storageService';
import { useToast } from '../../context/ToastContext';

export default function ProjectSubmissionPage() {
  const { conceptId } = useParams();
  const [project, setProject] = useState(null);
  const [selectedFile, setSelectedFile] = useState(null);
  const [notes, setNotes] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.CONCEPTS.PROJECT(conceptId || 'c_clustering'))
      .then((res) => {
        setProject(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, [conceptId]);

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleSubmitProject = async () => {
    if (!selectedFile) {
      addToast('Please attach your project file (.ipynb or .zip)', 'warning');
      return;
    }

    setIsUploading(true);
    try {
      await uploadProjectFile(selectedFile, conceptId || 'c_clustering', notes);
      setIsSubmitted(true);
      addToast('Project uploaded to S3 storage and queued for curator review!', 'success');
    } catch (err) {
      addToast('Project upload failed', 'error');
    } finally {
      setIsUploading(false);
    }
  };

  if (isLoading || !project) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Loading project requirements...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
        {/* Title & Description */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold text-indigo-400 uppercase tracking-wider">
              Applied Milestone
            </span>
          </div>
          <h1 className="font-display text-2xl font-extrabold text-white">{project.title}</h1>
          <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">{project.description}</p>
        </div>

        {/* Requirements Rubric */}
        <div className="bg-black/30 backdrop-blur-md p-5 rounded-2xl border border-white/10 space-y-3">
          <h3 className="text-xs font-bold text-white uppercase tracking-wider">Submission Requirements</h3>
          <ul className="space-y-2 text-xs text-slate-300">
            {project.requirements?.map((req, idx) => (
              <li key={idx} className="flex items-start gap-2">
                <CheckCircle2 className="w-4 h-4 text-indigo-400 shrink-0 mt-0.5" />
                <span>{req}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* Upload Dropzone */}
        <div className="space-y-3">
          <label className="block text-xs font-bold text-white uppercase tracking-wider">
            Your Submission Artifact (Jupyter Notebook / ZIP)
          </label>

          <div className="border-2 border-dashed border-slate-300 hover:border-blue-500 rounded-2xl p-8 text-center bg-slate-50/40 transition">
            <input
              type="file"
              id="file-upload"
              className="hidden"
              onChange={handleFileChange}
              accept=".ipynb,.zip,.pdf,.py"
            />
            <label htmlFor="file-upload" className="cursor-pointer flex flex-col items-center gap-2">
              <div className="w-12 h-12 rounded-2xl bg-indigo-900/40 backdrop-blur-sm text-indigo-400 flex items-center justify-center">
                <UploadCloud className="w-6 h-6" />
              </div>
              <p className="text-xs font-bold text-white">
                {selectedFile ? selectedFile.name : 'Click or drag files here to upload'}
              </p>
              <p className="text-[11px] text-slate-400">
                {selectedFile ? `${(selectedFile.size / 1024 / 1024).toFixed(2)} MB ready` : 'Direct upload to secure cloud storage (max 50MB)'}
              </p>
            </label>
          </div>
        </div>

        {/* Notes / Comments */}
        <div>
          <label className="block text-xs font-semibold text-white mb-1">
            Submission Notes & Execution Summary (Optional)
          </label>
          <textarea
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            className="w-full p-3 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600"
            placeholder="Mention any custom hyperparameters, dataset caveats, or key findings..."
          />
        </div>

        {/* Previous Attempts */}
        {project.previousAttempts?.length > 0 && (
          <div className="pt-2">
            <h4 className="text-xs font-bold text-white uppercase tracking-wider mb-2">Previous Attempts</h4>
            <div className="space-y-2">
              {project.previousAttempts.map((att, i) => (
                <div key={i} className="p-3 bg-black/30 backdrop-blur-md rounded-xl border border-white/10 flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2">
                    <FileCode className="w-4 h-4 text-slate-400" />
                    <span className="font-semibold text-white">{att.filename}</span>
                  </div>
                  <span className="px-2.5 py-0.5 bg-emerald-50 text-emerald-700 rounded-full font-bold text-[10px] uppercase">
                    {att.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Action Button */}
        <div className="pt-4 flex items-center justify-between border-t border-white/5">
          <button
            type="button"
            onClick={() => navigate('/roadmap')}
            className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white flex items-center gap-1"
          >
            <ArrowLeft className="w-4 h-4" /> Back to Roadmap
          </button>
          <button
            onClick={handleSubmitProject}
            disabled={isUploading}
            className="px-8 py-3.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-xs shadow-elevated transition flex items-center gap-2"
          >
            {isUploading ? 'Uploading to S3...' : 'Submit for Review'}
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
