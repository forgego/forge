import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { adminAPI } from '../api/client';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Loader2, Eye, EyeOff, Database, ShieldCheck, Zap } from 'lucide-react';

export default function LoginPage() {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await adminAPI.login({ username, password });
      localStorage.setItem('admin_token', response.token);
      navigate({ to: '/' });
    } catch (err: any) {
      const status = err.response?.status;
      const data = err.response?.data;
      const message =
        data?.error?.message ||
        data?.message ||
        (status === 503
          ? 'Admin login is not configured on the server.'
          : 'Invalid username or password');
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background px-4 py-10">
      <Card className="w-full max-w-3xl overflow-hidden shadow-lg md:grid md:grid-cols-2 p-0">
        {/* Brand panel */}
        <div className="relative hidden md:flex flex-col justify-between overflow-hidden bg-surface-sunken p-8 text-white dark:bg-primary/[0.12] dark:text-foreground dark:border-r dark:border-border">
          <div
            aria-hidden
            className="pointer-events-none absolute -top-24 -right-24 h-64 w-64 rounded-full bg-primary/30 blur-3xl"
          />
          <div
            aria-hidden
            className="pointer-events-none absolute -bottom-28 -left-20 h-72 w-72 rounded-full bg-primary/20 blur-3xl"
          />
          <div className="relative flex items-center gap-2.5">
            <div className="w-9 h-9 rounded-xl bg-primary flex items-center justify-center text-primary-foreground text-lg font-bold shadow-lg">
              F
            </div>
            <span className="text-lg font-bold tracking-tight">Forge Admin</span>
          </div>
          <div className="relative space-y-5">
            <p className="text-xl font-semibold leading-snug tracking-tight">
              Every model, record and audit trail — one quiet console.
            </p>
            <ul className="space-y-3 text-body text-muted-foreground">
              <li className="flex items-start gap-2.5">
                <Database className="h-4 w-4 mt-0.5 shrink-0 text-primary" />
                Schema-driven CRUD with search, filters and bulk actions
              </li>
              <li className="flex items-start gap-2.5">
                <ShieldCheck className="h-4 w-4 mt-0.5 shrink-0 text-primary" />
                Deny-by-default permissions and immutable audit history
              </li>
              <li className="flex items-start gap-2.5">
                <Zap className="h-4 w-4 mt-0.5 shrink-0 text-primary" />
                Command palette, keyboard shortcuts and saved views
              </li>
            </ul>
          </div>
          <p className="relative text-xs text-muted-foreground">
            Type-safe Go backend · React admin UI
          </p>
        </div>

        {/* Form */}
        <div className="p-6 sm:p-8">
          <div className="md:hidden mx-auto mb-4 w-11 h-11 rounded-xl bg-primary flex items-center justify-center text-primary-foreground text-xl font-bold">
            F
          </div>
          <CardHeader className="space-y-1.5 p-0 mb-6">
            <CardTitle className="text-2xl font-bold tracking-tight">Welcome back</CardTitle>
            <CardDescription>
              Sign in to access the admin dashboard
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="username" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                  Username
                </label>
                <Input
                  id="username"
                  data-testid="username-input"
                  type="text"
                  placeholder="admin"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete="username"
                  required
                />
              </div>
              <div className="space-y-2">
                <label htmlFor="password" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                  Password
                </label>
                <div className="relative">
                  <Input
                    id="password"
                    data-testid="password-input"
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    className="pr-10"
                    required
                  />
                  <button
                    type="button"
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword((v) => !v)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              {error && (
                <div role="alert" aria-live="assertive" className="text-sm text-destructive font-medium rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2">
                  {error}
                </div>
              )}
              <Button type="submit" data-testid="login-button" className="w-full h-10" disabled={loading}>
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Signing in...
                  </>
                ) : (
                  'Sign In'
                )}
              </Button>
            </form>
          </CardContent>
        </div>
      </Card>
    </div>
  );
}
