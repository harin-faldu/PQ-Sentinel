import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, Outlet, useOutletContext, useNavigate } from 'react-router-dom';
import Sidebar from './components/Sidebar';
import PageHeader from './components/PageHeader';
import { CbomHeaderActions } from './components/CbomHeaderActions';
import { DashboardPage } from './pages/DashboardPage';
import { InventoryPage } from './pages/InventoryPage';
import { RiskAnalysisPage } from './pages/RiskAnalysisPage';
import { MigrationPlannerPage } from './pages/MigrationPlannerPage';
import { PlaceholderPage } from './pages/PlaceholderPage';
import { IngestionModal } from './components/IngestionModal';
import { CycloneDXExportModal } from './components/CycloneDXExportModal';
import { useCbomScan } from './hooks/useCbomScan';
import { useTheme } from './context/ThemeContext';
import { AlertCircle, Info } from 'lucide-react';
import './assets/styles/App.css';

/* ── Main Layout with Sidebar ── */
const MainLayout = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const { isDark } = useTheme();

  return (
    <div className="app">
      {/* Mobile Header */}
      <header className="mobile-header">
        <button
          className="hamburger-toggle"
          onClick={() => setIsSidebarOpen(true)}
        >
          <span></span>
          <span></span>
          <span></span>
        </button>
        <div className="mobile-logo-container">
          <img
            src={isDark ? "/3.svg" : "/4.svg"}
            alt="QuantumVault"
            className="mobile-logo"
          />
        </div>
      </header>

      <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
};

