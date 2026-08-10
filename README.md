# 🛡️ QuantumVault CBOM Scanner 2.0

> **Enterprise-Grade Post-Quantum Cryptographic Bill of Materials (CBOM) & Risk Management Platform**  
> *Built with Go 1.22+, React + Vite Web Dashboard, Tree-Sitter AST, Embedded Gitleaks v8, Active TLS Probing, GitHub App Ephemeral Git Clone Ingestion, and CycloneDX 1.6 Exporter*

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat-square&logo=react)](https://reactjs.org)
[![Vite](https://img.shields.io/badge/Vite-5.1-646CFF?style=flat-square&logo=vite)](https://vitejs.dev)
[![NIST PQC](https://img.shields.io/badge/NIST%20PQC-FIPS%20203%20%7C%20204%20%7C%20205-7B2CBF?style=flat-square)](https://csrc.nist.gov/projects/post-quantum-cryptography)
[![NSA CNSA 2.0](https://img.shields.io/badge/NSA%20CNSA%202.0-Compliant-005F73?style=flat-square)](https://www.nsa.gov)
[![CycloneDX Spec](https://img.shields.io/badge/CycloneDX-1.6%20JSON%20%26%20XML-38B000?style=flat-square)](https://cyclonedx.org)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen?style=flat-square)](#-testing--verification)

---

## 📌 Table of Contents

- [Overview \& Vision](#-overview--vision)
- [Architecture \& Engine Workflow](#-architecture--engine-workflow)
- [Key Features \& Capabilities](#-key-features--capabilities)
- [Dual Codebase Ingestion Modes](#-dual-codebase-ingestion-modes)
- [React Web Dashboard \& UI Features](#-react-web-dashboard--ui-features)
- [Directory \& File Structure](#-directory--file-structure)
- [NSA CNSA 2.0 Compliance Timelines](#-nsa-cnsa-20-compliance-timelines)
- [Getting Started \& Installation](#-getting-started--installation)
- [How to Run the Server \& Dashboard](#-how-to-run-the-server--dashboard)
- [Testing \& Verification](#-testing--verification)
- [License \& Security](#-license--security)

---

## 🔍 Overview & Vision

**QuantumVault CBOM Scanner 2.0** is an enterprise-grade Post-Quantum Cryptographic Bill of Materials (CBOM) platform. It provides automated cryptographic discovery across source code ASTs, secrets, containers, cloud KMS keys, X.509 certificates, Java/PKCS12 keystores, and live network endpoints.

It calculates a **NIST PQC Readiness Score (0-100)**, tracks **NSA CNSA 2.0 regulatory compliance deadlines**, aggregates call-site symbol dependencies, and exports standard **CycloneDX 1.6 CBOM** specifications in both **JSON** and **XML**.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           QUANTUMVAULT CBOM PLATFORM                            │
│                                                                                 │
│  ┌─────────────────────────────────┐     ┌───────────────────────────────────┐  │
│  │ Mode 1: Local Codebase Directory│     │ Mode 2: GitHub App Ephemeral Clone│  │
│  └────────────────┬────────────────┘     └─────────────────┬─────────────────┘  │
│                   │                                        │                    │
│                   └──────────────────┬─────────────────────┘                    │
│                                      ▼                                          │
│ ┌─────────────────────────────────────────────────────────────────────────────┐ │
│ │ 1. AST Engine │ 2. Git & Secrets │ 3. AWS KMS │ 4. Syft Container │ 5. TLS  │ │
│ └────────────────────────────────────┬────────────────────────────────────────┘ │
│                                      ▼                                          │
│ ┌─────────────────────────────────────────────────────────────────────────────┐ │
│ │ 6. SymbolGraph Tracing & Scope Elevation (Prod 1.5x, Staging 0.7x, Test 0.3x)│ │
│ └────────────────────────────────────┬────────────────────────────────────────┘ │
│                                      ▼                                          │
│ ┌─────────────────────────────────────────────────────────────────────────────┐ │
│ │ 7. PQC Risk Score Engine & CNSA 2.0 Timeline Tagging                        │ │
│ └────────────────────────────────────┬────────────────────────────────────────┘ │
│                                      ▼                                          │
│ ┌────────────────────────────────────┴────────────────────────────────────────┐ │
│ │ 8. Web Dashboard & CycloneDX 1.6 Standard Exporter (JSON + XML)             │ │
│ └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔁 System Architecture & End-to-End Workflow

```
[1. User Action] ---> Click "Scan Codebase" on Web Dashboard
                               │
                               ▼
[2. Ingestion Modal] -> Choose Mode 1 (Local Folder) OR Mode 2 (GitHub App Repo)
                               │
                               ▼
[3. GitHub Auth] -----> (If Mode 2) Exchange RS256 JWT for 1-hr Installation Token
                               │
                               ▼
[4. HEAD Check] ------> Call `git ls-remote`. If commit SHA matches LastScannedSHA,
                        skip clone & return notice ("Repository commit unchanged").
                               │
                               ▼
[5. Ephemeral Clone] -> Perform full `git clone` into temp dir (`defer os.RemoveAll`).
                               │
                               ▼
[6. Discovery] -------> Run AST Tree-Sitter, Gitleaks history, AWS KMS, Docker, TLS.
                               │
                               ▼
[7. SymbolGraph] -----> Aggregate call-site consumers (`details`), elevate scope weight.
                               │
                               ▼
[8. Risk Engine] -----> Calculate PQC Score (0-100), CNSA 2.0 timeline & migration plan.
                               │
                               ▼
[9. Persistence] -----> Save summary & assets into MySQL / Memory repository.
                               │
                               ▼
[10. Presentation] ---> Render on React Dashboard & export CycloneDX 1.6 (JSON/XML).
```

---

## 🚀 Key Features & Capabilities

### 1. 🌳 Tree-Sitter AST Syntax Parsing
- Deep Concrete Syntax Tree (CST/AST) parsing for **12+ Language Families**:
  - `JavaScript`, `TypeScript`, `Python`, `Go`, `Java`, `Kotlin`, `C/C++`, `C#`, `Rust`, `Ruby`, `PHP`, `Terraform`.
- Directly targets `call_expression` nodes, walking AST argument trees to extract multi-line parameters (e.g. `modulusLength: 1024`) across line boundaries with **zero false positives** on comments or string logs.

### 2. 🌐 SymbolGraph Call-Site Tracing & Scope Weight Elevation
- **Single-Asset Call-Site Aggregation**: SymbolGraph creates exactly 1 `CryptoAsset` per definition site, aggregating caller files into `CryptoAsset.Details` (e.g. `Consumed by controllers/auth.js:L45`).
- **Scope Elevation**: If a production file consumes a symbol defined in a utility file, the asset's `BusinessScope` and `ScopeWeight` automatically elevate to `PRODUCTION` (1.5x), ensuring high-risk callers boost visibility.

### 3. 🔐 Keystore Decryption & Weak Credential Detection
- Decrypts **PKCS#12** (`.p12`, `.pfx`) and **Java Keystores** (`.jks`) using `go-pkcs12` and `keystore-go/v4`.
- **Weak Password Gate**: Generates a **CRITICAL** risk finding if a keystore is protected by weak default passwords (`changeit`, `password`, `123456`).
- **Exact Bit-Size Measurement**: Measures actual public key bit lengths (`rsa.N.BitLen()`, `ecdsa.Curve.Params().BitSize`).

### 4. 🗝️ Secrets & Git Commit History Scanning
- Embedded **Gitleaks v8 engine** scanning in-memory files and raw content.
- **Git Commit History Traversal**: Runs `git log --all -p` to inspect deleted files, past commits, and historical exposed credentials in `.git`.
- **Token Redaction**: Redacts sensitive tokens in logs (`strings.ReplaceAll(output, token, "***REDACTED***")`).

### 5. ☁️ Live AWS KMS & IaC Scanner
- Live AWS SDK v2 discovery (`ListKeys`, `DescribeKey`, `GetKeyRotationStatus`) across 30+ regions (including `ap-south-1` Mumbai).
- Matches live AWS keys against static Terraform/CloudFormation IaC declared keys using **KMS Alias Names (`alias/...`)**.

### 6. 🐳 Container & Syft Image Scanner
- Syft CLI binary orchestrator parsing SBOM packages, assigning **CRITICAL** risk to EOL OpenSSL versions (`libssl1.1`).
- Static Dockerfile fallback parser with regex security gate preventing command injection.

---

## ⚡ Dual Codebase Ingestion Modes

| Mode | Target Type | Backend Method | Description |
|---|---|---|---|
| **Mode 1: Local Codebase** | `TargetTypeRepo` (`REPO`) | `s.scanner.ScanDirectory(dir, scanID)` | Direct local filesystem scan for local codebases or staging directories. |
| **Mode 2: GitHub Repository** | `TargetTypeGitHubRepo` (`GITHUB_REPO`) | `github.CloneRepository` + `s.scanner.ScanDirectory(tempDir, scanID)` | Remote GitHub App ephemeral `git clone` with full history secret scanning and auto-cleanup. |

---

## 🎨 React Web Dashboard & UI Features

- **Ingestion Modal ([IngestionModal.jsx](file:///c:/CBOM/frontend/src/components/IngestionModal.jsx))**: Codebase selection modal for Local Directories & Remote GitHub App Repositories.
- **CycloneDX 1.6 Exporter Modal ([CycloneDXExportModal.jsx](file:///c:/CBOM/frontend/src/components/CycloneDXExportModal.jsx))**: Live JSON & XML viewers with 1-click **Copy to Clipboard** and **File Download**.
- **Interactive Asset Inventory ([AssetTable.jsx](file:///c:/CBOM/frontend/src/components/AssetTable.jsx))**: Multi-filter search by asset type, severity, and SymbolGraph consumer call sites.
- **Automated Migration Strategy ([MigrationPlannerPage.jsx](file:///c:/CBOM/frontend/src/pages/MigrationPlannerPage.jsx))**: Step-by-step PQC remediation playbooks mapping vulnerable assets to NIST ML-DSA & ML-KEM algorithms.
- **Fully Responsive Layout**: Built with custom CSS utilities and media queries (`@media (max-width: 1024px)`, `@media (max-width: 768px)`), adapting fluidly across mobile phones, tablets, and 4K desktops.

---

## 📁 Directory & File Structure

```
c:\CBOM
├── backend/
│   ├── cmd/
│   │   ├── scanner/main.go            # CLI Scanner Binary
│   │   └── server/main.go             # REST API Server Entry Point
│   ├── internal/
│   │   ├── controllers/
│   │   │   ├── cbom_controller.go     # Scans, Targets & Migration Handlers
│   │   │   └── github_controller.go   # GitHub App Repo Picker Handler
│   │   ├── github/
│   │   │   ├── auth.go                # RS256 JWT & Token Exchange
│   │   │   ├── client.go              # Installation Repos API
│   │   │   └── clone.go               # Ephemeral Git Clone & Redaction
│   │   ├── models/models.go           # Core Data Models & CycloneDX Structs
│   │   ├── repository/
│   │   │   └── cbom_repository.go     # MySQL & Memory Storage Layer
│   │   ├── risk/risk_engine.go        # Risk Score Engine & Remediation Playbooks
│   │   ├── scanner/
│   │   │   ├── scanner.go             # Core Scanner Orchestrator
│   │   │   ├── symbol_graph.go        # Symbol Call-Site Tracing & Scope Elevation
│   │   │   ├── aws_kms_scanner.go     # Live AWS KMS & IaC Scanner
│   │   │   ├── container_scanner.go   # Container & Syft Scanner
│   │   │   ├── secret_scanner.go      # Gitleaks History Scanner
│   │   │   ├── cert_scanner.go        # X.509 & Keystore Scanner
│   │   │   └── tls_prober.go          # Active Multi-Port TLS Prober
│   │   └── utils/secretbox.go         # AES-256-GCM Secret Encryption Box
│   └── rules.yaml                     # Crypto Detection Rules Definition
├── frontend/
│   ├── src/
│   │   ├── api/client.js              # REST API Client
│   │   ├── components/
│   │   │   ├── Navbar.jsx             # Top Navigation Header
│   │   │   ├── IngestionModal.jsx     # Ingestion Modal (Local + GitHub App)
│   │   │   ├── CycloneDXExportModal.jsx # CycloneDX JSON/XML Exporter Modal
│   │   │   ├── AssetTable.jsx         # Inventory Asset Table
│   │   │   ├── OverviewCards.jsx      # Summary Cards
│   │   │   └── RiskGauge.jsx          # PQC Risk Gauge Component
│   │   ├── hooks/useCbomScan.js       # Scan & Target State Hook
│   │   └── pages/                     # Dashboard, Inventory, Risk & Migration Pages
│   ├── vite.config.js                 # Vite Dev Server & API Proxy
│   └── package.json
└── sample_target/                     # Comprehensive Testbed Folder
```

---

## 🏛️ NSA CNSA 2.0 Compliance Timelines

| Category | Classical Algorithm | Target PQC Algorithm | NSA CNSA 2.0 Mandatory Deadline |
|---|---|---|---|
| **Software & Firmware Signing** | RSA-2048 / ECDSA | **ML-DSA-65 / ML-DSA-87** | **2025 (Begin) — 2030 (Mandatory)** |
| **Symmetric Key Encryption** | AES-128 | **AES-256** | **2026 (Begin) — 2030 (Mandatory)** |
| **Key Establishment & KEM** | ECDH / RSA Key Exchange | **ML-KEM-768 / ML-KEM-1024** | **2026 (Begin) — 2033 (Mandatory)** |
| **Web Browsers & TLS** | RSA / ECDSA Certificates | **ML-DSA-65 / Hybrid TLS** | **2025 (Begin) — 2033 (Mandatory)** |

---

## 💻 How to Run the Server & Dashboard

### Step 1: Start Backend REST API Server (Terminal 1)

Optionally configure environment variables for GitHub App ingestion:

```powershell
cd c:\CBOM\backend

# Optional GitHub App Credentials
$env:GITHUB_APP_ID="4534008"
$env:GITHUB_PRIVATE_KEY_PATH="C:\path\to\your-app-private-key.pem"

# Start backend API server
go run ./cmd/server/main.go
```
*The Go server runs on `http://localhost:8081` (or `8080`).*

### Step 2: Start Frontend Web Dashboard (Terminal 2)

```powershell
cd c:\CBOM\frontend
npm run dev
```
*The Vite Dev Server runs on `http://localhost:3000` with automatic API proxying.*

### Step 3: Test via Web Dashboard (`http://localhost:3000`)

1. **Local Codebase Scan**: Click **"Scan Codebase"** → Select **"Local Directory Codebase"** → Enter `sample_target` → Click **"Start Local Scan"**.
2. **GitHub App Repo Scan**: Click **"Scan Codebase"** → Select **"GitHub App Repository"** → Enter Installation ID → Click **"Fetch Repositories"** → Select repo → Click **"Scan Selected Repository"**.
3. **CycloneDX 1.6 Export**: Click **"CycloneDX 1.6"** in top navbar → Switch between JSON and XML views → Click **"Copy Content"** or **"Download File"**.

---

## 🧪 Testing & Verification

Run the full automated Go test suite covering all 16 modules:

```powershell
cd c:\CBOM\backend
go test -v ./...
```

### Verified Test Suite Breakdown:
- `TestGenerateAppJWT` & `TestGetInstallationTokenValidation`: RS256 JWT signing, 9-min expiry window, and Installation Access Token exchange error handling.
- `TestTokenScrubbingOnCloneFailure`: Verifies tokens are redacted (`***REDACTED***`) from git CLI error output.
- `TestLocalGitCloneAndHistorySecretDetection`: Ephemeral git clone history secret scanning against committed & deleted secrets.
- `TestSecretBoxEncryptionRoundtrip`: AES-256-GCM secret box encryption/decryption for GitHub App RSA `.pem` private keys.
- `TestSymbolGraphCallSiteAggregationAndScopeElevation`: Symbol call-site aggregation into `CryptoAsset.Details` and `ScopeWeight` elevation to `PRODUCTION` (1.5x).
- `TestAWSKMSScannerIaCFallback`: Live AWS KMS key discovery and Terraform / CloudFormation IaC resource parsing.
- `TestContainerScannerSyftJSONParsing` & `TestContainerScannerDockerfileFallback`: Syft SBOM package parsing & static Dockerfile fallback parsing.
- `TestCycloneDXSpecCompliance`: CycloneDX 1.6 specification compliance (`pkg:generic/rsa@2048` PURLs, dynamic NIST levels, XML export).
- `TestTargetEntityManagement`: Persistent Target CRUD, auto-generated IDs, and scan history.

---

## 📄 License & Security

- **License**: Proprietary / Enterprise Security Tool.
- **Security Notice**: Never commit RSA `.pem` private keys or AWS secret access keys into public repositories. Use environment variables (`GITHUB_PRIVATE_KEY_PATH`, `MASTER_KEY`) for secure key management.
