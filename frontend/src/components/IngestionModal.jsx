import React, { useState } from 'react';
import { Folder, GitBranch, Key, RefreshCw, X, ShieldAlert, CheckCircle2, ArrowRight } from 'lucide-react';
import { api } from '../api/client';

export function IngestionModal({ isOpen, onClose, onStartLocal, onFetchGitHubRepos, onStartGitHub, loadingRepos, githubRepos, scanning }) {
  const [activeTab, setActiveTab] = useState('local'); // 'local' | 'github'
  const [localPath, setLocalPath] = useState('sample_target');
  const [targetName, setTargetName] = useState('');

  // GitHub App state
  const [installationId, setInstallationId] = useState(() => localStorage.getItem('cbom_github_installation_id') || '');
  const [selectedRepo, setSelectedRepo] = useState(null);
  const [githubFetched, setGithubFetched] = useState(false);
  const [modalError, setModalError] = useState(null);

  React.useEffect(() => {
    const savedId = localStorage.getItem('cbom_github_installation_id');
    const idToUse = savedId || installationId;
    if (activeTab === 'github' && idToUse && (!githubRepos || githubRepos.length === 0) && !githubFetched && !loadingRepos) {
      setModalError(null);
      onFetchGitHubRepos(idToUse)
        .then(() => setGithubFetched(true))
        .catch((err) => {
          setModalError('GitHub Installation token expired or revoked. Please click "Install from GitHub" below.');
          localStorage.removeItem('cbom_github_installation_id');
          setInstallationId('');
          setGithubFetched(true);
        });
    }
  }, [activeTab, installationId, githubRepos, githubFetched, loadingRepos, onFetchGitHubRepos]);

  if (!isOpen) return null;

  const handleLocalSubmit = async (e) => {
    e.preventDefault();
    if (!localPath) return;
    try {
      await onStartLocal(localPath, targetName || localPath);
      onClose();
    } catch (err) {
      // Error handled in parent hook
    }
  };

  const handleFetchRepos = async (e) => {
    e.preventDefault();
    if (!installationId) return;
    setModalError(null);
    try {
      await onFetchGitHubRepos(installationId);
      setGithubFetched(true);
    } catch (err) {
      setModalError(err.message);
    }
  };

  const handleGitHubScanSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRepo) return;
    try {
      await onStartGitHub(selectedRepo, installationId);
      onClose();
    } catch (err) {
      // Error handled in hook
    }
  };

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(15, 23, 42, 0.8)',
      backdropFilter: 'blur(8px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: '24px',
    }}>
      <div className="glass-card" style={{
        width: '100%',
        maxWidth: 'min(680px, 95vw)',
        maxHeight: '92vh',
        overflowY: 'auto',
        borderRadius: '16px',
        border: '1px solid rgba(255, 255, 255, 0.15)',
        backgroundColor: '#0f172a',
        color: '#f8fafc',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
      }}>
        {/* Header */}
        <div style={{
          padding: '20px 24px',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <div>
            <h2 style={{ fontSize: '20px', fontWeight: '700', margin: 0, color: '#f8fafc', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <RefreshCw className={scanning ? 'animate-spin' : ''} style={{ color: '#38bdf8', width: '22px', height: '22px' }} />
              Codebase Ingestion Engine
            </h2>
            <p style={{ fontSize: '13px', color: '#94a3b8', margin: '4px 0 0 0' }}>
              Select target source codebase (Local Filesystem Directory or Remote GitHub Repository via GitHub App)
            </p>
          </div>
          <button onClick={onClose} style={{
            background: 'none',
            border: 'none',
            color: '#64748b',
            cursor: 'pointer',
            padding: '4px',
            borderRadius: '6px',
          }}>
            <X style={{ width: '20px', height: '20px' }} />
          </button>
        </div>

        {/* Tab Selection */}
        <div style={{
          display: 'flex',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          backgroundColor: 'rgba(30, 41, 59, 0.5)',
        }}>
          <button
            onClick={() => setActiveTab('local')}
            style={{
              flex: 1,
              padding: '14px 16px',
              border: 'none',
              background: activeTab === 'local' ? '#1e293b' : 'transparent',
              color: activeTab === 'local' ? '#38bdf8' : '#94a3b8',
              fontWeight: '600',
              fontSize: '14px',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '8px',
              borderBottom: activeTab === 'local' ? '2px solid #38bdf8' : '2px solid transparent',
            }}
          >
            <Folder style={{ width: '18px', height: '18px' }} />
            Local Directory Codebase
          </button>

          <button
            onClick={() => setActiveTab('github')}
            style={{
              flex: 1,
              padding: '14px 16px',
              border: 'none',
              background: activeTab === 'github' ? '#1e293b' : 'transparent',
              color: activeTab === 'github' ? '#38bdf8' : '#94a3b8',
              fontWeight: '600',
              fontSize: '14px',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '8px',
              borderBottom: activeTab === 'github' ? '2px solid #38bdf8' : '2px solid transparent',
            }}
          >
            <GitBranch style={{ width: '18px', height: '18px' }} />
            GitHub App Repository
          </button>
        </div>

        {/* Modal Body */}
        <div style={{ padding: '24px' }}>
          {activeTab === 'local' ? (
            <form onSubmit={handleLocalSubmit}>
              <div style={{ marginBottom: '16px' }}>
                <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '6px' }}>
                  Target Directory Path
                </label>
                <input
                  type="text"
                  value={localPath}
                  onChange={(e) => setLocalPath(e.target.value)}
                  placeholder="e.g. sample_target or C:\projects\payment-api"
                  required
                  style={{
                    width: '100%',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    border: '1px solid rgba(255, 255, 255, 0.15)',
                    backgroundColor: '#1e293b',
                    color: '#f8fafc',
                    fontSize: '14px',
                    outline: 'none',
                  }}
                />
                <span style={{ fontSize: '12px', color: '#64748b', marginTop: '4px', display: 'block' }}>
                  Scans source code AST, certificates, keystores, IaC files, secrets, and container specs.
                </span>
              </div>

              <div style={{ marginBottom: '24px' }}>
                <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '6px' }}>
                  Target Display Name (Optional)
                </label>
                <input
                  type="text"
                  value={targetName}
                  onChange={(e) => setTargetName(e.target.value)}
                  placeholder="e.g. Payment Microservice"
                  style={{
                    width: '100%',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    border: '1px solid rgba(255, 255, 255, 0.15)',
                    backgroundColor: '#1e293b',
                    color: '#f8fafc',
                    fontSize: '14px',
                    outline: 'none',
                  }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
                <button
                  type="button"
                  onClick={onClose}
                  style={{
                    padding: '10px 18px',
                    borderRadius: '8px',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    backgroundColor: 'transparent',
                    color: '#94a3b8',
                    cursor: 'pointer',
                    fontWeight: '600',
                  }}
                >
                  Cancel
                </button>

                <button
                  type="submit"
                  disabled={scanning}
                  style={{
                    padding: '10px 20px',
                    borderRadius: '8px',
                    border: 'none',
                    backgroundColor: '#0284c7',
                    color: '#ffffff',
                    fontWeight: '600',
                    cursor: scanning ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                  }}
                >
                  {scanning ? 'Scanning Codebase...' : 'Start Local Scan'}
                  <ArrowRight style={{ width: '16px', height: '16px' }} />
                </button>
              </div>
            </form>
          ) : (
            <div>
              {modalError && (
                <div style={{
                  marginBottom: '16px',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  backgroundColor: 'rgba(239, 68, 68, 0.15)',
                  border: '1px solid rgba(239, 68, 68, 0.3)',
                  color: '#fca5a5',
                  fontSize: '13px',
                  lineHeight: '1.4',
                }}>
                  <strong>GitHub Authorization Error:</strong> {modalError}
                </div>
              )}

              {/* State 1: Loading Repositories */}
              {loadingRepos ? (
                <div style={{
                  backgroundColor: 'rgba(30, 41, 59, 0.6)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  padding: '40px 20px',
                  textAlign: 'center',
                }}>
                  <RefreshCw className="animate-spin" style={{ color: '#38bdf8', width: '32px', height: '32px', margin: '0 auto 12px auto' }} />
                  <div style={{ fontSize: '15px', fontWeight: '600', color: '#f8fafc' }}>
                    Fetching Authorized Repositories...
                  </div>
                  <div style={{ fontSize: '13px', color: '#94a3b8', marginTop: '4px' }}>
                    Exchanging GitHub App credentials for active repository access tokens.
                  </div>
                </div>
              ) : githubRepos && githubRepos.length > 0 ? (
                /* State 2: Show Repositories List */
                <form onSubmit={handleGitHubScanSubmit}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                    <label style={{ fontSize: '13px', fontWeight: '600', color: '#cbd5e1' }}>
                      Select Repository to Scan ({githubRepos.length} Repositories Available)
                    </label>

                    <button
                      type="button"
                      onClick={async () => {
                        try {
                          const res = await api.getGitHubInstallUrl();
                          if (res && res.install_url) {
                            const width = 600;
                            const height = 750;
                            const left = window.screenX + (window.outerWidth - width) / 2;
                            const top = window.screenY + (window.outerHeight - height) / 2;
                            window.open(
                              res.install_url,
                              'GitHubAppAuthorization',
                              `width=${width},height=${height},left=${left},top=${top},status=yes,scrollbars=yes`
                            );
                          }
                        } catch (err) {
                          setModalError('Failed to load GitHub Install URL: ' + err.message);
                        }
                      }}
                      style={{
                        background: 'none',
                        border: 'none',
                        color: '#38bdf8',
                        fontSize: '12px',
                        fontWeight: '600',
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px',
                      }}
                    >
                      + Install on another account / Add Repos
                    </button>
                  </div>

                  <div style={{
                    maxHeight: '260px',
                    overflowY: 'auto',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '10px',
                    backgroundColor: '#1e293b',
                    marginBottom: '20px',
                  }}>
                    {githubRepos.map((repo) => (
                      <div
                        key={repo.id}
                        onClick={() => setSelectedRepo(repo)}
                        style={{
                          padding: '12px 16px',
                          borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          cursor: 'pointer',
                          backgroundColor: selectedRepo?.id === repo.id ? 'rgba(56, 189, 248, 0.15)' : 'transparent',
                          transition: 'background 0.15s ease',
                        }}
                      >
                        <div>
                          <div style={{ fontWeight: '600', color: '#f8fafc', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <GitBranch style={{ width: '15px', height: '15px', color: '#38bdf8' }} />
                            {repo.full_name}
                          </div>
                          <div style={{ fontSize: '12px', color: '#64748b', marginTop: '2px' }}>
                            Default Branch: <span style={{ color: '#cbd5e1' }}>{repo.default_branch || 'main'}</span> {repo.private ? '• Private' : '• Public'}
                          </div>
                        </div>
                        {selectedRepo?.id === repo.id && (
                          <CheckCircle2 style={{ color: '#38bdf8', width: '18px', height: '18px' }} />
                        )}
                      </div>
                    ))}
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px' }}>
                    <button
                      type="button"
                      onClick={onClose}
                      style={{
                        padding: '10px 18px',
                        borderRadius: '8px',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        backgroundColor: 'transparent',
                        color: '#94a3b8',
                        cursor: 'pointer',
                        fontWeight: '600',
                      }}
                    >
                      Cancel
                    </button>

                    <button
                      type="submit"
                      disabled={!selectedRepo || scanning}
                      style={{
                        padding: '10px 20px',
                        borderRadius: '8px',
                        border: 'none',
                        backgroundColor: !selectedRepo ? '#334155' : '#0284c7',
                        color: '#ffffff',
                        fontWeight: '600',
                        cursor: (!selectedRepo || scanning) ? 'not-allowed' : 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                      }}
                    >
                      {scanning ? 'Cloning & Scanning...' : 'Scan Selected Repository'}
                      <ArrowRight style={{ width: '16px', height: '16px' }} />
                    </button>
                  </div>
                </form>
              ) : (
                /* State 3: Show "Install from GitHub" 1-Click Button */
                <div style={{
                  backgroundColor: 'rgba(30, 41, 59, 0.6)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  padding: '28px 20px',
                  textAlign: 'center',
                }}>
                  <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '14px' }}>
                    <svg height="42" viewBox="0 0 16 16" version="1.1" width="42" fill="#f8fafc">
                      <path fillRule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.28.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
                    </svg>
                  </div>
                  <h3 style={{ fontSize: '18px', fontWeight: '700', color: '#f8fafc', marginBottom: '6px' }}>
                    Sign in & Install App on GitHub
                  </h3>
                  <p style={{ fontSize: '13px', color: '#94a3b8', marginBottom: '20px', maxWidth: '460px', margin: '0 auto 20px auto', lineHeight: '1.5' }}>
                    Click below to sign in to your GitHub account or organization and authorize the QuantumVault scanner app. Once signed in, your authorized repositories will automatically list here.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '12px' }}>
                    <button
                      type="button"
                      onClick={async () => {
                        try {
                          const res = await api.getGitHubInstallUrl();
                          if (res && res.install_url) {
                            const width = 600;
                            const height = 750;
                            const left = window.screenX + (window.outerWidth - width) / 2;
                            const top = window.screenY + (window.outerHeight - height) / 2;
                            window.open(
                              res.install_url,
                              'GitHubAppAuthorization',
                              `width=${width},height=${height},left=${left},top=${top},status=yes,scrollbars=yes`
                            );
                          }
                        } catch (err) {
                          setModalError('Failed to load GitHub Install URL: ' + err.message);
                        }
                      }}
                      style={{
                        padding: '12px 24px',
                        borderRadius: '10px',
                        border: 'none',
                        backgroundColor: '#16a34a',
                        color: '#ffffff',
                        fontWeight: '700',
                        fontSize: '14px',
                        cursor: 'pointer',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '10px',
                        boxShadow: '0 4px 14px rgba(22, 163, 74, 0.4)',
                      }}
                    >
                      <svg height="20" viewBox="0 0 16 16" version="1.1" width="20" fill="#ffffff">
                        <path fillRule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.28.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
                      </svg>
                      Install App on GitHub
                    </button>

                    <a
                      href="https://github.com/logout"
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{
                        fontSize: '12px',
                        color: '#38bdf8',
                        textDecoration: 'underline',
                        cursor: 'pointer',
                        marginTop: '4px',
                      }}
                    >
                      Need to log into a different GitHub account? Log out on GitHub first
                    </a>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
