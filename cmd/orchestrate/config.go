package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/cpuid/v2"
	"gopkg.in/yaml.v3"
)

// SetupLogging configures the global logger based on quiet flag.
// When quiet=true, only errors are shown. Otherwise, info and above.
func SetupLogging(quiet bool) {
	level := slog.LevelInfo
	if quiet {
		level = slog.LevelError
	}
	handler := &logHandler{
		level: level,
		w:     os.Stderr,
	}
	slog.SetDefault(slog.New(handler))
}

// logHandler is a minimal slog handler that outputs clean messages.
type logHandler struct {
	level slog.Level
	w     io.Writer
	attrs []slog.Attr
}

func (h *logHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *logHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message

	// For warnings/errors, prefix the message.
	switch r.Level {
	case slog.LevelWarn:
		msg = "Warning: " + msg
	case slog.LevelError:
		msg = "ERROR: " + msg
	}

	// Append any attributes as key=value.
	r.Attrs(func(a slog.Attr) bool {
		msg += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		return true
	})

	fmt.Fprintln(h.w, msg)
	return nil
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{level: h.level, w: h.w, attrs: append(h.attrs, attrs...)}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return h // groups not used
}

// optimalThreshold is the fraction of peak throughput used to determine
// the "optimal" configuration (smallest routine count achieving this threshold).
const optimalThreshold = 0.95

// Scenario defines a complete benchmark scenario.
type Scenario struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Defaults    RunConfig   `yaml:"defaults"`
	Runs        []RunGroup  `yaml:"runs"`
}

// RunConfig holds benchmark configuration values.
type RunConfig struct {
	Duration    string `yaml:"duration"`
	Concurrency int    `yaml:"concurrency"`
	Seed        int    `yaml:"seed"`
	Rate        int    `yaml:"rate"`
}

// RunGroup defines a group of benchmark runs with varying routines.
type RunGroup struct {
	Name          string `yaml:"name"`
	Mode          string `yaml:"mode"`
	Routines      []int  `yaml:"routines"`
	OtherRoutines int    `yaml:"other_routines"`
	RunConfig     `yaml:",inline"`
}

// BenchmarkRun is a fully-resolved single benchmark execution.
type BenchmarkRun struct {
	Name          string
	Mode          string
	Routines      int
	OtherRoutines int
	Duration      time.Duration
	Concurrency   int
	Seed          int
	Rate          int
}

// BenchConfig holds CLI benchmark configuration.
type BenchConfig struct {
	RepoRoot     string
	ScenarioPath string
	RoutinesStr  string
	Mode         string
	DurationStr  string
	Seed         int
	Concurrency  int
	Rate         int
}

// LoadScenario loads a scenario from a YAML file.
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario file: %w", err)
	}
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario YAML: %w", err)
	}
	return &s, nil
}

// BuildInlineScenario creates a scenario from CLI flags.
func BuildInlineScenario(mode string, routines []int, duration time.Duration, seed, concurrency, rate int) *Scenario {
	return &Scenario{
		Name:        "inline",
		Description: "CLI-generated scenario",
		Defaults: RunConfig{
			Duration:    duration.String(),
			Concurrency: concurrency,
			Seed:        seed,
			Rate:        rate,
		},
		Runs: []RunGroup{
			{
				Name:          fmt.Sprintf("%s-sweep", mode),
				Mode:          mode,
				Routines:      routines,
				OtherRoutines: 1,
			},
		},
	}
}

