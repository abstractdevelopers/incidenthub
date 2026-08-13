'use client';

import { useState } from 'react';
import { authApi } from '@/lib/api';

export default function LoginPage() {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (isLogin) {
        const response = await authApi.login({ email, password });
        localStorage.setItem('token', response.token);
        localStorage.setItem('user', JSON.stringify({ id: response.id, email: response.email, name: response.name }));
        window.location.href = '/dashboard';
      } else {
        if (!name.trim()) {
          setError('Name is required');
          setLoading(false);
          return;
        }
        await authApi.register({ email, password, name });
        const loginResponse = await authApi.login({ email, password });
        localStorage.setItem('token', loginResponse.token);
        localStorage.setItem('user', JSON.stringify({ id: loginResponse.id, email: loginResponse.email, name: loginResponse.name }));
        window.location.href = '/dashboard';
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-full items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white">IncidentHub</h1>
          <p className="text-zinc-400 mt-2">Track, manage, and resolve incidents</p>
        </div>

        <div className="card">
          <div className="flex gap-2 mb-6">
            <button
              className={`btn flex-1 ${isLogin ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setIsLogin(true)}
            >
              Login
            </button>
            <button
              className={`btn flex-1 ${!isLogin ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setIsLogin(false)}
            >
              Register
            </button>
          </div>

          {error && (
            <div className="mb-4 rounded-md bg-red-500/10 border border-red-500/20 p-3 text-sm text-red-400">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {!isLogin && (
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-1">Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="input"
                  placeholder="Your name"
                />
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="input"
                placeholder="you@example.com"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="input"
                placeholder="Min 6 characters"
                required
                minLength={6}
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="btn btn-primary w-full"
            >
              {loading ? 'Please wait...' : isLogin ? 'Login' : 'Register'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}