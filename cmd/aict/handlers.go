package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/branch"
	"github.com/y-hirakaw/ai-code-tracker/internal/period"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// ReportOptions holds options for the report command
type ReportOptions struct {
	Since        string
	From         string
	To           string
	Last         string
	Format       string
	Branch       string
	BranchRegex  string
	AllBranches  bool
}

// handleReportWithOptions handles the report command with period options
func handleReportWithOptions() {
	fs := flag.NewFlagSet("report", flag.ExitOnError)

	opts := &ReportOptions{}
	fs.StringVar(&opts.Since, "since", "", "Show report since this date/time")
	fs.StringVar(&opts.From, "from", "", "Start date for report range")
	fs.StringVar(&opts.To, "to", "", "End date for report range")
	fs.StringVar(&opts.Last, "last", "", "Show report for last N days/weeks/months (e.g., '7d', '2w', '1m')")
	fs.StringVar(&opts.Format, "format", "table", "Output format: table, graph, json")
	fs.StringVar(&opts.Branch, "branch", "", "Filter by specific branch name")
	fs.StringVar(&opts.BranchRegex, "branch-regex", "", "Filter by branch regex pattern")
	fs.BoolVar(&opts.AllBranches, "all-branches", false, "Show all branches summary")

	fs.Parse(os.Args[2:])

	baseDir := defaultBaseDir

	// Check if initialized
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("Error: AI Code Tracker not initialized. Run 'aict init' first.\n")
		os.Exit(1)
	}

	// Load configuration
	metricsStorage := storage.NewMetricsStorage(baseDir)
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Read JSONL records
	recorder := tracker.NewCheckpointRecorder(baseDir)
	records, err := recorder.ReadAllRecords()
	if err != nil {
		fmt.Printf("Error reading records: %v\n", err)
		os.Exit(1)
	}

	// Validate branch options
	if err := validateBranchOptions(opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Determine report type based on options
	hasBranch := hasBranchOptions(opts)
	hasPeriod := hasPeriodOptions(opts)
	
	switch {
	case hasBranch && hasPeriod:
		// Combined filtering - new functionality
		handleCombinedReport(records, config, opts)
	case hasBranch && !hasPeriod:
		// Branch only - existing functionality
		handleBranchReport(records, config, opts)
	case !hasBranch && hasPeriod:
		// Period only - existing functionality
		handlePeriodReport(records, config, opts)
	default:
		// Default report - existing functionality
		handleDefaultReport(records, config, opts)
	}
}

