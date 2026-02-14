package status

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// RunStatusCommand implements the `darkpipe status` CLI command
func RunStatusCommand(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output JSON for scripting/Home Assistant integration")
	watch := fs.Bool("watch", false, "Auto-refresh every N seconds")
	watchInterval := fs.Int("watch-interval", 5, "Refresh interval in seconds (for --watch mode)")

	fs.Parse(args)

	// TODO: Initialize aggregator with real dependencies
	// For now, this is a skeleton that will be wired up in main.go
	// aggregator := initializeAggregator()

	if *watch {
		runWatchMode(*watchInterval, *jsonOutput)
	} else {
		runOnce(*jsonOutput)
	}
}

func runOnce(jsonOutput bool) {
	// TODO: Replace with actual aggregator call
	// status, err := aggregator.GetStatus(context.Background())
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }

	// Placeholder implementation
	fmt.Fprintln(os.Stderr, "Status command not yet wired to aggregator")
	os.Exit(1)
}

func runWatchMode(interval int, jsonOutput bool) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		clearTerminal()
		runOnce(jsonOutput)
		<-ticker.C
	}
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

// DisplayStatus outputs the system status in human-readable format
func DisplayStatus(status *SystemStatus) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	statusColor := green
	statusText := "HEALTHY"
	switch status.OverallStatus {
	case "degraded":
		statusColor = yellow
		statusText = "DEGRADED"
	case "unhealthy":
		statusColor = red
		statusText = "UNHEALTHY"
	}

	fmt.Println("DarkPipe System Status")
	fmt.Println("======================")
	fmt.Printf("Overall: %s\n\n", statusColor(statusText))

	// Services section
	fmt.Println("Services:")
	for _, check := range status.Health.Checks {
		checkStatus := green("OK")
		if check.Status != "ok" {
			checkStatus = red("FAIL")
		}
		fmt.Printf("  %-20s %s\n", check.Name+":", checkStatus)
	}
	fmt.Println()

	// Mail Queue section
	fmt.Println("Mail Queue:")
	fmt.Printf("  Depth:    %d\n", status.Queue.Depth)
	fmt.Printf("  Deferred: %d\n", status.Queue.Deferred)
	fmt.Printf("  Stuck:    %d\n", status.Queue.Stuck)
	fmt.Println()

	// Recent Deliveries section
	fmt.Println("Recent Deliveries (last 24h):")
	fmt.Printf("  Delivered: %d\n", status.Delivery.Delivered)
	fmt.Printf("  Deferred:  %d\n", status.Delivery.Deferred)
	fmt.Printf("  Bounced:   %d\n", status.Delivery.Bounced)
	fmt.Println()

	// Certificates section
	fmt.Println("Certificates:")
	for _, cert := range status.Certificates.Certificates {
		daysLeft := cert.DaysLeft
		certStatus := green("OK")
		if daysLeft <= 7 {
			certStatus = red("CRITICAL")
		} else if daysLeft <= 14 {
			certStatus = yellow("WARNING")
		}

		certName := cert.Subject
		if certName == "" {
			certName = cert.Path
		}

		fmt.Printf("  %-30s %3d days remaining (%s)\n", certName+":", daysLeft, certStatus)
	}
	fmt.Println()

	// Check for pending CLI alerts
	checkCLIAlerts()
}

// DisplayStatusJSON outputs the system status as JSON
func DisplayStatusJSON(status *SystemStatus) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(status)
}

// checkCLIAlerts looks for pending alerts in the CLI alerts file
func checkCLIAlerts() {
	alertPath := os.Getenv("MONITOR_CLI_ALERT_PATH")
	if alertPath == "" {
		alertPath = "/data/monitoring/cli-alerts.json"
	}

	// Check if file exists and has content
	info, err := os.Stat(alertPath)
	if err != nil || info.Size() == 0 {
		return
	}

	// Count lines in NDJSON file
	data, err := os.ReadFile(alertPath)
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	if count > 0 {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s\n", yellow(fmt.Sprintf("WARNING: %d pending alert(s). Run 'darkpipe alerts' to view.", count)))
	}
}
