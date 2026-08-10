import React, { useState } from 'react';
import { Folder, GitBranch, Key, RefreshCw, X, ShieldAlert, CheckCircle2, ArrowRight } from 'lucide-react';

export function IngestionModal({ isOpen, onClose, onStartLocal, onFetchGitHubRepos, onStartGitHub, loadingRepos, githubRepos, scanning }) {
  const [activeTab, setActiveTab] = useState('local'); // 'local' | 'github'
  const [localPath, setLocalPath] = useState('sample_target');
  const [targetName, setTargetName] = useState('');

  // GitHub App state
  const [installationId, setInstallationId] = useState('123456');
  const [selectedRepo, setSelectedRepo] = useState(null);
  const [githubFetched, setGithubFetched] = useState(false);
  const [modalError, setModalError] = useState(null);

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
              {/* Step 1: Fetch GitHub Repos */}
              <form onSubmit={handleFetchRepos} style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '6px' }}>
                  GitHub App Installation ID
                </label>
                <div style={{ display: 'flex', gap: '12px' }}>
                  <div style={{ position: 'relative', flex: 1 }}>
                    <Key style={{ position: 'absolute', left: '12px', top: '12px', width: '16px', height: '16px', color: '#64748b' }} />
                    <input
                      type="text"
                      value={installationId}
                      onChange={(e) => setInstallationId(e.target.value)}
                      placeholder="Enter GitHub App Installation ID (e.g. 123456)"
                      required
                      style={{
                        width: '100%',
                        padding: '10px 14px 10px 36px',
                        borderRadius: '8px',
                        border: '1px solid rgba(255, 255, 255, 0.15)',
                        backgroundColor: '#1e293b',
                        color: '#f8fafc',
                        fontSize: '14px',
                        outline: 'none',
                      }}
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={loadingRepos}
                    style={{
                      padding: '10px 18px',
                      borderRadius: '8px',
                      border: 'none',
                      backgroundColor: '#334155',
                      color: '#f8fafc',
                      fontWeight: '600',
                      cursor: loadingRepos ? 'not-allowed' : 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '8px',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    <RefreshCw className={loadingRepos ? 'animate-spin' : ''} style={{ width: '16px', height: '16px' }} />
                    {loadingRepos ? 'Fetching...' : 'Fetch Repositories'}
                  </button>
                </div>

                {modalError && (
                  <div style={{
                    marginTop: '12px',
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
                <span style={{ fontSize: '12px', color: '#64748b', marginTop: '4px', display: 'block' }}>
                  Uses RS256 App JWT to exchange for short-lived Installation Access Tokens.
                </span>
              </form>

              {/* Step 2: Repository Picker */}
              {githubFetched && (
                <form onSubmit={handleGitHubScanSubmit}>
                  <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '8px' }}>
                    Select Repository to Scan ({githubRepos.length} Available)
                  </label>

                  <div style={{
                    maxHeight: '200px',
                    overflowY: 'auto',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '8px',
                    backgroundColor: '#1e293b',
                    marginBottom: '20px',
                  }}>
                    {githubRepos.length === 0 ? (
                      <div style={{ padding: '20px', textAlign: 'center', color: '#94a3b8', fontSize: '14px' }}>
                        No accessible repositories found for Installation ID {installationId}.
                      </div>
                    ) : (
                      githubRepos.map((repo) => (
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
                          }}
                        >
                          <div>
                            <div style={{ fontWeight: '600', color: '#f8fafc', fontSize: '14px' }}>
                              {repo.full_name}
                            </div>
                            <div style={{ fontSize: '12px', color: '#64748b' }}>
                              Default Branch: {repo.default_branch || 'main'} {repo.private ? '• Private' : '• Public'}
                            </div>
                          </div>
                          {selectedRepo?.id === repo.id && (
                            <CheckCircle2 style={{ color: '#38bdf8', width: '18px', height: '18px' }} />
                          )}
                        </div>
                      ))
                    )}
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
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