// handlePeriodReport handles period-specific reports
func handlePeriodReport(records []tracker.CheckpointRecord, config *tracker.Config, opts *ReportOptions) {
	var timeRange *period.TimeRange
	var err error

	// Parse time range based on options
	if opts.Last != "" {
		timeRange, err = period.ParseLastDuration(opts.Last)
		if err != nil {
			fmt.Printf("Error parsing last duration: %v\n", err)
			os.Exit(1)
		}
	} else if opts.Since != "" {
		timeRange, err = period.ParseTimeRange(opts.Since)
		if err != nil {
			fmt.Printf("Error parsing since time: %v\n", err)
			os.Exit(1)
		}
	} else if opts.From != "" && opts.To != "" {
		timeRange, err = period.ParseFromTo(opts.From, opts.To)
		if err != nil {
			fmt.Printf("Error parsing from/to range: %v\n", err)
			os.Exit(1)
		}
	} else if opts.From != "" {
		// From only - use current time as end
		fromTime, err := period.ParseTimeRange(opts.From)
		if err != nil {
			fmt.Printf("Error parsing from time: %v\n", err)
			os.Exit(1)
		}
		timeRange = &period.TimeRange{
			From: fromTime.From,
			To:   time.Now(),
		}
	} else {
		fmt.Printf("Error: Please specify a time range using --since, --from/--to, or --last\n")
		os.Exit(1)
	}

	// Analyze period
	analyzer := period.NewAnalyzer(config)
	report, err := analyzer.AnalyzePeriod(records, timeRange)
	if err != nil {
		fmt.Printf("Error analyzing period: %v\n", err)
		os.Exit(1)
	}

	// Format output
	format := period.ReportFormat(strings.ToLower(opts.Format))
	formatter := period.NewFormatter(config.TargetAIPercentage)

	output, err := formatter.Format(report, format)
	if err != nil {
		fmt.Printf("Error formatting report: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

// handleBranchReport handles branch-specific reports
func handleBranchReport(records []tracker.CheckpointRecord, config *tracker.Config, opts *ReportOptions) {
	analyzer := branch.NewBranchAnalyzer(records)

	// Handle different branch report types
	if opts.AllBranches {
		handleAllBranchesReport(analyzer, config.TargetAIPercentage)
	} else if opts.Branch != "" {
		handleSingleBranchReport(analyzer, opts.Branch, config.TargetAIPercentage)
	} else if opts.BranchRegex != "" {
		handleRegexBranchReport(analyzer, opts.BranchRegex, config.TargetAIPercentage)
	}
}

// handleAllBranchesReport shows summary of all branches
func handleAllBranchesReport(analyzer *branch.BranchAnalyzer, targetPercentage float64) {
	groupReport, err := analyzer.AnalyzeAllBranches()
	if err != nil {
		fmt.Printf("Error analyzing all branches: %v\n", err)
		os.Exit(1)
	}

	if len(groupReport.MatchingBranches) == 0 {
		fmt.Println("No branches found in tracking records.")
		return
	}

	fmt.Printf("All Branches Report\n")
	fmt.Printf("===================\n\n")
	
	// Show overall stats
	stats := analyzer.GetRecordStats()
	fmt.Printf("Overall Statistics:\n")
	fmt.Printf("  Total Records: %d\n", stats.TotalRecords)
	fmt.Printf("  Unique Branches: %d\n", stats.UniqueBranches)
	fmt.Printf("  Records with Branch Info: %d\n", stats.RecordsWithBranch)
	fmt.Printf("  Records without Branch Info: %d (shown as 'main (inferred)')\n\n", stats.RecordsWithoutBranch)

	fmt.Printf("Group Summary:\n")
	fmt.Printf("  Total Records: %d\n", groupReport.TotalRecords)
	fmt.Printf("  Total Added Lines: %d\n", groupReport.TotalAdded)
	fmt.Printf("  Group AI Ratio: %.1f%% (target: %.1f%%)\n", groupReport.GroupAIRatio, targetPercentage)
	
	// Progress indicator
	if groupReport.GroupAIRatio >= targetPercentage {
		fmt.Printf("  Progress: ✅ Target achieved (%.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100)
	} else {
		fmt.Printf("  Progress: 📊 %.1f%% of target\n\n", (groupReport.GroupAIRatio/targetPercentage)*100)
	}

	fmt.Printf("Per-Branch Breakdown:\n")
	for _, branchName := range groupReport.MatchingBranches {
		branchReport := groupReport.BranchReports[branchName]
		displayBranchName := branchName
		if branchName == "main" && len(groupReport.BranchReports) > 0 {
			// Check if this might be inferred
			if branchReport.RecordCount > 0 {
				// We can't easily check HasBranchInfo here, so just show the name as-is
				displayBranchName = branchName
			}
		}
		
		fmt.Printf("  %s: AI %.1f%% (%d/%d lines) [%d records]\n", 
			displayBranchName, 
			branchReport.AIRatio, 
			int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100), 
			branchReport.TotalAdded,
			branchReport.RecordCount)
	}
}

// handleSingleBranchReport shows detailed report for a specific branch
func handleSingleBranchReport(analyzer *branch.BranchAnalyzer, branchName string, targetPercentage float64) {
	branchReport, err := analyzer.AnalyzeByBranch(branchName)
	if err != nil {
		fmt.Printf("Error analyzing branch '%s': %v\n", branchName, err)
		os.Exit(1)
	}

	if branchReport.RecordCount == 0 {
		fmt.Printf("No records found for branch '%s'.\n", branchName)
		fmt.Println("\nAvailable branches:")
		branches := analyzer.GetUniqueBranches()
		for _, branch := range branches {
			fmt.Printf("  %s\n", branch)
		}
		return
	}

	fmt.Printf("Branch Report: %s\n", branchName)
	fmt.Printf("================================\n")
	fmt.Printf("Records: %d (%s to %s)\n", 
		branchReport.RecordCount,
		branchReport.FirstRecord.Format("2006-01-02"),
		branchReport.LastRecord.Format("2006-01-02"))
	fmt.Printf("Added Lines: %d (AI: %d, Human: %d)\n", 
		branchReport.TotalAdded,
		int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100),
		branchReport.TotalAdded-int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100))
	fmt.Printf("AI Ratio: %.1f%%\n", branchReport.AIRatio)
	
	if branchReport.AIRatio >= targetPercentage {
		fmt.Printf("Progress: ✅ %.1f%% (target: %.1f%%)\n", (branchReport.AIRatio/targetPercentage)*100, targetPercentage)
	} else {
		fmt.Printf("Progress: 📊 %.1f%% (target: %.1f%%)\n", (branchReport.AIRatio/targetPercentage)*100, targetPercentage)
	}

	if len(branchReport.Authors) > 0 {
		fmt.Printf("Authors: %s\n", strings.Join(branchReport.Authors, ", "))
	}
}

// handleRegexBranchReport shows report for branches matching a regex pattern
func handleRegexBranchReport(analyzer *branch.BranchAnalyzer, pattern string, targetPercentage float64) {
	groupReport, err := analyzer.AnalyzeByPattern(pattern, true)
	if err != nil {
		fmt.Printf("Error analyzing branches with pattern '%s': %v\n", pattern, err)
		os.Exit(1)
	}

	if len(groupReport.MatchingBranches) == 0 {
		fmt.Printf("No branches found matching pattern '%s'.\n", pattern)
		fmt.Println("\nAvailable branches:")
		branches := analyzer.GetUniqueBranches()
		for _, branch := range branches {
			fmt.Printf("  %s\n", branch)
		}
		return
	}

	fmt.Printf("Branch Pattern Report: \"%s\"\n", pattern)
	fmt.Printf("==================================\n")
	fmt.Printf("Matching Branches: %s\n", strings.Join(groupReport.MatchingBranches, ", "))
	fmt.Printf("Total Records: %d\n", groupReport.TotalRecords)
	fmt.Printf("Added Lines: %d (AI: %d, Human: %d)\n", 
		groupReport.TotalAdded,
		int(float64(groupReport.TotalAdded)*groupReport.GroupAIRatio/100),
		groupReport.TotalAdded-int(float64(groupReport.TotalAdded)*groupReport.GroupAIRatio/100))
	fmt.Printf("Group AI Ratio: %.1f%%\n", groupReport.GroupAIRatio)

	if groupReport.GroupAIRatio >= targetPercentage {
		fmt.Printf("Progress: ✅ %.1f%% (target: %.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100, targetPercentage)
	} else {
		fmt.Printf("Progress: 📊 %.1f%% (target: %.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100, targetPercentage)
	}

	fmt.Printf("Per-Branch Breakdown:\n")
	for _, branchName := range groupReport.MatchingBranches {
		branchReport := groupReport.BranchReports[branchName]
		fmt.Printf("  %s: AI %.1f%% (%d/%d lines) [%d records]\n", 
			branchName, 
			branchReport.AIRatio, 
			int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100), 
			branchReport.TotalAdded,
			branchReport.RecordCount)
	}
}

// Helper functions for Phase 4: Combined period and branch filtering

// hasBranchOptions checks if any branch filtering options are specified
func hasBranchOptions(opts *ReportOptions) bool {
	return opts.Branch != "" || opts.BranchRegex != "" || opts.AllBranches
}

// hasPeriodOptions checks if any period filtering options are specified
func hasPeriodOptions(opts *ReportOptions) bool {
	return opts.Since != "" || opts.From != "" || opts.Last != ""
}

// validateBranchOptions ensures only one branch option is specified
func validateBranchOptions(opts *ReportOptions) error {
	branchOptionsCount := 0
	if opts.Branch != "" {
		branchOptionsCount++
	}
	if opts.BranchRegex != "" {
		branchOptionsCount++
	}
	if opts.AllBranches {
		branchOptionsCount++
	}

	if branchOptionsCount > 1 {
		return fmt.Errorf("please specify only one branch option (--branch, --branch-regex, or --all-branches)")
	}
	return nil
}

// handleDefaultReport handles the default report when no specific options are provided
func handleDefaultReport(records []tracker.CheckpointRecord, config *tracker.Config, opts *ReportOptions) {
	if len(records) > 0 {
		analyzer := tracker.NewAnalyzer(config)
		report, err := analyzer.GenerateReportFromRecords(records, 0)
		if err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(report)
	} else {
		// Fallback to legacy metrics
		baseDir := defaultBaseDir
		metricsStorage := storage.NewMetricsStorage(baseDir)
		metrics, err := metricsStorage.LoadMetrics()
		if err != nil {
			fmt.Printf("Error loading metrics: %v\n", err)
			os.Exit(1)
		}

		analyzer := tracker.NewAnalyzer(config)
		fmt.Println(analyzer.GenerateReport(metrics))
	}
}

// handleCombinedReport handles reports with both period and branch filtering
func handleCombinedReport(records []tracker.CheckpointRecord, config *tracker.Config, opts *ReportOptions) {
	// Step 1: Apply period filtering first
	var timeRange *period.TimeRange
	var err error

	// Parse time range based on options (same logic as handlePeriodReport)
	if opts.Last != "" {
		timeRange, err = period.ParseLastDuration(opts.Last)
		if err != nil {
			fmt.Printf("Error parsing last duration: %v\n", err)
			os.Exit(1)
		}
	} else if opts.Since != "" {
		timeRange, err = period.ParseTimeRange(opts.Since)
		if err != nil {
			fmt.Printf("Error parsing since time: %v\n", err)
			os.Exit(1)
		}
	} else if opts.From != "" && opts.To != "" {
		timeRange, err = period.ParseFromTo(opts.From, opts.To)
		if err != nil {
			fmt.Printf("Error parsing from/to range: %v\n", err)
			os.Exit(1)
		}
	} else if opts.From != "" {
		// From only - use current time as end
		fromTime, err := period.ParseTimeRange(opts.From)
		if err != nil {
			fmt.Printf("Error parsing from time: %v\n", err)
			os.Exit(1)
		}
		timeRange = &period.TimeRange{
			From: fromTime.From,
			To:   time.Now(),
		}
	} else {
		fmt.Printf("Error: Please specify a time range using --since, --from/--to, or --last\n")
		os.Exit(1)
	}

	// Filter records by time range
	var filteredRecords []tracker.CheckpointRecord
	for _, record := range records {
		if record.Timestamp.After(timeRange.From) && record.Timestamp.Before(timeRange.To) {
			filteredRecords = append(filteredRecords, record)
		}
	}

	if len(filteredRecords) == 0 {
		fmt.Printf("No records found in the specified time range.\n")
		return
	}

	// Step 2: Apply branch filtering to time-filtered records
	analyzer := branch.NewBranchAnalyzer(filteredRecords)
	targetPercentage := config.TargetAIPercentage

	// Display period information
	fmt.Printf("Combined Report: Period + Branch Filtering\n")
	fmt.Printf("==========================================\n")
	fmt.Printf("Period: %s to %s\n", 
		timeRange.From.Format("2006-01-02"), 
		timeRange.To.Format("2006-01-02"))
	fmt.Printf("Filtered Records: %d (from %d total)\n\n", len(filteredRecords), len(records))

	// Apply branch filtering (same logic as handleBranchReport)
	if opts.AllBranches {
		handleAllBranchesCombinedReport(analyzer, targetPercentage)
	} else if opts.Branch != "" {
		handleSingleBranchCombinedReport(analyzer, opts.Branch, targetPercentage)
	} else if opts.BranchRegex != "" {
		handleRegexBranchCombinedReport(analyzer, opts.BranchRegex, targetPercentage)
	}
}

// handleAllBranchesCombinedReport shows all branches summary for combined filtering
func handleAllBranchesCombinedReport(analyzer *branch.BranchAnalyzer, targetPercentage float64) {
	groupReport, err := analyzer.AnalyzeAllBranches()
	if err != nil {
		fmt.Printf("Error analyzing all branches: %v\n", err)
		os.Exit(1)
	}

	if len(groupReport.MatchingBranches) == 0 {
		fmt.Println("No branches found in filtered records.")
		return
	}

	fmt.Printf("All Branches Summary (Filtered):\n")
	fmt.Printf("=================================\n")
	
	// Show overall stats for filtered data
	stats := analyzer.GetRecordStats()
	fmt.Printf("  Filtered Records: %d\n", stats.TotalRecords)
	fmt.Printf("  Unique Branches: %d\n", stats.UniqueBranches)
	fmt.Printf("  Records with Branch Info: %d\n", stats.RecordsWithBranch)
	fmt.Printf("  Records without Branch Info: %d\n\n", stats.RecordsWithoutBranch)

	fmt.Printf("Group Summary:\n")
	fmt.Printf("  Total Added Lines: %d\n", groupReport.TotalAdded)
	fmt.Printf("  Group AI Ratio: %.1f%% (target: %.1f%%)\n", groupReport.GroupAIRatio, targetPercentage)
	
	// Progress indicator
	if groupReport.GroupAIRatio >= targetPercentage {
		fmt.Printf("  Progress: ✅ Target achieved (%.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100)
	} else {
		fmt.Printf("  Progress: 📊 %.1f%% of target\n\n", (groupReport.GroupAIRatio/targetPercentage)*100)
	}

	fmt.Printf("Per-Branch Breakdown:\n")
	for _, branchName := range groupReport.MatchingBranches {
		branchReport := groupReport.BranchReports[branchName]
		fmt.Printf("  %s: AI %.1f%% (%d/%d lines) [%d records]\n", 
			branchName, 
			branchReport.AIRatio, 
			int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100), 
			branchReport.TotalAdded,
			branchReport.RecordCount)
	}
}

// handleSingleBranchCombinedReport shows detailed report for a specific branch with period filtering
func handleSingleBranchCombinedReport(analyzer *branch.BranchAnalyzer, branchName string, targetPercentage float64) {
	branchReport, err := analyzer.AnalyzeByBranch(branchName)
	if err != nil {
		fmt.Printf("Error analyzing branch '%s': %v\n", branchName, err)
		os.Exit(1)
	}

	if branchReport.RecordCount == 0 {
		fmt.Printf("No records found for branch '%s' in the specified period.\n", branchName)
		fmt.Println("\nAvailable branches in filtered data:")
		branches := analyzer.GetUniqueBranches()
		for _, branch := range branches {
			fmt.Printf("  %s\n", branch)
		}
		return
	}

	fmt.Printf("Branch Report (Filtered): %s\n", branchName)
	fmt.Printf("==================================\n")
	fmt.Printf("Records: %d (%s to %s)\n", 
		branchReport.RecordCount,
		branchReport.FirstRecord.Format("2006-01-02"),
		branchReport.LastRecord.Format("2006-01-02"))
	fmt.Printf("Added Lines: %d (AI: %d, Human: %d)\n", 
		branchReport.TotalAdded,
		int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100),
		branchReport.TotalAdded-int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100))
	fmt.Printf("AI Ratio: %.1f%%\n", branchReport.AIRatio)
	
	if branchReport.AIRatio >= targetPercentage {
		fmt.Printf("Progress: ✅ %.1f%% (target: %.1f%%)\n", (branchReport.AIRatio/targetPercentage)*100, targetPercentage)
	} else {
		fmt.Printf("Progress: 📊 %.1f%% (target: %.1f%%)\n", (branchReport.AIRatio/targetPercentage)*100, targetPercentage)
	}

	if len(branchReport.Authors) > 0 {
		fmt.Printf("Authors: %s\n", strings.Join(branchReport.Authors, ", "))
	}
}

