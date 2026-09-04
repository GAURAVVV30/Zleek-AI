import React, { useState, useRef, useEffect } from 'react';
import { Bot, Send, AlertTriangle } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function ModuleGeminiChatbot({ roleId, roleName, moduleId, moduleName, resources }) {
  const [messages, setMessages] = useState([
    {
      sender: 'assistant',
      text: 'Hello this is Tia- your personal learning ai assistant , how can i help you',
      blocked: false,
    },
  ]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef(null);
  const textareaRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  const handleSend = async (e) => {
    e?.preventDefault();
    const query = input.trim();
    if (!query || isLoading) return;

    const userMsg = { sender: 'user', text: query };
    setMessages((prev) => [...prev, userMsg]);
    setInput('');
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
    setIsLoading(true);

    try {
      const res = await apiClient.post(ENDPOINTS.CONCEPTS.MODULE_CHAT, {
        role_id: roleId,
        module_id: moduleId,
        user_message: query,
      });

      const replyText = res?.reply || res?.data?.reply || 'No response generated.';
      const isBlocked = res?.blocked || res?.data?.blocked || false;

      setMessages((prev) => [
        ...prev,
        {
          sender: 'assistant',
          text: replyText,
          blocked: isBlocked,
        },
      ]);
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          sender: 'assistant',
          text: 'Sorry, I encountered an error communicating with Tia-AI assistant.',
          blocked: false,
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleTextareaInput = (e) => {
    setInput(e.target.value);
    e.target.style.height = 'auto';
    e.target.style.height = `${Math.min(e.target.scrollHeight, 120)}px`;
  };

  return (
    <div className="bg-slate-950/80 backdrop-blur-xl border border-white/10 rounded-[32px] shadow-[0_0_40px_rgba(79,70,229,0.15)] flex flex-col h-full min-h-[600px] overflow-hidden relative">
      {/* Header — Tia-AI assistant strictly */}
      <div className="p-4 sm:p-5 border-b border-white/10 bg-slate-900/60 backdrop-blur-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-2xl bg-indigo-600/30 border border-indigo-500/40 flex items-center justify-center shadow-inner">
            <Bot className="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-white tracking-wide">
              Tia-AI assistant
            </h3>
          </div>
        </div>
      </div>

      {/* Scrollable Message Area */}
      <div className="flex-1 p-4 sm:p-6 overflow-y-auto space-y-4 min-h-0">
        {messages.map((msg, idx) => (
          <div
            key={idx}
            className={`flex flex-col ${msg.sender === 'user' ? 'items-end' : 'items-start'}`}
          >
            <div
              className={`max-w-[88%] rounded-2xl p-3.5 text-xs sm:text-sm leading-relaxed ${
                msg.sender === 'user'
                  ? 'bg-indigo-600 text-white rounded-br-none shadow-[0_0_20px_rgba(79,70,229,0.25)]'
                  : msg.blocked
                  ? 'bg-red-950/50 border border-red-500/30 text-red-200 rounded-bl-none'
                  : 'bg-slate-900/80 border border-white/10 text-slate-200 rounded-bl-none shadow-sm'
              }`}
            >
              {msg.blocked && (
                <div className="flex items-center gap-1.5 text-red-400 font-bold mb-1.5 text-xs">
                  <AlertTriangle className="w-4 h-4" />
                  <span>Guardrails Warning</span>
                </div>
              )}
              <p className="whitespace-pre-wrap">{msg.text}</p>
            </div>
            <span className="text-[10px] text-slate-500 mt-1 px-1 font-medium">
              {msg.sender === 'user' ? 'You' : 'Tia-AI'}
            </span>
          </div>
        ))}
        {isLoading && (
          <div className="flex items-center gap-2 text-xs text-slate-400 bg-slate-900/60 p-3 rounded-2xl border border-white/10 w-max animate-pulse">
            <Bot className="w-4 h-4 text-indigo-400 animate-spin" />
            <span>Tia is thinking...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 21st.dev AI Chat Input component pattern */}
      <div className="p-3 sm:p-4 border-t border-white/10 bg-slate-900/80 backdrop-blur-md">
        <form onSubmit={handleSend} className="relative bg-black/50 border border-white/10 rounded-2xl p-2 focus-within:border-indigo-500/50 focus-within:ring-1 focus-within:ring-indigo-500/30 transition-all flex flex-col gap-2">
          <textarea
            ref={textareaRef}
            rows={1}
            value={input}
            onChange={handleTextareaInput}
            onKeyDown={handleKeyDown}
            placeholder="Ask Tia anything..."
            disabled={isLoading}
            className="w-full bg-transparent text-xs sm:text-sm text-white placeholder:text-slate-500 focus:outline-none resize-none px-2 py-1 min-h-[38px] max-h-[120px]"
          />
          <div className="flex items-center justify-end px-2 pt-1 border-t border-white/5">
            <button
              type="submit"
              disabled={!input.trim() || isLoading}
              className="p-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-30 disabled:hover:bg-indigo-600 text-white rounded-xl transition-all shrink-0 flex items-center justify-center shadow-md shadow-indigo-600/30"
              title="Send message"
            >
              <Send className="w-3.5 h-3.5" />
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
