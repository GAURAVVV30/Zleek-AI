import React, { useState, useEffect } from 'react';
import { FileText, Filter, ShieldCheck } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';

export default function AuditLogPage() {
  const [logs, setLogs] = useState([]);
  const [filterAction, setFilterAction] = useState('all');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.ADMIN.AUDIT_LOG)
      .then((res) => {
        setLogs(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-[0_0_10px_rgba(79,70,229,0.1)] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <span className="text-[11px] font-bold text-white uppercase tracking-wider block mb-1">
            Governance & Compliance
          </span>
          <h1 className="font-display text-xl font-extrabold text-white">
            Immutable Audit Log
          </h1>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
            className="px-3 py-2 border border-white/10 rounded-xl text-xs font-semibold bg-black/40 backdrop-blur-xl outline-none"
          >
            <option value="all">All System Events</option>
            <option value="curator">Curator Actions</option>
            <option value="admin">Admin Actions</option>
          </select>
        </div>
      </div>

      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl shadow-[0_0_20px_rgba(79,70,229,0.15)] overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-black/30 backdrop-blur-md text-slate-400 uppercase font-bold border-b border-white/5">
              <tr>
                <th className="py-3.5 px-6">Actor Email</th>
                <th className="py-3.5 px-4">Event Description</th>
                <th className="py-3.5 px-4">Target Entity</th>
                <th className="py-3.5 px-6 text-right">Timestamp</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {logs.map((log) => (
                <tr key={log.id} className="hover:bg-black/30 backdrop-blur-md transition">
                  <td className="py-4 px-6 font-bold text-white flex items-center gap-2">
                    <ShieldCheck className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                    <span>{log.actor}</span>
                  </td>
                  <td className="py-4 px-4 font-medium text-white">{log.action}</td>
                  <td className="py-4 px-4 font-mono text-[11px] text-slate-400">{log.target}</td>
                  <td className="py-4 px-6 text-right text-slate-400">{log.timestamp}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
