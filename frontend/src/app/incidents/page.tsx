'use client';

import { useEffect, useState, useCallback, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams, useRouter } from 'next/navigation';
import { incidentsApi } from '@/lib/api';
import type { Incident, Status, Severity } from '@/lib/types';
import { StatusBadge, SeverityBadge } from '@/components/Badges';

function IncidentsPageInner() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [filters, setFilters] = useState({
    status: searchParams.get('status') || '',
    severity: searchParams.get('severity') || '',
    search: searchParams.get('search') || '',
  });

  const page = parseInt(searchParams.get('page') || '1');
  const limit = 10;

  const fetchIncidents = useCallback(() => {
    setLoading(true);
    setError(null);
    incidentsApi.list({
      status: (filters.status as Status) || undefined,
      severity: (filters.severity as Severity) || undefined,
      search: filters.search || undefined,
      page,
      limit,
    })
      .then((res) => {
        setIncidents(res.items);
        setTotal(res.total);
        setTotalPages(res.total_pages);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [filters, page]);

  useEffect(() => {
    fetchIncidents();
  }, [fetchIncidents]);

  const handleFilterChange = (key: string, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    const params = new URLSearchParams(searchParams.toString());
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    params.delete('page');
    router.push(`/incidents?${params.toString()}`);
  };

  const goToPage = (p: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set('page', String(p));
    router.push(`/incidents?${params.toString()}`);
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4" />
          <div className="h-12 bg-gray-200 rounded" />
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 rounded" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
          Failed to load incidents: {error}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Incidents</h1>
        <Link
          href="/incidents/new"
          className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700"
        >
          New Incident
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow mb-6 p-4">
        <div className="flex flex-wrap gap-4">
          <input
            type="text"
            placeholder="Search..."
            value={filters.search}
            onChange={(e) => handleFilterChange('search', e.target.value)}
            className="flex-1 min-w-[200px] px-3 py-2 border border-gray-300 rounded-md text-sm"
          />
          <select
            value={filters.status}
            onChange={(e) => handleFilterChange('status', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="">All Statuses</option>
            <option value="OPEN">Open</option>
            <option value="INVESTIGATING">Investigating</option>
            <option value="MITIGATED">Mitigated</option>
            <option value="RESOLVED">Resolved</option>
          </select>
          <select
            value={filters.severity}
            onChange={(e) => handleFilterChange('severity', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="">All Severities</option>
            <option value="LOW">Low</option>
            <option value="MEDIUM">Medium</option>
            <option value="HIGH">High</option>
            <option value="CRITICAL">Critical</option>
          </select>
        </div>
      </div>

      {incidents.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500 mb-4">No incidents found.</p>
          <Link
            href="/incidents/new"
            className="text-blue-600 hover:text-blue-800 font-medium"
          >
            Create your first incident →
          </Link>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Title</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Severity</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Assignee</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {incidents.map((incident) => (
                <tr key={incident.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <Link href={`/incidents/${incident.id}`} className="text-blue-600 hover:text-blue-800 font-medium">
                      {incident.title}
                    </Link>
                  </td>
                  <td className="px-6 py-4"><StatusBadge status={incident.status} /></td>
                  <td className="px-6 py-4"><SeverityBadge severity={incident.severity} /></td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {incident.assignee?.name || '—'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {new Date(incident.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-6">
          <button
            onClick={() => goToPage(page - 1)}
            disabled={page <= 1}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="px-3 py-1 text-sm text-gray-700">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => goToPage(page + 1)}
            disabled={page >= totalPages}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

export default function IncidentsPage() {
  return (
    <Suspense fallback={
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4" />
          <div className="h-12 bg-gray-200 rounded" />
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-200 rounded" />
          ))}
        </div>
      </div>
    }>
      <IncidentsPageInner />
    </Suspense>
  );
}

