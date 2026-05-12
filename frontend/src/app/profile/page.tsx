"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import AuthGuard from "@/components/AuthGuard";
import { User, Mail, Phone, Edit3, Save, X, Shield } from "lucide-react";

interface UserData {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  status?: string;
  created_at?: string;
}

export default function ProfilePage() {
  const [user, setUser] = useState<UserData | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [form, setForm] = useState({ first_name: "", last_name: "", phone: "" });
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const userStr = localStorage.getItem("user");
    if (userStr) {
      const parsed = JSON.parse(userStr);
      const normalized: UserData = {
        id: parsed.id || parsed.ID || "",
        email: parsed.email || parsed.Email || "",
        first_name: parsed.first_name || parsed.FirstName || "",
        last_name: parsed.last_name || parsed.LastName || "",
        phone: parsed.phone || parsed.Phone || "",
        status: parsed.status || parsed.Status || "",
        created_at: parsed.created_at || parsed.CreatedAt || "",
      };
      setUser(normalized);
      setForm({
        first_name: normalized.first_name,
        last_name: normalized.last_name,
        phone: normalized.phone || "",
      });
    }
    setLoading(false);
  }, []);

  const handleSave = async () => {
    if (!user) return;
    setSaving(true);
    setError("");
    try {
      const data = await api.profile.update(user.id, form);
      setUser(data.user);
      localStorage.setItem("user", JSON.stringify(data.user));
      setEditing(false);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to update profile");
    } finally {
      setSaving(false);
    }
  };

  const memberSince = user?.created_at
    ? new Date(user.created_at).toLocaleDateString("en-US", {
        month: "long",
        year: "numeric",
      })
    : null;

  return (
    <AuthGuard>
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12 animate-fade-in">
        <h1 className="text-3xl font-bold text-slate-900 mb-8">My Profile</h1>

        {loading ? (
          <div className="card animate-pulse space-y-4">
            <div className="h-8 bg-slate-200 rounded w-1/3" />
            <div className="h-4 bg-slate-100 rounded w-1/2" />
            <div className="h-4 bg-slate-100 rounded w-2/3" />
          </div>
        ) : (
          <div className="space-y-6">
            <div className="card !p-8">
              <div className="flex items-center gap-4 mb-6">
                <div className="w-16 h-16 rounded-2xl bg-primary-100 flex items-center justify-center text-primary-600 font-bold text-2xl">
                  {user?.first_name?.[0]}{user?.last_name?.[0]}
                </div>
                <div>
                  <h2 className="text-xl font-bold text-slate-900">
                    {user?.first_name} {user?.last_name}
                  </h2>
                  <p className="text-sm text-slate-500 flex items-center gap-1">
                    <Shield size={14} />
                    {user?.status || "Member"}
                  </p>
                </div>
              </div>

              {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-sm text-red-700 animate-slide-down">
                  {error}
                </div>
              )}

              {editing ? (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1.5">First Name</label>
                    <input className="input" value={form.first_name} onChange={(e) => setForm({ ...form, first_name: e.target.value })} />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1.5">Last Name</label>
                    <input className="input" value={form.last_name} onChange={(e) => setForm({ ...form, last_name: e.target.value })} />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1.5">Phone</label>
                    <input className="input" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
                  </div>
                  <div className="flex gap-3 pt-2">
                    <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
                      {saving ? (
                        <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      ) : (
                        <Save size={16} />
                      )}
                      Save Changes
                    </button>
                    <button className="btn btn-secondary" onClick={() => setEditing(false)} disabled={saving}>
                      <X size={16} />
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div className="space-y-5">
                  <div className="flex items-center gap-3 py-3 border-b border-slate-100">
                    <Mail size={18} className="text-slate-400 shrink-0" />
                    <div>
                      <div className="text-xs text-slate-500">Email</div>
                      <div className="text-sm font-medium text-slate-900">{user?.email}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 py-3 border-b border-slate-100">
                    <User size={18} className="text-slate-400 shrink-0" />
                    <div>
                      <div className="text-xs text-slate-500">Name</div>
                      <div className="text-sm font-medium text-slate-900">{user?.first_name} {user?.last_name}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 py-3 border-b border-slate-100">
                    <Phone size={18} className="text-slate-400 shrink-0" />
                    <div>
                      <div className="text-xs text-slate-500">Phone</div>
                      <div className="text-sm font-medium text-slate-900">{user?.phone || "Not set"}</div>
                    </div>
                  </div>
                  {memberSince && (
                    <div className="py-3 text-xs text-slate-400">
                      Member since {memberSince}
                    </div>
                  )}
                  <button className="btn btn-primary mt-2" onClick={() => setEditing(true)}>
                    <Edit3 size={16} />
                    Edit Profile
                  </button>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </AuthGuard>
  );
}