// Expand converts a Scenario into individual BenchmarkRuns.
func (s *Scenario) Expand() ([]BenchmarkRun, error) {
	defaultDur, err := time.ParseDuration(s.Defaults.Duration)
	if err != nil && s.Defaults.Duration != "" {
		return nil, fmt.Errorf("parse default duration: %w", err)
	}
	if defaultDur == 0 {
		defaultDur = 2 * time.Minute
	}
	defaultConc := s.Defaults.Concurrency
	if defaultConc == 0 {
		defaultConc = 64
	}
	defaultSeed := s.Defaults.Seed
	if defaultSeed == 0 {
		defaultSeed = 2048
	}

	var runs []BenchmarkRun
	for _, rg := range s.Runs {
		// Resolve duration (group override takes precedence).
		dur := defaultDur
		if rg.Duration != "" {
			d, err := time.ParseDuration(rg.Duration)
			if err != nil {
				return nil, fmt.Errorf("parse duration for %s: %w", rg.Name, err)
			}
			dur = d
		}

		// Resolve other fields.
		conc := defaultConc
		if rg.Concurrency > 0 {
			conc = rg.Concurrency
		}
		seed := defaultSeed
		if rg.Seed > 0 {
			seed = rg.Seed
		}
		rate := s.Defaults.Rate
		if rg.Rate > 0 {
			rate = rg.Rate
		}

		otherRoutines := rg.OtherRoutines
		if otherRoutines == 0 {
			otherRoutines = 1
		} else if otherRoutines < 1 {
			return nil, fmt.Errorf("invalid other_routines %d in run %q: must be >= 1", rg.OtherRoutines, rg.Name)
		}

		// Expand across all routine counts.
		for _, r := range rg.Routines {
			if r < 1 {
				return nil, fmt.Errorf("invalid routine count %d in run %q: must be >= 1", r, rg.Name)
			}
			runs = append(runs, BenchmarkRun{
				Name:          fmt.Sprintf("%s-%d", rg.Name, r),
				Mode:          rg.Mode,
				Routines:      r,
				OtherRoutines: otherRoutines,
				Duration:      dur,
				Concurrency:   conc,
				Seed:          seed,
				Rate:          rate,
			})
		}
	}
	return runs, nil
}

// ParseRoutinesList parses a comma-separated list of integers.
func ParseRoutinesList(s string) ([]int, error) {
	if s == "" {
		return []int{4, 8, 16, 32}, nil
	}
	parts := strings.Split(s, ",")
	var result []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid routine count %q: %w", p, err)
		}
		if n < 1 {
			return nil, fmt.Errorf("routine count must be >= 1: %d", n)
		}
		result = append(result, n)
	}
	if len(result) == 0 {
		return []int{4, 8, 16, 32}, nil
	}
	return result, nil
}

// LoadOrBuildScenario loads from file or builds from CLI flags.
func LoadOrBuildScenario(cfg BenchConfig) (*Scenario, error) {
	if cfg.ScenarioPath != "" {
		scenarioPath := cfg.ScenarioPath
		if !filepath.IsAbs(scenarioPath) {
			scenarioPath = filepath.Join(cfg.RepoRoot, scenarioPath)
		}
		return LoadScenario(scenarioPath)
	}

	routines, err := ParseRoutinesList(cfg.RoutinesStr)
	if err != nil {
		return nil, fmt.Errorf("parse routines: %w", err)
	}
	duration, err := time.ParseDuration(cfg.DurationStr)
	if err != nil {
		return nil, fmt.Errorf("parse duration: %w", err)
	}
	return BuildInlineScenario(cfg.Mode, routines, duration, cfg.Seed, cfg.Concurrency, cfg.Rate), nil
}

// SetupOutputDir creates and returns the output directory.
func SetupOutputDir(repoRoot, outputDir, mode string) (string, error) {
	if outputDir == "" {
		outputDir = filepath.Join(repoRoot, "bench", "results", fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), mode))
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	return outputDir, nil
}

// BuildBenchArgs constructs benchmark command arguments.
func BuildBenchArgs(run BenchmarkRun, target, resultFile, statsURL string) []string {
	args := []string{
		"-mode", run.Mode,
		"-target", target,
		"-duration", run.Duration.String(),
		"-concurrency", fmt.Sprintf("%d", run.Concurrency),
		"-seed", fmt.Sprintf("%d", run.Seed),
		"-out", resultFile,
		"-quiet",
	}
	if run.Rate > 0 {
		args = append(args, "-rate", fmt.Sprintf("%d", run.Rate))
	}
	if statsURL != "" {
		args = append(args, "-stats", statsURL)
	}
	return args
}

// DetermineRoutines calculates UDP/HTTP routine counts based on mode.
func DetermineRoutines(run BenchmarkRun) (udpRoutines, httpRoutines int) {
	udpRoutines = run.OtherRoutines
	httpRoutines = run.OtherRoutines
	if run.Mode == "udp" {
		udpRoutines = run.Routines
	} else if run.Mode == "http" {
		httpRoutines = run.Routines
	}
	return udpRoutines, httpRoutines
}

// WaitForInput waits for user to press Enter.
func WaitForInput() error {
	_, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return err
}

// FindRepoRoot attempts to find the repository root.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goMod); err == nil {
			if strings.Contains(string(data), "github.com/crimist/trakx") {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find trakx repository root")
}

