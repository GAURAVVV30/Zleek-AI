import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Award, ArrowRight, ArrowLeft, CheckCircle2, AlertTriangle, Sparkles } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function AssessmentPage() {
  const { conceptId } = useParams();
  const [assessment, setAssessment] = useState(null);
  const [answers, setAnswers] = useState({});
  const [result, setResult] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.CONCEPTS.ASSESSMENT(conceptId || 'c_pandas'))
      .then((res) => {
        setAssessment(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, [conceptId]);

  const handleSelectOption = (questionId, optionId) => {
    setAnswers((prev) => ({ ...prev, [questionId]: optionId }));
  };

  const handleSubmitAssessment = async () => {
    if (Object.keys(answers).length < (assessment?.questions?.length || 1)) {
      addToast('Please answer all assessment items before submitting.', 'warning');
      return;
    }

    setIsSubmitting(true);
    try {
      const payload = {
        answers: Object.entries(answers).map(([qId, optId]) => ({
          questionId: qId,
          selectedOptionId: optId,
        })),
      };

      const res = await apiClient.post(
        ENDPOINTS.CONCEPTS.SUBMIT_ASSESSMENT(conceptId || 'c_pandas'),
        payload
      );
      setResult(res.data);
      addToast('Assessment submitted! Competency record updated.', 'success');
    } catch (err) {
      addToast('Failed to submit assessment', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading || !assessment) {
    return (
      <div className="py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Preparing concept assessment...</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Result View if Already Submitted */}
      {result ? (
        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] text-center space-y-6 animate-in zoom-in-95 duration-200">
          <div className="w-16 h-16 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center mx-auto">
            <Award className="w-8 h-8" />
          </div>

          <div>
            <span className="px-3 py-1 bg-emerald-50 text-emerald-700 font-bold text-xs rounded-full uppercase tracking-wider">
              {result.newCompetencyState}
            </span>
            <h2 className="font-display text-2xl font-bold text-white mt-2">
              Demonstrated Competence!
            </h2>
            <p className="text-xs text-slate-300 max-w-md mx-auto mt-2 leading-relaxed">
              {result.feedback}
            </p>
          </div>

          <div className="p-4 bg-black/30 backdrop-blur-md rounded-2xl border border-white/10 text-xs text-white">
            Score: <strong className="text-white font-bold">{result.scorePercentage}%</strong> · Evidence verified & logged to audit record.
          </div>

          <button
            onClick={() => navigate('/roadmap')}
            className="px-8 py-3.5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl text-xs shadow-elevated transition inline-flex items-center gap-2"
          >
            <span>Return to Roadmap</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      ) : (
        /* Assessment Question Cards */
        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
          <div className="flex items-center justify-between pb-4 border-b border-white/5">
            <div>
              <span className="text-[11px] font-bold text-indigo-400 uppercase tracking-wider block">
                Evidence Collection
              </span>
              <h1 className="text-lg font-bold text-white">{assessment.conceptTitle} — Assessment</h1>
            </div>
            <span className="text-xs text-slate-400 font-medium">{assessment.questions?.length} Questions</span>
          </div>

          <div className="space-y-6">
            {assessment.questions?.map((q) => (
              <div key={q.id} className="p-5 bg-slate-50/60 rounded-2xl border border-white/10 space-y-3">
                <p className="text-xs sm:text-sm font-bold text-white">
                  {q.number}. {q.prompt}
                </p>

                <div className="space-y-2">
                  {q.options?.map((opt) => {
                    const isSelected = answers[q.id] === opt.id;
                    return (
                      <button
                        key={opt.id}
                        type="button"
                        onClick={() => handleSelectOption(q.id, opt.id)}
                        className={`w-full p-3.5 rounded-xl border text-left text-xs font-medium transition flex items-center gap-3 ${
                          isSelected
                            ? 'border-blue-600 bg-indigo-900/40 backdrop-blur-sm text-blue-900 font-semibold'
                            : 'border-white/10 bg-black/40 backdrop-blur-xl text-white hover:bg-black/30 backdrop-blur-md'
                        }`}
                      >
                        <div
                          className={`w-3.5 h-3.5 rounded-full border flex items-center justify-center shrink-0 ${
                            isSelected ? 'border-blue-600 bg-indigo-600' : 'border-slate-300'
                          }`}
                        >
                          {isSelected && <div className="w-1.5 h-1.5 rounded-full bg-black/40 backdrop-blur-xl"></div>}
                        </div>
                        <span>{opt.text}</span>
                      </button>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>

          <div className="pt-4 flex items-center justify-between border-t border-white/5">
            <button
              type="button"
              onClick={() => navigate(`/learn/${conceptId || 'c_pandas'}`)}
              className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white flex items-center gap-1"
            >
              <ArrowLeft className="w-4 h-4" /> Back to Learning
            </button>
            <button
              onClick={handleSubmitAssessment}
              disabled={isSubmitting}
              className="px-8 py-3.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-xs shadow-elevated transition flex items-center gap-2"
            >
              {isSubmitting ? 'Evaluating evidence...' : 'Submit Answers for Verification'}
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
