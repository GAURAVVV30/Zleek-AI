import React, { useState, useRef, useEffect } from 'react';
import { Bot, Send, Sparkles, Shield, AlertTriangle } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function ModuleGeminiChatbot({ roleId, roleName, moduleId, moduleName, resources }) {
  const [messages, setMessages] = useState([
    {
      sender: 'assistant',
      text: `Hello! I am your AI Learning Assistant for ${moduleName || 'this module'}. Ask me any question related to this module's topics or resources.`,
      blocked: false,
    },
  ]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef(null);

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
          text: 'Sorry, I encountered an error communicating with the module chatbot.',
          blocked: false,
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-slate-950/80 backdrop-blur-xl border border-white/10 rounded-[32px] shadow-[0_0_40px_rgba(79,70,229,0.15)] flex flex-col h-[580px] overflow-hidden relative">
      {/* Header */}
      <div className="p-4 border-b border-white/10 bg-slate-900/60 backdrop-blur-md flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-indigo-600/30 border border-indigo-500/40 flex items-center justify-center">
            <Bot className="w-4 h-4 text-indigo-400" />
          </div>
          <div>
            <h3 className="text-xs font-bold text-white flex items-center gap-1.5">
              <span>Gemini Module Assistant</span>
              <Sparkles className="w-3 h-3 text-cyan-400" />
            </h3>
            <p className="text-[10px] text-slate-400 truncate max-w-[200px]">
              Scoped to {moduleName || 'Current Module'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-mono text-emerald-400">
          <Shield className="w-3 h-3" />
          <span>Scoped</span>
        </div>
      </div>

      {/* Scope Info Banner */}
      <div className="px-4 py-2 bg-indigo-950/40 border-b border-indigo-500/20 text-[11px] text-indigo-300 flex items-center gap-2">
        <Sparkles className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
        <span className="truncate">Context: {roleName || 'Active Role'} &bull; {moduleName}</span>
      </div>

      {/* Message List */}
      <div className="flex-1 p-4 overflow-y-auto space-y-3.5">
        {messages.map((msg, idx) => (
          <div
            key={idx}
            className={`flex flex-col ${msg.sender === 'user' ? 'items-end' : 'items-start'}`}
          >
            <div
              className={`max-w-[85%] rounded-2xl p-3 text-xs leading-relaxed ${
                msg.sender === 'user'
                  ? 'bg-indigo-600 text-white rounded-br-none shadow-[0_0_15px_rgba(79,70,229,0.3)]'
                  : msg.blocked
                  ? 'bg-red-950/50 border border-red-500/30 text-red-200 rounded-bl-none'
                  : 'bg-slate-900/80 border border-white/10 text-slate-200 rounded-bl-none shadow-sm'
              }`}
            >
              {msg.blocked && (
                <div className="flex items-center gap-1.5 text-red-400 font-bold mb-1 text-[11px]">
                  <AlertTriangle className="w-3.5 h-3.5" />
                  <span>Guardrails Warning</span>
                </div>
              )}
              <p className="whitespace-pre-wrap">{msg.text}</p>
            </div>
            <span className="text-[9px] text-slate-500 mt-1 px-1">
              {msg.sender === 'user' ? 'You' : 'Gemini'}
            </span>
          </div>
        ))}
        {isLoading && (
          <div className="flex items-center gap-2 text-xs text-slate-400 bg-slate-900/60 p-3 rounded-2xl border border-white/10 w-max animate-pulse">
            <Bot className="w-4 h-4 text-indigo-400 animate-spin" />
            <span>Consulting module knowledge...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input Box */}
      <form onSubmit={handleSend} className="p-3 border-t border-white/10 bg-slate-900/80 backdrop-blur-md flex items-center gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={`Ask about ${moduleName || 'this module'}...`}
          disabled={isLoading}
          className="flex-1 bg-black/40 border border-white/10 rounded-xl px-3.5 py-2 text-xs text-white placeholder:text-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition"
        />
        <button
          type="submit"
          disabled={!input.trim() || isLoading}
          className="p-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:hover:bg-indigo-600 text-white rounded-xl transition shrink-0"
        >
          <Send className="w-3.5 h-3.5" />
        </button>
      </form>
    </div>
  );
}
