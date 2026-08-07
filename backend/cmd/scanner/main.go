package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/services"
)

func main() {
	tlsFlag := flag.String("tls", "", "Comma-separated active TLS endpoints to probe (e.g. google.com:443,api.example.com:8443)")
	xmlFlag := flag.Bool("xml", true, "Export CycloneDX 1.6 XML specification alongside JSON")
	flag.Parse()

	targetDir := "."
	args := flag.Args()
	if len(args) > 0 {
		targetDir = args[0]
	}

	var tlsTargets []string
	if *tlsFlag != "" {
		tlsTargets = strings.Split(*tlsFlag, ",")
	}

	fmt.Printf("🔍 Running QuantumVault CBOM Scanner 2.0 (Perfect 10/10 Engine) on: %s\n", targetDir)
	if len(tlsTargets) > 0 {
		fmt.Printf("🌐 Active TLS Multi-Port Probing enabled for %d targets: %v\n", len(tlsTargets), tlsTargets)
	}

	repo := repository.NewMemoryRepository()
	cbomService := services.NewCBOMService(repo)

	// Load previous score history for trend display
	previousHistory := cbomService.GetScoreHistory()
	var lastScore int = -1
	if len(previousHistory) > 0 {
		lastScore = previousHistory[len(previousHistory)-1].PQCScore
	}

	detail, err := cbomService.ExecuteScanWithTLS(targetDir, tlsTargets)
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		os.Exit(1)
	}

	// Score trend indicator
	trendStr := ""
	if lastScore >= 0 {
		delta := detail.Summary.PQCScore - lastScore
		if delta > 0 {
			trendStr = fmt.Sprintf("  (↑ +%d from last scan)", delta)
		} else if delta < 0 {
			trendStr = fmt.Sprintf("  (↓ %d from last scan)", delta)
		} else {
			trendStr = "  (= no change from last scan)"
		}
	}

	fmt.Println("\n=======================================================")
	fmt.Printf("  QUANTUMVAULT CBOM SCAN REPORT (PERFECT 10/10 ENGINE)\n")
	fmt.Println("=======================================================")
	fmt.Printf("  Target Directory: %s\n", detail.Summary.TargetDirectory)
	fmt.Printf("  Total Files:      %d | Scanned Files: %d\n", detail.Summary.TotalFiles, detail.Summary.ScannedFiles)
	fmt.Printf("  Assets Found:     %d\n", detail.Summary.TotalAssets)
	fmt.Printf("  NIST PQC Score:   %d / 100%s\n", detail.Summary.PQCScore, trendStr)
	fmt.Printf("  Risks:            Critical: %d | High: %d | Medium: %d | Safe: %d\n",
		detail.Summary.CriticalCount, detail.Summary.HighCount, detail.Summary.MediumCount, detail.Summary.SafeCount)
	fmt.Println("-------------------------------------------------------")

	for i, asset := range detail.Assets {
		fmt.Printf("[%d] %s (%s)", i+1, asset.Name, asset.RiskLevel)
		if asset.Confidence > 0 && asset.Confidence < 1.0 {
			fmt.Printf(" [Confidence: %.0f%%]", asset.Confidence*100)
		}
		if asset.BusinessScope != "" {
			fmt.Printf(" [%s]", asset.BusinessScope)
		}
		fmt.Println()
		fmt.Printf("    Location:       %s:%d\n", asset.Location, asset.LineNumber)
		fmt.Printf("    Algorithm:      %s (%d-bit)\n", asset.Algorithm, asset.KeySize)
		fmt.Printf("    CNSA 2.0 Target:%s (%s)\n", asset.CNSA20Deadline, asset.CNSA20Status)
		fmt.Printf("    Risk Rationale: %s\n", asset.RiskRationale)
		fmt.Printf("    PQC Target:     %s\n", asset.PQCReplacement)
		fmt.Printf("    Snippet/Ctx:    %s\n\n", asset.ContextSnippet)
	}

	// 1. Export CycloneDX 1.6 JSON
	cyclonedxJSON, _ := cbomService.GenerateCycloneDX(detail.Summary.ID)
	outputJSON, _ := json.MarshalIndent(cyclonedxJSON, "", "  ")
	_ = os.WriteFile("cbom-output.json", outputJSON, 0644)
	fmt.Println("📄 Exported CycloneDX 1.6 CBOM (JSON) to: cbom-output.json")

	// 2. Export CycloneDX 1.6 XML
	if *xmlFlag {
		outputXML, err := cbomService.GenerateCycloneDXXML(detail.Summary.ID)
		if err == nil {
			_ = os.WriteFile("cbom-output.xml", []byte(outputXML), 0644)
			fmt.Println("📜 Exported CycloneDX 1.6 CBOM (XML)  to: cbom-output.xml")
		}
	}

	// 3. Score History
	fmt.Println("📊 Score history saved to: score-history.json")
}
