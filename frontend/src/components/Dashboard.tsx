'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { dashboardApi } from '@/lib/api';
import type { DashboardStats, Incident } from '@/lib/types';
import { StatusBadge, SeverityBadge } from '@/components/Badges';

const statCards: { label: string; key: keyof DashboardStats; color: string }[] = [
  { label: 'Total', key: 'total', color: 'bg-gray-100 text-gray-800' },
  { label: 'Open', key: 'open', color: 'bg-red-100 text-red-800' },
  { label: 'Investigating', key: 'investigating', color: 'bg-yellow-100 text-yellow-800' },
  { label: 'Critical', key: 'critical', color: 'bg-orange-100 text-orange-800' },
  { label: 'Resolved', key: 'resolved', color: 'bg-green-100 text-green-800' },
];

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recent, setRecent] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    dashboardApi
      .stats()
      .then((res) => {
        setStats(res.stats);
        setRecent(res.recent);
      })
      .catch((err) => setError(err.message || 'Failed to load dashboard'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4" />
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-24 bg-gray-200 rounded" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">{error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <Link
          href="/incidents/new"
          className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700"
        >
          Create incident
        </Link>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
        {statCards.map((card) => (
          <div key={card.key} className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">{card.label}</div>
            <div className={`mt-1 text-2xl font-bold ${card.color.split(' ')[1]}`}>
              {stats?.[card.key] ?? 0}
            </div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Recent incidents</h2>
          <Link href="/incidents" className="text-sm text-blue-600 hover:text-blue-800">
            View all
          </Link>
        </div>
        <div className="divide-y divide-gray-200">
          {recent.length === 0 && (
            <div className="px-6 py-8 text-center text-gray-500">No incidents yet.</div>
          )}
          {recent.map((incident) => (
            <Link
              key={incident.id}
              href={`/incidents/${incident.id}`}
              className="block px-6 py-4 hover:bg-gray-50"
            >
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-sm font-medium text-gray-900">{incident.title}</div>
                  <div className="text-xs text-gray-500 mt-1">
                    {new Date(incident.created_at).toLocaleString()}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <SeverityBadge severity={incident.severity} />
                  <StatusBadge status={incident.status} />
                </div>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
