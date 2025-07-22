import { useState, useEffect } from 'react';
import { apiClient } from '../services/api';

const getDefaultRedirectUrl = () => {
  const host = import.meta.env.VITE_HOST || 'http://localhost';
  const port = import.meta.env.VITE_PORT || '8080';
  return `${host}:${port}/api/v1/callback`;
};

// SVG icons for eye (open/closed)
const EyeIcon = ({ open }: { open: boolean }) => (
  open ? (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ verticalAlign: 'middle' }}><ellipse cx="12" cy="12" rx="9" ry="5"/><circle cx="12" cy="12" r="2.5"/></svg>
  ) : (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ verticalAlign: 'middle' }}><ellipse cx="12" cy="12" rx="9" ry="5"/><path d="M3 3l18 18"/><circle cx="12" cy="12" r="2.5"/></svg>
  )
);

export function LinkedInSettings() {
  const [clientId, setClientId] = useState('');
  const [clientSecret, setClientSecret] = useState('');
  const [redirectUrl, setRedirectUrl] = useState(getDefaultRedirectUrl());
  const [saved, setSaved] = useState(false);
  const [showClientId, setShowClientId] = useState(false);
  const [showClientSecret, setShowClientSecret] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiClient.getLinkedInConfig().then(res => {
      if (res.success && res.data) {
        setClientId(res.data.client_id);
        setClientSecret(res.data.client_secret);
        setRedirectUrl(res.data.redirect_url);
      }
    });
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      const res = await apiClient.updateLinkedInConfig({ client_id: clientId, client_secret: clientSecret, redirect_url: redirectUrl });
      if (res.success && res.data) {
        setClientId(res.data.client_id);
        setClientSecret(res.data.client_secret);
        setRedirectUrl(res.data.redirect_url);
        setSaved(true);
        setTimeout(() => setSaved(false), 2000);
      } else {
        setError(res && typeof res === 'object' && 'error' in res ? (res as { error?: string }).error ?? 'Failed to save' : 'Failed to save');
      }
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'message' in err && typeof (err as { message?: string }).message === 'string') {
        setError((err as { message: string }).message);
      } else {
        setError('Failed to save');
      }
    }
  };

  return (
    <div className="linkedin-settings">
      <h3>LinkedIn API Settings</h3>
      <form onSubmit={handleSubmit} className="linkedin-form">
        <div className="form-group">
          <label htmlFor="linkedin-client-id">Client ID</label>
          <div style={{ position: 'relative' }}>
            <input
              id="linkedin-client-id"
              type={showClientId ? 'text' : 'password'}
              value={clientId}
              onChange={e => setClientId(e.target.value)}
              placeholder="Enter LinkedIn Client ID"
              required
              autoComplete="off"
              style={{ paddingRight: 40 }}
            />
            <button
              type="button"
              onClick={() => setShowClientId(v => !v)}
              style={{
                position: 'absolute',
                right: 12,
                top: '50%',
                transform: 'translateY(-50%)',
                zIndex: 2,
                background: 'none',
                border: 'none',
                padding: 0,
                margin: 0,
                cursor: 'pointer',
                height: 28,
                width: 28,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                lineHeight: 1,
              }}
              tabIndex={-1}
              aria-label={showClientId ? 'Hide Client ID' : 'Show Client ID'}
            >
              <EyeIcon open={showClientId} />
            </button>
          </div>
        </div>
        <div className="form-group">
          <label htmlFor="linkedin-client-secret">Client Secret</label>
          <div style={{ position: 'relative' }}>
            <input
              id="linkedin-client-secret"
              type={showClientSecret ? 'text' : 'password'}
              value={clientSecret}
              onChange={e => setClientSecret(e.target.value)}
              placeholder="Enter LinkedIn Client Secret"
              required
              autoComplete="off"
              style={{ paddingRight: 40 }}
            />
            <button
              type="button"
              onClick={() => setShowClientSecret(v => !v)}
              style={{
                position: 'absolute',
                right: 12,
                top: '50%',
                transform: 'translateY(-50%)',
                zIndex: 2,
                background: 'none',
                border: 'none',
                padding: 0,
                margin: 0,
                cursor: 'pointer',
                height: 28,
                width: 28,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                lineHeight: 1,
              }}
              tabIndex={-1}
              aria-label={showClientSecret ? 'Hide Client Secret' : 'Show Client Secret'}
            >
              <EyeIcon open={showClientSecret} />
            </button>
          </div>
        </div>
        <div className="form-group">
          <label htmlFor="linkedin-redirect-url">Redirect URL</label>
          <input
            id="linkedin-redirect-url"
            type="text"
            value={redirectUrl}
            onChange={e => setRedirectUrl(e.target.value)}
            placeholder="Redirect URL"
            required
          />
          <small className="form-help">Default: {getDefaultRedirectUrl()}</small>
        </div>
        <button type="submit">Save Settings</button>
        {saved && <span style={{ color: 'green', marginLeft: 8 }}>Saved!</span>}
        {error && <span style={{ color: 'red', marginLeft: 8 }}>{error}</span>}
      </form>
    </div>
  );
} 