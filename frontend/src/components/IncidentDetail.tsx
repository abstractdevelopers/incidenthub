'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { incidentsApi, commentsApi } from '../lib/api';
import type { Incident, Comment, Severity, Status } from '../lib/types';
import Navbar from './Navbar';

export default function IncidentDetailPage({ incidentId }: { incidentId: string }) {
  const router = useRouter();
  const [incident, setIncident] = useState<Incident | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editSeverity, setEditSeverity] = useState<Severity>('LOW');
  const [editStatus, setEditStatus] = useState<Status>('OPEN');
  const [commentBody, setCommentBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    try {
      const [inc, comms] = await Promise.all([
        incidentsApi.get(incidentId),
        commentsApi.list(incidentId),
      ]);
      setIncident(inc);
      setComments(comms);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load incident');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [incidentId]);

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      const updated = await incidentsApi.update(incidentId, {
        title: editTitle,
        description: editDescription,
        severity: editSeverity,
        status: editStatus,
      });
      setIncident(updated);
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update');
    } finally {
      setSubmitting(false);
    }
  };

  const handleStatusChange = async (newStatus: Status) => {
    try {
      const updated = await incidentsApi.update(incidentId, { status: newStatus });
      setIncident(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update status');
    }
  };

  const handleSeverityChange = async (newSeverity: Severity) => {
    try {
      const updated = await incidentsApi.update(incidentId, { severity: newSeverity });
      setIncident(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update severity');
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentBody.trim()) return;
    try {
      const newComment = await commentsApi.create(incidentId, { body: commentBody });
      setComments([...comments, newComment]);
      setCommentBody('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add comment');
    }
  };

  const handleDelete = async () => {
    if (!confirm('Delete this incident? This cannot be undone.')) return;
    try {
      await incidentsApi.delete(incidentId);
      router.push('/dashboard');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  if (loading) {
    return (
      <div className="min-h-full">
        <Navbar />
        <div className="flex items-center justify-center py-24">
          <div className="text-zinc-400">Loading incident...</div>
        </div>
      </div>
    );
  }

  if (error && !incident) {
    return (
      <div className="min-h-full">
        <Navbar />
        <div className="mx-auto max-w-4xl px-4 py-12">
          <div className="rounded-md bg-red-500/10 border border-red-500/20 p-4 text-red-400">
            {error}
          </div>
          <Link href="/dashboard" className="btn btn-secondary mt-4">
            Back to Dashboard
          </Link>
        </div>
      </div>
    );
  }

  if (!incident) return null;

  return (
    <div className="min-h-full">
      <Navbar />
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <Link href="/dashboard" className="btn btn-ghost mb-6 text-sm">
          ← Back to Dashboard
        </Link>

        {error && (
          <div className="mb-4 rounded-md bg-red-500/10 border border-red-500/20 p-3 text-sm text-red-400">
            {error}
          </div>
        )}

        <div className="card">
          {editing ? (
            <form onSubmit={handleUpdate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">Title</label>
                <input type="text" value={editTitle} onChange={(e) => setEditTitle(e.target.value)} className="input" required />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">Description</label>
                <textarea value={editDescription} onChange={(e) => setEditDescription(e.target.value)} className="textarea" required />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-1">Severity</label>
                  <select value={editSeverity} onChange={(e) => setEditSeverity(e.target.value as Severity)} className="select">
                    <option value="LOW">Low</option>
                    <option value="MEDIUM">Medium</option>
                    <option value="HIGH">High</option>
                    <option value="CRITICAL">Critical</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-300 mb-1">Status</label>
                  <select value={editStatus} onChange={(e) => setEditStatus(e.target.value as Status)} className="select">
                    <option value="OPEN">Open</option>
                    <option value="INVESTIGATING">Investigating</option>
                    <option value="MITIGATED">Mitigated</option>
                    <option value="RESOLVED">Resolved</option>
                  </select>
                </div>
              </div>
              <div className="flex gap-2 justify-end">
                <button type="button" onClick={() => setEditing(false)} className="btn btn-secondary">Cancel</button>
                <button type="submit" disabled={submitting} className="btn btn-primary">
                  {submitting ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          ) : (
            <div>
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h1 className="text-2xl font-bold text-white">{incident.title}</h1>
                  <p className="text-zinc-400 mt-1">{incident.description}</p>
                </div>
                <button onClick={() => { setEditing(true); setEditTitle(incident.title); setEditDescription(incident.description); setEditSeverity(incident.severity); setEditStatus(incident.status); }} className="btn btn-secondary text-sm shrink-0">
                  Edit
                </button>
              </div>

              <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
                <div>
                  <p className="text-sm text-zinc-500">Status</p>
                  <select value={incident.status} onChange={(e) => handleStatusChange(e.target.value as Status)} className="select mt-1">
                    <option value="OPEN">Open</option>
                    <option value="INVESTIGATING">Investigating</option>
                    <option value="MITIGATED">Mitigated</option>
                    <option value="RESOLVED">Resolved</option>
                  </select>
                </div>
                <div>
                  <p className="text-sm text-zinc-500">Severity</p>
                  <select value={incident.severity} onChange={(e) => handleSeverityChange(e.target.value as Severity)} className="select mt-1">
                    <option value="LOW">Low</option>
                    <option value="MEDIUM">Medium</option>
                    <option value="HIGH">High</option>
                    <option value="CRITICAL">Critical</option>
                  </select>
                </div>
                <div>
                  <p className="text-sm text-zinc-500">Assignee</p>
                  <p className="mt-1 text-sm text-zinc-300">{incident.assignee?.name || 'Unassigned'}</p>
                </div>
                <div>
                  <p className="text-sm text-zinc-500">Created</p>
                  <p className="mt-1 text-sm text-zinc-300">{new Date(incident.created_at).toLocaleString()}</p>
                </div>
              </div>

              <div className="mt-6 flex gap-2">
                <button onClick={handleDelete} className="btn btn-danger text-sm">Delete Incident</button>
              </div>
            </div>
          )}
        </div>

        {/* Comments */}
        <div className="card mt-6">
          <h2 className="text-lg font-semibold text-white mb-4">Comments ({comments.length})</h2>
          {comments.length === 0 ? (
            <p className="text-zinc-400 text-sm">No comments yet</p>
          ) : (
            <div className="space-y-4 mb-6">
              {comments.map((comment) => (
                <div key={comment.id} className="rounded-lg bg-zinc-800/50 p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-sm font-medium text-white">{comment.author?.name || 'Unknown'}</span>
                    <span className="text-xs text-zinc-500">{new Date(comment.created_at).toLocaleString()}</span>
                  </div>
                  <p className="text-sm text-zinc-300">{comment.body}</p>
                </div>
              ))}
            </div>
          )}
          <form onSubmit={handleAddComment} className="flex gap-2">
            <input
              type="text"
              value={commentBody}
              onChange={(e) => setCommentBody(e.target.value)}
              className="input flex-1"
              placeholder="Add a comment..."
            />
            <button type="submit" className="btn btn-primary">Post</button>
          </form>
        </div>
      </main>
    </div>
  );
}