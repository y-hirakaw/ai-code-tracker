package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/period"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// ReportOptions holds options for the report command
type ReportOptions struct {
	Since  string
	From   string
	To     string
	Last   string
	Format string
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

	// Check if period options are specified
	if opts.Since != "" || opts.From != "" || opts.Last != "" {
		handlePeriodReport(records, config, opts)
		return
	}

	// Default report (existing functionality)
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
		metrics, err := metricsStorage.LoadMetrics()
		if err != nil {
			fmt.Printf("Error loading metrics: %v\n", err)
			os.Exit(1)
		}

		analyzer := tracker.NewAnalyzer(config)
		fmt.Println(analyzer.GenerateReport(metrics))
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