// EnsureTrakxBinary builds the trakx binary if it doesn't exist.
func EnsureTrakxBinary(repoRoot, binPath string) error {
	if binPath == "" {
		binPath = filepath.Join(repoRoot, "bench", "bin", "trakx")
	}
	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/trakx")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnsureBenchBinary builds the bench binary if it doesn't exist.
func EnsureBenchBinary(repoRoot, binPath string) error {
	if binPath == "" {
		binPath = filepath.Join(repoRoot, "bench", "bin", "bench")
	}
	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/bench")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunResult captures the result of a single benchmark run.
type RunResult struct {
	Sequence    int                    `json:"sequence"`
	RunID       string                 `json:"run_id"`
	Mode        string                 `json:"mode"`
	Routines    int                    `json:"routines"`
	OtherRoutes int                    `json:"other_routines"`
	ResultFile  string                 `json:"result_file"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	DurationMS  int64                  `json:"duration_ms"`
	Summary     *ResultSummary         `json:"summary,omitempty"`
	RawResult   map[string]interface{} `json:"-"`
}

// ResultSummary is a condensed view of benchmark results.
type ResultSummary struct {
	Throughput float64 `json:"throughput"`
	P50        string  `json:"p50"`
	P90        string  `json:"p90"`
	P95        string  `json:"p95"`
	P99        string  `json:"p99"`
	Success    int64   `json:"success"`
	Errors     int64   `json:"errors"`
	Timeouts   int64   `json:"timeouts"`
}

// OrchestrationResult is the complete orchestration output.
type OrchestrationResult struct {
	Version     string                 `json:"version"`
	Timestamp   string                 `json:"timestamp"`
	Scenario    string                 `json:"scenario"`
	Mode        string                 `json:"mode"`
	Runs        []RunResult            `json:"runs"`
	Analysis    *Analysis              `json:"analysis,omitempty"`
	Environment map[string]interface{} `json:"environment"`
}

// Analysis provides insights from the benchmark runs.
type Analysis struct {
	PeakThroughput float64       `json:"peak_throughput"`
	PeakConfig     OptimalConfig `json:"peak_config"`
	OptimalConfig  OptimalConfig `json:"optimal_config"`
	OptimalReason  string        `json:"optimal_reason"`
}

// OptimalConfig describes an optimal configuration.
type OptimalConfig struct {
	Routines   int     `json:"routines"`
	Throughput float64 `json:"throughput"`
	P99        string  `json:"p99"`
}

// ResultCollector collects and aggregates benchmark results.
type ResultCollector struct {
	OutputDir string
	Scenario  string
	Mode      string
	StartTime time.Time

	mu      sync.Mutex
	results []RunResult
}

// NewResultCollector creates a new result collector.
func NewResultCollector(outputDir, scenario, mode string) *ResultCollector {
	return &ResultCollector{
		OutputDir: outputDir,
		Scenario:  scenario,
		Mode:      mode,
		results:   make([]RunResult, 0),
		StartTime: time.Now(),
	}
}

// AddResult adds a benchmark result.
func (rc *ResultCollector) AddResult(result RunResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	result.Sequence = len(rc.results) + 1
	rc.results = append(rc.results, result)
}

// GetResults returns a copy of all collected results.
func (rc *ResultCollector) GetResults() []RunResult {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	results := make([]RunResult, len(rc.results))
	copy(results, rc.results)
	return results
}

// LoadResultFile loads and parses a bench result file.
func LoadResultFile(path string) (*ResultSummary, map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, err
	}

	summary := &ResultSummary{}

	if metrics, ok := raw["metrics"].(map[string]interface{}); ok {
		for _, v := range metrics {
			if m, ok := v.(map[string]interface{}); ok {
				if rate, ok := m["rate_per_sec"].(float64); ok {
					summary.Throughput = rate
				}
				if success, ok := m["success"].(float64); ok {
					summary.Success = int64(success)
				}
				if errors, ok := m["errors"].(float64); ok {
					summary.Errors = int64(errors)
				}
				if timeouts, ok := m["timeouts"].(float64); ok {
					summary.Timeouts = int64(timeouts)
				}
				if lat, ok := m["latency"].(map[string]interface{}); ok {
					if p50, ok := lat["p50"].(string); ok {
						summary.P50 = p50
					}
					if p90, ok := lat["p90"].(string); ok {
						summary.P90 = p90
					}
					if p95, ok := lat["p95"].(string); ok {
						summary.P95 = p95
					}
					if p99, ok := lat["p99"].(string); ok {
						summary.P99 = p99
					}
				}
				break
			}
		}
	}

	return summary, raw, nil
}

// Analyze performs analysis on collected results.
func (rc *ResultCollector) Analyze() *Analysis {
	results := rc.GetResults()
	if len(results) == 0 {
		return nil
	}

	var peak RunResult
	var peakThroughput float64
	for _, r := range results {
		if r.Summary != nil && r.Summary.Throughput > peakThroughput {
			peakThroughput = r.Summary.Throughput
			peak = r
		}
	}

	if peakThroughput == 0 {
		return nil
	}

	threshold := peakThroughput * optimalThreshold
	var optimal RunResult
	optimalFound := false

	sorted := make([]RunResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Routines < sorted[j].Routines
	})

	for _, r := range sorted {
		if r.Summary != nil && r.Summary.Throughput >= threshold {
			optimal = r
			optimalFound = true
			break
		}
	}

	if !optimalFound {
		optimal = peak
	}

	analysis := &Analysis{
		PeakThroughput: peakThroughput,
		PeakConfig: OptimalConfig{
			Routines:   peak.Routines,
			Throughput: peakThroughput,
		},
		OptimalConfig: OptimalConfig{
			Routines: optimal.Routines,
		},
	}

	if peak.Summary != nil {
		analysis.PeakConfig.P99 = peak.Summary.P99
	}
	if optimal.Summary != nil {
		analysis.OptimalConfig.Throughput = optimal.Summary.Throughput
		analysis.OptimalConfig.P99 = optimal.Summary.P99
	}

	if optimal.Routines == peak.Routines {
		analysis.OptimalReason = "Peak throughput configuration"
	} else {
		analysis.OptimalReason = fmt.Sprintf("Smallest config achieving >= %.0f%% of peak (%.0f req/s)", optimalThreshold*100, threshold)
	}

	return analysis
}

// WriteResults writes the complete orchestration results.
func (rc *ResultCollector) WriteResults() error {
	hostname, _ := os.Hostname()

	result := OrchestrationResult{
		Version:   version,
		Timestamp: rc.StartTime.Format(time.RFC3339),
		Scenario:  rc.Scenario,
		Mode:      rc.Mode,
		Runs:      rc.GetResults(),
		Analysis:  rc.Analyze(),
		Environment: map[string]interface{}{
			"hostname":   hostname,
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"num_cpu":    runtime.NumCPU(),
			"cpu_model":  cpuid.CPU.BrandName,
		},
	}

	summaryPath := filepath.Join(rc.OutputDir, "summary.json")
	f, err := os.Create(summaryPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// PrintSummary prints a human-readable summary.
func (rc *ResultCollector) PrintSummary() {
	results := rc.GetResults()

	fmt.Println()
	fmt.Println("=== Results Summary ===")
	fmt.Println()

	for _, r := range results {
		status := r.Status
		if r.Summary != nil {
			fmt.Printf("%-20s %8.0f req/s  p99: %-10s  [%s]\n",
				r.RunID,
				r.Summary.Throughput,
				r.Summary.P99,
				status)
		} else {
			fmt.Printf("%-20s  %s: %s\n", r.RunID, status, r.Error)
		}
	}

	analysis := rc.Analyze()
	if analysis != nil {
		fmt.Println()
		fmt.Printf("Peak:    %d routines @ %.0f req/s (p99: %s)\n",
			analysis.PeakConfig.Routines,
			analysis.PeakConfig.Throughput,
			analysis.PeakConfig.P99)
		fmt.Printf("Optimal: %d routines @ %.0f req/s (p99: %s)\n",
			analysis.OptimalConfig.Routines,
			analysis.OptimalConfig.Throughput,
			analysis.OptimalConfig.P99)
		fmt.Printf("         %s\n", analysis.OptimalReason)
	}

	fmt.Println()
	fmt.Printf("Results: %s\n", rc.OutputDir)
}

// FinalizeResults writes results and prints summary.
func FinalizeResults(collector *ResultCollector) error {
	if err := collector.WriteResults(); err != nil {
		return fmt.Errorf("write results: %w", err)
	}
	collector.PrintSummary()
	return nil
}
