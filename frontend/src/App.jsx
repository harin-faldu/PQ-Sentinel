import React, { useState } from 'react';
import { Navbar } from './components/Navbar';
import { DashboardPage } from './pages/DashboardPage';
import { InventoryPage } from './pages/InventoryPage';
import { RiskAnalysisPage } from './pages/RiskAnalysisPage';
import { MigrationPlannerPage } from './pages/MigrationPlannerPage';
import { useCbomScan } from './hooks/useCbomScan';

export function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const { currentScan, loading, scanning, startScan } = useCbomScan();

  const handleScanTrigger = async () => {
    const target = prompt('Enter target directory path to scan:', '.');
    if (target !== null) {
      await startScan(target);
    }
  };

  return (
    <div style={{ minHeight: '100vh', paddingBottom: '40px' }}>
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        currentScan={currentScan}
        onScanTrigger={handleScanTrigger}
        scanning={scanning}
      />

      <main style={{ maxWidth: '1400px', margin: '24px auto 0 auto', padding: '0 24px' }}>
        {loading && !currentScan ? (
          <div className="glass-card" style={{ padding: '48px', textAlign: 'center', color: '#94a3b8' }}>
            Initializing QuantumVault CBOM Intelligence Engine...
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
    </div>
  );
}

export default App;
