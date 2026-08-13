'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { incidentsApi } from '@/lib/api';
import type { Severity, Status } from '@/lib/types';

export default function NewIncidentPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    title: '',
    description: '',
    severity: 'MEDIUM' as Severity,
    status: 'OPEN' as Status,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const incident = await incidentsApi.create(form);
      router.push(`/incidents/${incident.id}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">New Incident</h1>

      {error && (
        <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow p-6 space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
          <input
            type="text"
            required
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
            placeholder="e.g. Production database down"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea
            required
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            rows={5}
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm resize-none"
            placeholder="Describe the incident in detail..."
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Severity</label>
            <select
              value={form.severity}
              onChange={(e) => setForm({ ...form, severity: e.target.value as Severity })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
            >
              <option value="LOW">Low</option>
              <option value="MEDIUM">Medium</option>
              <option value="HIGH">High</option>
              <option value="CRITICAL">Critical</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
            <select
              value={form.status}
              onChange={(e) => setForm({ ...form, status: e.target.value as Status })}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
            >
              <option value="OPEN">Open</option>
              <option value="INVESTIGATING">Investigating</option>
              <option value="MITIGATED">Mitigated</option>
              <option value="RESOLVED">Resolved</option>
            </select>
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Creating...' : 'Create Incident'}
          </button>
        </div>
      </form>
    </div>
  );
}