/* ── CBOM Layout (wraps CBOM pages with shared scan state, banners, and modals) ── */
const CbomLayout = () => {
  const [isIngestionOpen, setIsIngestionOpen] = useState(false);
  const [isExportOpen, setIsExportOpen] = useState(false);

  const {
    scans,
    targets,
    currentScan,
    loading,
    scanning,
    error,
    githubRepos,
    loadingRepos,
    selectScan,
    startLocalScan,
    fetchGitHubRepositories,
    startGitHubScan,
    setError,
  } = useCbomScan();

  // GitHub App auth postMessage & redirect handling
  React.useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const instId = params.get('installation_id');

    // 1. If this window was opened or redirected with installation_id
    if (instId) {
      localStorage.setItem('cbom_github_installation_id', instId);

      // Check if running inside a popup window
      const isPopup = (window.opener && window.opener !== window) ||
                      window.name === 'GitHubAppAuthorization' ||
                      (window.outerWidth < 750 && window.outerHeight < 850);

      if (isPopup) {
        try {
          if (window.opener) {
            window.opener.postMessage({ type: 'GITHUB_AUTH_SUCCESS', installation_id: instId }, '*');
          }
        } catch (e) {}
        
        // Attempt to close popup window
        window.close();

        // If window.close() was blocked by browser, render a clean completion card
        document.body.innerHTML = `
          <div style="background:#0f172a;color:#f8fafc;height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;font-family:sans-serif;text-align:center;padding:20px;">
            <div style="background:#1e293b;border:1px solid rgba(255,255,255,0.1);padding:32px;border-radius:16px;max-width:400px;box-shadow:0 20px 25px -5px rgba(0,0,0,0.5);">
              <svg height="48" viewBox="0 0 16 16" width="48" fill="#38bdf8" style="margin-bottom:16px;">
                <path fill-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.28.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
              </svg>
              <h2 style="font-size:20px;font-weight:700;margin:0 0 8px 0;color:#f8fafc;">GitHub App Authorized!</h2>
              <p style="font-size:14px;color:#94a3b8;margin:0 0 20px 0;line-height:1.4;">Your repositories are now accessible in QuantumVault.</p>
              <button onclick="window.close()" style="background:#0284c7;color:#ffffff;border:none;padding:12px 24px;border-radius:10px;font-weight:700;font-size:14px;cursor:pointer;width:100%;">
                Close This Window
              </button>
            </div>
          </div>
        `;
        return;
      }

      // If main window direct redirect
      window.history.replaceState({}, document.title, window.location.pathname);
      setIsIngestionOpen(true);
      fetchGitHubRepositories(instId).catch(() => {});
    }

    // 2. Listen to storage events from popup window
    const handleStorage = (event) => {
      if (event.key === 'cbom_github_installation_id' && event.newValue) {
        setIsIngestionOpen(true);
        fetchGitHubRepositories(event.newValue).catch(() => {});
      }
    };
    window.addEventListener('storage', handleStorage);

    // 3. Listen to postMessage from popup window
    const handleMessage = (event) => {
      if (event.data && event.data.type === 'GITHUB_AUTH_SUCCESS' && event.data.installation_id) {
        const id = event.data.installation_id;
        localStorage.setItem('cbom_github_installation_id', id);
        setIsIngestionOpen(true);
        fetchGitHubRepositories(id).catch(() => {});
      }
    };
    window.addEventListener('message', handleMessage);

    // Detect backend restart
    import('./api/client').then(({ api }) => {
      api.getGitHubInstallUrl()
        .then((res) => {
          if (res && res.server_instance_id) {
            const storedServerID = localStorage.getItem('cbom_server_instance_id');
            if (storedServerID && storedServerID !== res.server_instance_id) {
              const savedInstID = localStorage.getItem('cbom_github_installation_id');
              localStorage.clear();
              if (savedInstID) {
                localStorage.setItem('cbom_github_installation_id', savedInstID);
              }
            }
            localStorage.setItem('cbom_server_instance_id', res.server_instance_id);
          }
        })
        .catch(() => {});
    });

    return () => {
      window.removeEventListener('storage', handleStorage);
      window.removeEventListener('message', handleMessage);
    };
  }, []);

  const headerActionsProps = {
    currentScan,
    scans,
    onSelectScan: selectScan,
    onOpenIngestion: () => setIsIngestionOpen(true),
    onOpenExport: () => setIsExportOpen(true),
    scanning,
  };

  return (
    <div>
      {/* Error Banner */}
      {error && (
        <div style={{
          marginBottom: '16px',
          padding: '12px 20px',
          backgroundColor: 'rgba(239, 68, 68, 0.15)',
          border: '1px solid rgba(239, 68, 68, 0.3)',
          borderRadius: '10px',
          color: '#fca5a5',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          fontSize: '14px',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <AlertCircle style={{ width: '18px', height: '18px', flexShrink: 0 }} />
            <span>{error}</span>
          </div>
          <button onClick={() => setError(null)} style={{ background: 'none', border: 'none', color: '#fca5a5', cursor: 'pointer', fontWeight: 'bold' }}>
            Dismiss
          </button>
        </div>
      )}

      {/* Notice Banner */}
      {currentScan?.summary?.notice && (
        <div style={{
          marginBottom: '16px',
          padding: '12px 20px',
          backgroundColor: 'rgba(0, 168, 255, 0.15)',
          border: '1px solid rgba(0, 168, 255, 0.3)',
          borderRadius: '10px',
          color: 'var(--color-primary)',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          fontSize: '14px',
        }}>
          <Info style={{ width: '18px', height: '18px', flexShrink: 0 }} />
          <span>{currentScan.summary.notice}</span>
        </div>
      )}

      {/* Page Content */}
      {loading && !currentScan ? (
        <div className="card" style={{ padding: '60px', textAlign: 'center', color: 'var(--color-text-muted)' }}>
          <div style={{ fontSize: '18px', fontWeight: '600', marginBottom: '8px', color: 'var(--color-primary)' }}>
            Initializing CBOM Intelligence Engine...
          </div>
          <div style={{ fontSize: '13px' }}>
            Loading cryptographic asset inventories, risk scores, and PQC migration strategies.
          </div>
        </div>
      ) : (
        <Outlet context={{ currentScan, scanning, headerActionsProps }} />
      )}

      {/* Modals */}
      <IngestionModal
        isOpen={isIngestionOpen}
        onClose={() => setIsIngestionOpen(false)}
        onStartLocal={startLocalScan}
        onFetchGitHubRepos={fetchGitHubRepositories}
        onStartGitHub={startGitHubScan}
        loadingRepos={loadingRepos}
        githubRepos={githubRepos}
        scanning={scanning}
      />

      <CycloneDXExportModal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
        currentScan={currentScan}
      />
    </div>
  );
};

/* ── App ── */
export function App() {
  return (
    <Router>
      <Routes>
        <Route element={<MainLayout />}>
          {/* Default redirect to CBOM Dashboard */}
          <Route index element={<Navigate to="/cbom" replace />} />

          {/* QuantaVault placeholder pages */}
          <Route path="dashboard" element={<PlaceholderPage title="Dashboard" />} />
          <Route path="pqc-keys" element={<PlaceholderPage title="PQC Keys" />} />
          <Route path="authentication-keys" element={<PlaceholderPage title="Authentication Keys" />} />
          <Route path="policies" element={<PlaceholderPage title="Policies" />} />
          <Route path="audit-logs" element={<PlaceholderPage title="Audit Logs" />} />
          <Route path="billing" element={<PlaceholderPage title="Billing" />} />
          <Route path="settings" element={<PlaceholderPage title="Settings" />} />

          {/* Certificate Manager placeholder pages */}
          <Route path="cert-manager" element={<PlaceholderPage title="Certificate Manager" />} />
          <Route path="cert-manager/certificates" element={<PlaceholderPage title="Certificates" />} />
          <Route path="cert-manager/keys" element={<PlaceholderPage title="Cert Keys" />} />
          <Route path="cert-manager/csr" element={<PlaceholderPage title="Generate CSR" />} />
          <Route path="cert-manager/internal-ca" element={<PlaceholderPage title="Internal CA" />} />
          <Route path="cert-manager/audit" element={<PlaceholderPage title="Cert Audit Logs" />} />

          {/* CBOM pages with shared scan state */}
          <Route path="cbom" element={<CbomLayout />}>
            <Route index element={<CbomDashboardWrapper />} />
            <Route path="inventory" element={<CbomInventoryWrapper />} />
            <Route path="risk" element={<CbomRiskWrapper />} />
            <Route path="migration" element={<CbomMigrationWrapper />} />
          </Route>
        </Route>
      </Routes>
    </Router>
  );
}

/* ── CBOM Page Wrappers with QuantaVault PageHeader & Page Layout ── */

function CbomDashboardWrapper() {
  const { currentScan, headerActionsProps } = useOutletContext();
  const navigate = useNavigate();
  return (
    <div className="page">
      <PageHeader
        title="CBOM Intelligence"
        subtitle="Cryptographic Inventory & Post-Quantum Cryptography (PQC) Risk Assessment"
        action={<CbomHeaderActions {...headerActionsProps} />}
      />
      <div className="page-content">
        <DashboardPage currentScan={currentScan} onNavigateToMigration={() => navigate('/cbom/migration')} />
      </div>
    </div>
  );
}

function CbomInventoryWrapper() {
  const { currentScan, headerActionsProps } = useOutletContext();
  return (
    <div className="page">
      <PageHeader
        title="Asset Inventory"
        subtitle="Complete inventory of discovered cryptographic algorithms, key lengths, certificates, and libraries"
        action={<CbomHeaderActions {...headerActionsProps} />}
      />
      <div className="page-content">
        <InventoryPage currentScan={currentScan} />
      </div>
    </div>
  );
}

function CbomRiskWrapper() {
  const { currentScan, headerActionsProps } = useOutletContext();
  return (
    <div className="page">
      <PageHeader
        title="Risk Analysis"
        subtitle="Quantum vulnerability findings scored against NIST SP 800-131A & CNSA 2.0 standards"
        action={<CbomHeaderActions {...headerActionsProps} />}
      />
      <div className="page-content">
        <RiskAnalysisPage currentScan={currentScan} />
      </div>
    </div>
  );
}

function CbomMigrationWrapper() {
  const { currentScan, headerActionsProps } = useOutletContext();
  return (
    <div className="page">
      <PageHeader
        title="PQC Migration Strategy"
        subtitle="NIST PQC standardized replacement algorithms & modernization roadmap"
        action={<CbomHeaderActions {...headerActionsProps} />}
      />
      <div className="page-content">
        <MigrationPlannerPage currentScan={currentScan} />
      </div>
    </div>
  );
}

export default App;
