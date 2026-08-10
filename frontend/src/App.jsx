import React, { useState } from 'react';
import { Navbar } from './components/Navbar';
import { DashboardPage } from './pages/DashboardPage';
import { InventoryPage } from './pages/InventoryPage';
import { RiskAnalysisPage } from './pages/RiskAnalysisPage';
import { MigrationPlannerPage } from './pages/MigrationPlannerPage';
import { IngestionModal } from './components/IngestionModal';
import { CycloneDXExportModal } from './components/CycloneDXExportModal';
import { useCbomScan } from './hooks/useCbomScan';
import { AlertCircle, CheckCircle2, Info } from 'lucide-react';

export function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
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

  return (
    <div style={{ minHeight: '100vh', paddingBottom: '40px', backgroundColor: '#0f172a', color: '#f8fafc' }}>
      {/* Navigation Header Bar */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        currentScan={currentScan}
        scans={scans}
        targets={targets}
        onSelectScan={selectScan}
        onOpenIngestion={() => setIsIngestionOpen(true)}
        onOpenExport={() => setIsExportOpen(true)}
        scanning={scanning}
      />

      {/* Global Error Banner */}
      {error && (
        <div style={{
          maxWidth: '1400px',
          margin: '16px auto 0 auto',
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

      {/* Scan Notice Banner (e.g. Unchanged Git Commit Skip) */}
      {currentScan?.summary?.notice && (
        <div style={{
          maxWidth: '1400px',
          margin: '16px auto 0 auto',
          padding: '12px 20px',
          backgroundColor: 'rgba(56, 189, 248, 0.15)',
          border: '1px solid rgba(56, 189, 248, 0.3)',
          borderRadius: '10px',
          color: '#7dd3fc',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          fontSize: '14px',
        }}>
          <Info style={{ width: '18px', height: '18px', flexShrink: 0 }} />
          <span>{currentScan.summary.notice}</span>
        </div>
      )}

      {/* Main Page Content */}
      <main style={{ maxWidth: '1400px', width: '100%', margin: '20px auto 0 auto', padding: '0 16px' }}>
        {loading && !currentScan ? (
          <div className="glass-card" style={{ padding: '60px', textAlign: 'center', color: '#94a3b8' }}>
            <div style={{ fontSize: '18px', fontWeight: '600', marginBottom: '8px', color: '#38bdf8' }}>
              Initializing QuantumVault CBOM Intelligence Engine...
            </div>
            <div style={{ fontSize: '13px' }}>
              Loading cryptographic asset inventories, risk scores, and PQC migration strategies.
            </div>
          </div>
        ) : (
          <>
            {activeTab === 'dashboard' && (
              <DashboardPage
                currentScan={currentScan}
                onNavigateToMigration={() => setActiveTab('migration')}
              />
            )}
            {activeTab === 'inventory' && <InventoryPage currentScan={currentScan} />}
            {activeTab === 'risk' && <RiskAnalysisPage currentScan={currentScan} />}
            {activeTab === 'migration' && <MigrationPlannerPage currentScan={currentScan} />}
          </>
        )}
      </main>

      {/* Ingestion Modal (Local Filesystem Directory & GitHub App Repository Picker) */}
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

      {/* CycloneDX 1.6 Export Modal */}
      <CycloneDXExportModal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
        currentScan={currentScan}
      />
    </div>
  );
}

export default App;
