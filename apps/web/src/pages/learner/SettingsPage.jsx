import React, { useState } from 'react';
import { User, Lock, Bell, Trash2, Save, Crown } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { apiClient } from '../../services/apiClient';
import { ENDPOINTS } from '../../utils/endpoints';
import AvatarSelector from '../../components/dashboard/AvatarSelector';

export default function SettingsPage() {
  const { user } = useAuth();
  const [name, setName] = useState(user.fullName || 'Gokulnaath N');
  const [email, setEmail] = useState(user.email || 'gokul@example.com');
  const [timezone, setTimezone] = useState(user.timezone || 'Asia/Kolkata (GMT +5:30)');
  const [theme, setTheme] = useState('system');
  const [gender, setGender] = useState(localStorage.getItem('onboardingGender') || 'female');
  const [isSaving, setIsSaving] = useState(false);
  const { addToast } = useToast();

  const handleSaveProfile = async (e) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      await apiClient.patch(ENDPOINTS.PROFILE.SETTINGS, {
        fullName: name,
        timezone,
        theme,
      });
      addToast('Profile settings updated successfully!', 'success');
    } catch (err) {
      addToast('Failed to update settings', 'error');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6 shadow-[0_0_10px_rgba(79,70,229,0.1)]">
        <h1 className="font-display text-xl font-extrabold text-white">Profile & Settings</h1>
        <p className="text-xs text-slate-400 mt-0.5">Manage your personal details, learning constraints, and 3D companion customization.</p>
      </div>

      {/* Learning Avatar Selector */}
      <div>
        <AvatarSelector key={gender} onSave={() => addToast(`Learning avatar updated successfully!`, 'success')} />
      </div>

      {/* Account Info Form */}
      <form onSubmit={handleSaveProfile} className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-3xl p-6 sm:p-8 shadow-[0_0_20px_rgba(79,70,229,0.15)] space-y-6">
        <div className="flex items-center gap-4 pb-6 border-b border-white/5">
          <img
            src={user.avatarUrl}
            alt={user.fullName}
            className="w-16 h-16 rounded-2xl object-cover ring-2 ring-blue-100"
          />
          <div>
            <h3 className="text-sm font-bold text-white">{name}</h3>
            <p className="text-xs text-slate-400">{email}</p>
            <span className="inline-block mt-1 px-2.5 py-0.5 bg-indigo-900/40 backdrop-blur-sm text-indigo-400 font-bold rounded text-[10px] uppercase">
              Role: {user.role}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold text-white mb-1">Full Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2.5 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-white mb-1">Email Address</label>
            <input
              type="email"
              disabled
              value={email}
              className="w-full p-2.5 text-xs border border-white/10 bg-black/30 backdrop-blur-md text-slate-400 rounded-xl cursor-not-allowed"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-white mb-1">Timezone</label>
            <select
              value={timezone}
              onChange={(e) => setTimezone(e.target.value)}
              className="w-full p-2.5 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600 bg-black/40 backdrop-blur-xl text-slate-300"
            >
              <option value="Asia/Kolkata (GMT +5:30)">Asia/Kolkata (GMT +5:30)</option>
              <option value="America/New_York (GMT -5:00)">America/New_York (GMT -5:00)</option>
              <option value="Europe/London (GMT +0:00)">Europe/London (GMT +0:00)</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-white mb-1">Gender</label>
            <select
              value={gender}
              onChange={(e) => {
                const newGender = e.target.value;
                setGender(newGender);
                localStorage.setItem('onboardingGender', newGender);
                addToast('Gender updated! Please review your avatar options.', 'success');
              }}
              className="w-full p-2.5 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600 bg-black/40 backdrop-blur-xl text-slate-300"
            >
              <option value="female">Female</option>
              <option value="male">Male</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-white mb-1">Interface Theme</label>
            <select
              value={theme}
              onChange={(e) => setTheme(e.target.value)}
              className="w-full p-2.5 text-xs border border-white/10 rounded-xl outline-none focus:ring-2 focus:ring-blue-600 bg-black/40 backdrop-blur-xl text-slate-300"
            >
              <option value="system">System Default</option>
              <option value="light">Light Mode</option>
              <option value="dark">Dark Mode</option>
            </select>
          </div>
        </div>

        <div className="pt-4 flex justify-end">
          <button
            type="submit"
            disabled={isSaving}
            className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold shadow-[0_0_10px_rgba(79,70,229,0.1)] transition flex items-center gap-2"
          >
            <Save className="w-4 h-4" />
            <span>{isSaving ? 'Saving...' : 'Save Settings'}</span>
          </button>
        </div>
      </form>

      {/* Danger Zone */}
      <div className="bg-red-50/50 border border-red-200 rounded-3xl p-6 shadow-[0_0_10px_rgba(79,70,229,0.1)]">
        <h3 className="text-xs font-bold text-red-800 uppercase tracking-wider mb-1">Danger Zone</h3>
        <p className="text-xs text-red-600 mb-4">
          Permanently delete your account and remove all verified evidence records.
        </p>
        <button
          onClick={() => addToast('Account deletion protection triggered (Demo mode).', 'warning')}
          className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-1.5"
        >
          <Trash2 className="w-3.5 h-3.5" /> Delete Account
        </button>
      </div>
    </div>
  );
}
