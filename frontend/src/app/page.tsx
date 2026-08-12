import { useState, useEffect } from 'react';
import { authApi, incidentsApi, commentsApi, dashboardApi } from '../lib/api';
import { User, Incident, Comment, DashboardStats } from '../lib/types';

export default function App() {
  const [user, setUser] = useState<User | null>(null);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [comments, setComments] = useState<Comment[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const token = localStorage.getItem('token');
        if (token) {
          const response = await fetch('/api/auth/me', {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });
          if (response.ok) {
            const userData: User = await response.json();
            setUser(userData);
          }
        }
      } catch (error) {
        console.error(error);
      }
    };
    fetchUser();
  }, []);

  useEffect(() => {
    const fetchIncidents = async () => {
      try {
        const response = await incidentsApi.list();
        setIncidents(response.items);
      } catch (error) {
        console.error(error);
      }
    };
    fetchIncidents();
  }, []);

  useEffect(() => {
    const fetchComments = async () => {
      try {
        const response = await commentsApi.list('incident-123');
        setComments(response);
      } catch (error) {
        console.error(error);
      }
    };
    fetchComments();
  }, []);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await dashboardApi.stats();
        setStats(response.stats);
      } catch (error) {
        console.error(error);
      }
    };
    fetchStats();
  }, []);

  const handleLogin = async (email: string, password: string) => {
    try {
      const response = await authApi.login({ email, password });
      localStorage.setItem('token', response.token);
      setUser(response);
    } catch (error) {
      console.error(error);
    }
  };

  const handleRegister = async (email: string, password: string, name: string) => {
    try {
      const response = await authApi.register({ email, password, name });
      localStorage.setItem('token', response.token);
      setUser(response);
    } catch (error) {
      console.error(error);
    }
  };

  const handleCreateIncident = async (title: string, description: string, severity: string, status: string, assignee_id: string) => {
    try {
      const response = await incidentsApi.create({ title, description, severity, status, assignee_id });
      setIncidents([...incidents, response]);
    } catch (error) {
      console.error(error);
    }
  };

  const handleCreateComment = async (incidentId: string, body: string) => {
    try {
      const response = await commentsApi.create(incidentId, { body });
      setComments([...comments, response]);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <div>
      <header>
        <nav>
          <ul>
            <li>
              <a href="#">Home</a>
            </li>
            <li>
              <a href="#">Incidents</a>
            </li>
            <li>
              <a href="#">Dashboard</a>
            </li>
          </ul>
        </nav>
      </header>
      <main>
        {user ? (
          <div>
            <h1>Welcome, {user.name}!</h1>
            <p>Email: {user.email}</p>
          </div>
        ) : (
          <div>
            <h1>Please log in or register</h1>
            <form onSubmit={(e) => handleLogin('test@example.com', 'password123')}>
              <input type="email" placeholder="Email" />
              <input type="password" placeholder="Password" />
              <button type="submit">Login</button>
            </form>
            <form onSubmit={(e) => handleRegister('new@example.com', 'password123', 'New User')}>
              <input type="email" placeholder="Email" />
              <input type="password" placeholder="Password" />
              <input type="text" placeholder="Name" />
              <button type="submit">Register</button>
            </form>
          </div>
        )}
        <h1>Incidents</h1>
        <ul>
          {incidents.map((incident) => (
            <li key={incident.id}>{incident.title}</li>
          ))}
        </ul>
        <form onSubmit={(e) => handleCreateIncident('New Incident', 'Description', 'LOW', 'OPEN', 'user-123')}>
          <input type="text" placeholder="Title" />
          <textarea placeholder="Description" />
          <select>
            <option value="LOW">Low</option>
            <option value="MEDIUM">Medium</option>
            <option value="HIGH">High</option>
            <option value="CRITICAL">Critical</option>
          </select>
          <select>
            <option value="OPEN">Open</option>
            <option value="INVESTIGATING">Investigating</option>
            <option value="MITIGATED">Mitigated</option>
            <option value="RESOLVED">Resolved</option>
          </select>
          <input type="text" placeholder="Assignee ID" />
          <button type="submit">Create Incident</button>
        </form>
        <h1>Comments</h1>
        <ul>
          {comments.map((comment) => (
            <li key={comment.id}>{comment.body}</li>
          ))}
        </ul>
        <form onSubmit={(e) => handleCreateComment('incident-123', 'New Comment')}>
          <textarea placeholder="Comment" />
          <button type="submit">Create Comment</button>
        </form>
        <h1>Dashboard</h1>
        {stats ? (
          <div>
            <p>Total: {stats.total}</p>
            <p>Open: {stats.open}</p>
            <p>Investigating: {stats.investigating}</p>
            <p>Mitigated: {stats.mitigated}</p>
            <p>Resolved: {stats.resolved}</p>
            <p>Critical: {stats.critical}</p>
          </div>
        ) : (
          <div>
            <p>Loading...</p>
          </div>
        )}
      </main>
    </div>
  );
}