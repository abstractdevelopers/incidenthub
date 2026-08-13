'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import DashboardPage from '@/components/Dashboard';

export default function Dashboard() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    if (!token) {
      router.push('/');
    }
    setLoading(false);
  }, [router]);

  if (loading) {
    return (
      <div className="flex min-h-full items-center justify-center">
        <div className="text-zinc-400">Loading...</div>
      </div>
    );
  }

  return <DashboardPage />;
}