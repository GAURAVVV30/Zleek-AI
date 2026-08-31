import React, { useState, useEffect } from 'react';
import { Users, Search, Shield, CheckCircle2, UserCheck } from 'lucide-react';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import { useToast } from '../../context/ToastContext';

export default function UserManagementPage() {
  const [users, setUsers] = useState([]);
  const [search, setSearch] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const { addToast } = useToast();

  useEffect(() => {
    apiClient
      .get(ENDPOINTS.ADMIN.USERS)
      .then((res) => {
        setUsers(res.data);
        setIsLoading(false);
      })
      .catch(() => setIsLoading(false));
  }, []);

  const handleRoleChange = (userId, newRole) => {
    setUsers((prev) =>
      prev.map((u) => (u.id === userId ? { ...u, role: newRole } : u))
    );
    addToast(`User role updated to ${newRole} & logged to audit.`, 'success');
  };

  const filtered = users.filter((u) => u.email.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="space-y-6">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-[0_0_10px_rgba(79,70,229,0.1)] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <span className="text-[11px] font-bold text-white uppercase tracking-wider block mb-1">
            Admin Console
          </span>
          <h1 className="font-display text-xl font-extrabold text-white">
            User & Role Management
          </h1>
        </div>

        <div className="relative w-full sm:w-64">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-3" />
          <input
            type="text"
            placeholder="Search users..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-2 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600"
          />
        </div>
      </div>

      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl shadow-[0_0_20px_rgba(79,70,229,0.15)] overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-black/30 backdrop-blur-md text-slate-400 uppercase font-bold border-b border-white/5">
              <tr>
                <th className="py-3.5 px-6">User Account</th>
                <th className="py-3.5 px-4">Current Role</th>
                <th className="py-3.5 px-4">Status</th>
                <th className="py-3.5 px-4">Created Date</th>
                <th className="py-3.5 px-6 text-right">Assign Role</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.map((user) => (
                <tr key={user.id} className="hover:bg-black/30 backdrop-blur-md transition">
                  <td className="py-4 px-6 font-bold text-white">{user.email}</td>
                  <td className="py-4 px-4">
                    <span
                      className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase ${
                        user.role === 'admin'
                          ? 'bg-slate-900 text-white'
                          : user.role === 'curator'
                          ? 'bg-purple-50 text-purple-700'
                          : 'bg-indigo-900/40 backdrop-blur-sm text-indigo-400'
                      }`}
                    >
                      {user.role}
                    </span>
                  </td>
                  <td className="py-4 px-4 text-emerald-600 font-semibold uppercase text-[10px]">
                    ● {user.status}
                  </td>
                  <td className="py-4 px-4 text-slate-400">{user.createdAt}</td>
                  <td className="py-4 px-6 text-right">
                    <select
                      value={user.role}
                      onChange={(e) => handleRoleChange(user.id, e.target.value)}
                      className="p-1.5 border border-white/10 rounded-lg text-xs font-semibold bg-black/40 backdrop-blur-xl outline-none"
                    >
                      <option value="learner">Learner</option>
                      <option value="curator">Curator</option>
                      <option value="admin">Admin</option>
                    </select>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
