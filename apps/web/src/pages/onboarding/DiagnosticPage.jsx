import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowRight, ArrowLeft, CheckCircle, HelpCircle } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function DiagnosticPage() {
  const [currentQuestion, setCurrentQuestion] = useState(null);
  const [sessionId, setSessionId] = useState('');
  const [selectedOption, setSelectedOption] = useState('');
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

  const handleNext = async () => {
    if (!selectedOption) {
      addToast('Please select an answer to proceed.', 'warning');
      return;
    }

    try {
      const res = await apiClient.post(ENDPOINTS.DIAGNOSTIC.ANSWER(sessionId), {
        questionId: currentQuestion.questionId,
        selectedOptionId: selectedOption,
      });

      if (res.data.isComplete || !res.data.nextQuestion) {
        localStorage.setItem('diagnosticSessionId', sessionId);
        addToast('Diagnostic complete! Computing skill baseline...', 'success');
        navigate('/diagnostic/results');
      } else {
        setCurrentQuestion(res.data.nextQuestion);
        setSelectedOption('');
        setQuestionIndex((prev) => prev + 1);
      }
    } catch (err) {
      addToast('Failed to submit response', 'error');
    }
  };

  if (isLoading || !currentQuestion) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-20 text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-xs text-slate-400">Preparing personalized diagnostic assessment...</p>
      </div>
    );
  }

  const progressPercent = Math.round((questionIndex / totalQuestions) * 100);

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
          <div className="w-full h-2 bg-slate-100 rounded-full overflow-hidden">
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
          {currentQuestion.options.map((option) => {
            const isSelected = selectedOption === option.id;
            return (
              <button
                key={option.id}
                type="button"
                onClick={() => setSelectedOption(option.id)}
                className={`w-full p-4 rounded-2xl border text-left text-xs sm:text-sm font-medium transition flex items-start gap-3 ${
                  isSelected
                    ? 'border-blue-600 bg-indigo-900/40 backdrop-blur-sm/70 text-blue-900 font-semibold shadow-[0_0_10px_rgba(79,70,229,0.1)]'
                    : 'border-white/10 text-white hover:bg-black/30 backdrop-blur-md hover:border-slate-300'
                }`}
              >
                <div
                  className={`w-4 h-4 rounded-full border mt-0.5 flex items-center justify-center shrink-0 ${
                    isSelected ? 'border-blue-600 bg-indigo-600' : 'border-slate-300'
                  }`}
                >
                  {isSelected && <div className="w-1.5 h-1.5 rounded-full bg-black/40 backdrop-blur-xl"></div>}
                </div>
                <span className="leading-relaxed">{option.text}</span>
              </button>
            );
          })}
        </div>

        {/* CTAs */}
        <div className="pt-4 flex items-center justify-between border-t border-white/5">
          <button
            type="button"
            onClick={() => navigate('/onboarding/preferences')}
            className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white flex items-center gap-1"
          >
            <ArrowLeft className="w-4 h-4" /> Save & Exit
          </button>
          <button
            onClick={handleNext}
            className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-semibold text-xs shadow-[0_0_15px_rgba(79,70,229,0.2)] shadow-blue-500/20 transition flex items-center gap-1.5"
          >
            {questionIndex === totalQuestions ? 'Complete Diagnostic' : 'Next Question'}
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