// handleRegexBranchCombinedReport shows report for branches matching a regex pattern with period filtering
func handleRegexBranchCombinedReport(analyzer *branch.BranchAnalyzer, pattern string, targetPercentage float64) {
	groupReport, err := analyzer.AnalyzeByPattern(pattern, true)
	if err != nil {
		fmt.Printf("Error analyzing branches with pattern '%s': %v\n", pattern, err)
		os.Exit(1)
	}

	if len(groupReport.MatchingBranches) == 0 {
		fmt.Printf("No branches matching pattern '%s' found in filtered records.\n", pattern)
		return
	}

	fmt.Printf("Branch Pattern Report (Filtered): \"%s\"\n", pattern)
	fmt.Printf("==========================================\n")
	fmt.Printf("Matching Branches: %s\n", strings.Join(groupReport.MatchingBranches, ", "))
	fmt.Printf("Total Records: %d\n", groupReport.TotalRecords)
	fmt.Printf("Added Lines: %d (AI: %d, Human: %d)\n", 
		groupReport.TotalAdded,
		int(float64(groupReport.TotalAdded)*groupReport.GroupAIRatio/100),
		groupReport.TotalAdded-int(float64(groupReport.TotalAdded)*groupReport.GroupAIRatio/100))
	fmt.Printf("Group AI Ratio: %.1f%%\n", groupReport.GroupAIRatio)
	
	if groupReport.GroupAIRatio >= targetPercentage {
		fmt.Printf("Progress: ✅ %.1f%% (target: %.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100, targetPercentage)
	} else {
		fmt.Printf("Progress: 📊 %.1f%% (target: %.1f%%)\n\n", (groupReport.GroupAIRatio/targetPercentage)*100, targetPercentage)
	}

	fmt.Printf("Per-Branch Breakdown:\n")
	for _, branchName := range groupReport.MatchingBranches {
		branchReport := groupReport.BranchReports[branchName]
		fmt.Printf("  %s: AI %.1f%% (%d/%d lines) [%d records]\n", 
			branchName, 
			branchReport.AIRatio, 
			int(float64(branchReport.TotalAdded)*branchReport.AIRatio/100), 
			branchReport.TotalAdded,
			branchReport.RecordCount)
	}
}
