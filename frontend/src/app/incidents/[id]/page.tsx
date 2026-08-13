'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import IncidentDetailPage from '../../components/IncidentDetail';

export default function IncidentPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const [id, setId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      const p = await params;
      setId(p.id);
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      if (!token) {
        router.push('/');
      }
      setLoading(false);
    })();
  }, [params, router]);

  if (loading || !id) {
    return (
      <div className="flex min-h-full items-center justify-center">
        <div className="text-zinc-400">Loading...</div>
      </div>
    );
  }

  return <IncidentDetailPage incidentId={id} />;
}