'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from './AuthProvider';

export default function Navbar() {
  const { user, logout } = useAuth();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/');
  };

  return (
    <nav className="border-b border-zinc-800 bg-zinc-900/80 backdrop-blur-sm">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <Link href="/dashboard" className="text-xl font-bold text-white">
            IncidentHub
          </Link>
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className="text-sm text-zinc-300 hover:text-white">
              Dashboard
            </Link>
            {user ? (
              <div className="flex items-center gap-3">
                <span className="text-sm text-zinc-400">{user.name}</span>
                <button onClick={handleLogout} className="btn btn-ghost text-sm">
                  Logout
                </button>
              </div>
            ) : (
              <Link href="/" className="btn btn-primary text-sm">
                Login
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
