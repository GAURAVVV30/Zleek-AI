import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowRight, ArrowLeft, CheckCircle, XCircle, HelpCircle, Check } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function DiagnosticPage() {
  const [currentQuestion, setCurrentQuestion] = useState(null);
  const [sessionId, setSessionId] = useState('');
  const [selectedOption, setSelectedOption] = useState('');
  const [checkedResult, setCheckedResult] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [questionIndex, setQuestionIndex] = useState(1);
  const [totalQuestions, setTotalQuestions] = useState(5);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const { addToast } = useToast();

  useEffect(() => {
    // Start diagnostic session
    apiClient
      .post(ENDPOINTS.DIAGNOSTIC.START)
      .then((res) => {
        setSessionId(res.data.sessionId);
        setCurrentQuestion(res.data.firstQuestion);
        setTotalQuestions(res.data.totalQuestions || 5);
        setIsLoading(false);
      })
      .catch(() => {
        setIsLoading(false);
      });
  }, []);

  const handleCheckAnswer = async () => {
    if (!selectedOption) {
      addToast('Please select an answer to check.', 'warning');
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await apiClient.post(ENDPOINTS.DIAGNOSTIC.ANSWER(sessionId), {
        questionId: currentQuestion.questionId,
        selectedOptionId: selectedOption,
      });

      setCheckedResult(res.data);
    } catch (err) {
      addToast('Failed to validate response', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleNext = () => {
    if (!checkedResult) return;

    if (checkedResult.isComplete || !checkedResult.nextQuestion) {
      localStorage.setItem('diagnosticSessionId', sessionId);
      addToast('Diagnostic complete! Computing skill baseline...', 'success');
      navigate('/diagnostic/results');
    } else {
      setCurrentQuestion(checkedResult.nextQuestion);
      setSelectedOption('');
      setCheckedResult(null);
      setQuestionIndex((prev) => prev + 1);
    }
  };

  if (isLoading || !currentQuestion) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-20 text-center">
        <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Preparing personalized diagnostic assessment...</p>
      </div>
    );
  }

  const progressPercent = Math.round((questionIndex / totalQuestions) * 100);

  // Derive correct option object and human-readable text robustly across shape schemas
  const correctOptObj = checkedResult && currentQuestion && currentQuestion.options
    ? currentQuestion.options.find((o, idx) => {
        const oid = typeof o === 'object' && o && o.id ? o.id : `opt_${idx + 1}`;
        return oid === checkedResult.correctOptionId;
      })
    : null;

  const correctOptText = correctOptObj
    ? typeof correctOptObj === 'string'
      ? correctOptObj
      : correctOptObj.text || correctOptObj.label || correctOptObj.optionText || correctOptObj.content || correctOptObj.value || ''
    : '';

  return (
    <div className="max-w-2xl mx-auto px-4 py-12">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
        {/* Progress Bar & Header */}
        <div>
          <div className="flex items-center justify-between text-xs font-semibold text-slate-400 mb-2">
            <span>
              Question <strong className="text-indigo-400 font-bold">{questionIndex}</strong> of {totalQuestions}
            </span>
            <span className="text-indigo-400">{progressPercent}% complete</span>
          </div>
          <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-indigo-600 transition-all duration-300 rounded-full"
              style={{ width: `${progressPercent}%` }}
            ></div>
          </div>
        </div>

        {/* Concept Badge & Prompt */}
        <div className="space-y-3 pt-2">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-900/40 backdrop-blur-sm text-indigo-400 rounded-full text-xs font-semibold">
            <HelpCircle className="w-3.5 h-3.5" />
            {currentQuestion.conceptName}
          </div>
          <h2 className="text-base sm:text-lg font-bold text-white leading-snug">
            {currentQuestion.prompt}
          </h2>
        </div>

        {/* Options */}
        <div className="space-y-2.5 pt-2">
          {currentQuestion.options && currentQuestion.options.map((option, idx) => {
            const optionId = typeof option === 'object' && option && option.id ? option.id : `opt_${idx + 1}`;
            const optionText = typeof option === 'string'
              ? option
              : option.text || option.label || option.optionText || option.content || option.value || '';
            
            const isSelected = selectedOption === optionId;
            const letterLabel = String.fromCharCode(65 + idx); // 'A', 'B', 'C', 'D', 'E'

            let optionStyles = 'border-white/10 text-white hover:bg-black/30 backdrop-blur-md hover:border-slate-300';

            if (checkedResult) {
              if (optionId === checkedResult.correctOptionId) {
                optionStyles = 'border-emerald-500 bg-emerald-950/50 text-emerald-200 font-semibold shadow-[0_0_10px_rgba(16,185,129,0.2)]';
              } else if (optionId === checkedResult.selectedOptionId && !checkedResult.isCorrect) {
                optionStyles = 'border-rose-500 bg-rose-950/50 text-rose-200 font-semibold shadow-[0_0_10px_rgba(244,63,94,0.2)]';
              } else {
                optionStyles = 'border-white/5 text-slate-500 opacity-50 cursor-not-allowed';
              }
            } else if (isSelected) {
              optionStyles = 'border-indigo-500 bg-indigo-900/40 backdrop-blur-sm text-white font-semibold shadow-[0_0_10px_rgba(79,70,229,0.2)]';
            }

            return (
              <button
                key={optionId}
                type="button"
                disabled={!!checkedResult}
                onClick={() => setSelectedOption(optionId)}
                className={`w-full p-4 rounded-2xl border text-left text-xs sm:text-sm font-medium transition flex items-start gap-3.5 ${optionStyles}`}
              >
                {/* Option Letter Badge: A, B, C, D, E */}
                <div
                  className={`w-6 h-6 rounded-lg text-xs font-bold flex items-center justify-center shrink-0 border mt-0.5 ${
                    checkedResult && optionId === checkedResult.correctOptionId
                      ? 'border-emerald-400 bg-emerald-500 text-black'
                      : checkedResult && optionId === checkedResult.selectedOptionId && !checkedResult.isCorrect
                      ? 'border-rose-400 bg-rose-500 text-white'
                      : isSelected
                      ? 'border-indigo-400 bg-indigo-600 text-white'
                      : 'border-white/10 bg-white/5 text-slate-400'
                  }`}
                >
                  {letterLabel}
                </div>

                {/* Human Readable Option Text */}
                <span className="leading-relaxed pt-0.5 flex-1">{optionText}</span>

                {/* Status Indicator */}
                {checkedResult && optionId === checkedResult.correctOptionId && (
                  <Check className="w-4 h-4 text-emerald-400 shrink-0 mt-1 stroke-[3]" />
                )}
              </button>
            );
          })}
        </div>

        {/* Immediate Feedback Card after Checking */}
        {checkedResult && (
          <div className="pt-2 animate-fadeIn">
            {checkedResult.isCorrect ? (
              <div className="p-4 rounded-2xl bg-emerald-950/40 border border-emerald-500/40 text-emerald-300 flex items-center gap-3">
                <CheckCircle className="w-5 h-5 text-emerald-400 shrink-0" />
                <div>
                  <p className="font-bold text-xs sm:text-sm">Correct!</p>
                  <p className="text-[11px] sm:text-xs text-emerald-400/80 mt-0.5">
                    Excellent work. You have demonstrated mastery of {currentQuestion.conceptName}.
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-4 rounded-2xl bg-rose-950/40 border border-rose-500/40 text-rose-300 flex items-center gap-3">
                <XCircle className="w-5 h-5 text-rose-400 shrink-0" />
                <div>
                  <p className="font-bold text-xs sm:text-sm">Incorrect</p>
                  <p className="text-[11px] sm:text-xs text-rose-200/90 mt-0.5">
                    Correct answer:{' '}
                    <strong className="text-white font-semibold">
                      {correctOptText ? correctOptText : 'Option ' + checkedResult.correctOptionId}
                    </strong>
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

        {/* CTAs: Check Answer vs Next Question */}
        <div className="pt-4 flex items-center justify-between border-t border-white/5">
          <button
            type="button"
            onClick={() => navigate('/onboarding/preferences')}
            className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white flex items-center gap-1"
          >
            <ArrowLeft className="w-4 h-4" /> Save & Exit
          </button>

          {!checkedResult ? (
            <button
              onClick={handleCheckAnswer}
              disabled={!selectedOption || isSubmitting}
              className={`px-6 py-2.5 rounded-xl font-semibold text-xs transition flex items-center gap-1.5 ${
                selectedOption && !isSubmitting
                  ? 'bg-indigo-600 hover:bg-indigo-700 text-white shadow-[0_0_15px_rgba(79,70,229,0.2)]'
                  : 'bg-slate-800 text-slate-500 cursor-not-allowed'
              }`}
            >
              {isSubmitting ? 'Checking...' : 'Check Answer'}
              <CheckCircle className="w-4 h-4" />
            </button>
          ) : (
            <button
              onClick={handleNext}
              className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-semibold text-xs shadow-[0_0_15px_rgba(79,70,229,0.2)] transition flex items-center gap-1.5"
            >
              {checkedResult.isComplete || questionIndex === totalQuestions ? 'Complete Diagnostic' : 'Next Question'}
              <ArrowRight className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
