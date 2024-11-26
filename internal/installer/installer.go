package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Abhaythakor/dev-tools-installer/internal/config"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"
	clearLine   = "\033[K"
)

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Progress represents a progress indicator
type Progress struct {
	message string
	stop    chan bool
	stopped bool
	mu      sync.Mutex
}

// NewProgress creates a new progress indicator
func NewProgress(message string) *Progress {
	return &Progress{
		message: message,
		stop:    make(chan bool),
		stopped: false,
	}
}

// Start starts the progress indicator
func (p *Progress) Start() {
	go func() {
		i := 0
		for {
			p.mu.Lock()
			if p.stopped {
				p.mu.Unlock()
				return
			}
			p.mu.Unlock()

			select {
			case <-p.stop:
				// Clear the line before returning
				fmt.Printf("\r%s", strings.Repeat(" ", 80))
				fmt.Printf("\r")
				return
			default:
				fmt.Printf("\r%s│ %s%s %s",
					colorBlue,
					colorYellow,
					spinnerChars[i%len(spinnerChars)],
					p.message)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the progress indicator
func (p *Progress) Stop() {
	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()
	close(p.stop)
}

// Installer manages tool installation
type Installer struct {
	config *config.InstallerConfig
}

// New creates a new Installer instance
func New(config *config.InstallerConfig) *Installer {
	return &Installer{
		config: config,
	}
}

// Run checks and installs tools as needed
func (i *Installer) Run() error {
	fmt.Printf("\n%s╭─── System Tools Check ───╮%s\n", colorBlue+"\033[1m", colorReset)

	installed := 0
	for _, name := range i.config.ToolList {
		if i.checkTool(name) {
			installed++
		} else {
			if err := i.installTool(name); err != nil {
				fmt.Printf("%s│%s Failed to install %s: %v%s\n", colorBlue, colorRed, name, err, colorReset)
				continue
			}
			installed++
		}
	}

	fmt.Printf("%s╰─── %s%d/%d tools installed %s───╯%s\n\n",
		colorBlue,
		colorGreen,
		installed,
		len(i.config.ToolList),
		colorBlue,
		colorReset)

	return nil
}

// checkTool checks if a tool is installed and returns true if installed
func (i *Installer) checkTool(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		fmt.Printf("%s│ %s✗ %-9s%s │ Not installed\n", colorBlue, colorRed, name, colorReset)
		return false
	}

	version := i.getToolVersion(name)
	if version == "" {
		fmt.Printf("%s│ %s✓ %-9s%s │ Installed (version unknown)\n", colorBlue, colorGreen, name, colorReset)
		return true
	}

	fmt.Printf("%s│ %s✓ %-9s%s │ %s%s\n", colorBlue, colorGreen, name, colorReset, version, colorReset)
	return true
}

// getToolVersion returns the version of a tool
func (i *Installer) getToolVersion(name string) string {
	// If version is defined in YAML, use that
	if version := i.config.Tools[name].Version; version != "" {
		return version
	}

	// Common version flags to try
	versionFlags := []string{
		"--version", // Most common
		"-version",  // Some tools like subfinder
		"version",   // Tools like go, amass
		"-v",        // Short version
		"-V",        // Uppercase short version
		"--ver",     // Abbreviated
		"-ver",      // Abbreviated with single dash
	}

	// Get version flag from config if specified
	if i.config.Tools[name].VersionFlag != "" {
		versionFlags = []string{i.config.Tools[name].VersionFlag}
	}

	var version string
	for _, flag := range versionFlags {
		cmd := exec.Command(name, flag)
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		// Try to extract version from output
		version = extractVersion(string(output))
		if version != "" {
			break
		}
	}

	return version
}

// installTool attempts to install a tool using the first available method
func (i *Installer) installTool(name string) error {
	toolConfig := i.config.Tools[name]
	if toolConfig == nil || len(toolConfig.Methods) == 0 {
		return fmt.Errorf("no installation methods available for %s", name)
	}

	// Try each installation method until one succeeds
	for _, method := range toolConfig.Methods {
		fmt.Printf("%s│%s 📦 Installing %s using %s method...%s\n", colorBlue, colorYellow, name, method.Name, colorReset)

		for _, command := range method.Commands {
			// Replace environment variables and version
			command = os.ExpandEnv(command)
			if version := toolConfig.Version; version != "" {
				command = strings.ReplaceAll(command, "${version}", version)
			}

			// Split the command into parts
			parts := strings.Fields(command)
			if len(parts) == 0 {
				continue
			}

			// Create the command
			execCmd := exec.Command(parts[0], parts[1:]...)

			// Create pipes for stdout and stderr
			stdout, err := execCmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create stdout pipe: %v", err)
			}
			stderr, err := execCmd.StderrPipe()
			if err != nil {
				return fmt.Errorf("failed to create stderr pipe: %v", err)
			}

			// Start the command
			if err := execCmd.Start(); err != nil {
				fmt.Printf("%s│%s ❌ Failed to start command: %s%s\n", colorBlue, colorRed, command, colorReset)
				continue
			}

			// Create progress indicator with tool name and method
			progress := NewProgress(fmt.Sprintf("Installing %s (%s): %s", name, method.Name, filepath.Base(parts[0])))
			progress.Start()

			// Create a WaitGroup for the scanner goroutine
			var wg sync.WaitGroup
			wg.Add(1)

			// Read command output in the background
			go func() {
				defer wg.Done()
				scanner := NewSafeScanner(io.MultiReader(stdout, stderr))
				for scanner.Scan() {
					line := scanner.Text()
					// Only show output for go install commands
					if strings.Contains(command, "go install") || strings.Contains(command, "go get") {
						if show, formatted := formatGoInstallOutput(line); show {
							progress.Stop()
							fmt.Printf("%s│ %s%s%s\n", colorBlue, colorGray, formatted, colorReset)
							progress = NewProgress(fmt.Sprintf("Installing %s (%s): %s", name, method.Name, filepath.Base(parts[0])))
							progress.Start()
						}
					}
				}
			}()

			// Wait for command to complete
			err = execCmd.Wait()

			// Wait for scanner to finish
			wg.Wait()

			// Stop the progress indicator and clear the line
			progress.Stop()
			fmt.Printf("\r%s", strings.Repeat(" ", 80)) // Clear the line
			fmt.Printf("\r")                            // Return to start of line

			if err != nil {
				fmt.Printf("%s│%s ❌ Failed to install %s: %v%s\n", colorBlue, colorRed, name, err, colorReset)
				continue
			}

			return nil
		}
	}

	return fmt.Errorf("all installation methods failed for %s", name)
}

// SafeScanner wraps bufio.Scanner with error handling
type SafeScanner struct {
	*bufio.Scanner
}

// NewSafeScanner creates a new SafeScanner
func NewSafeScanner(r io.Reader) *SafeScanner {
	return &SafeScanner{bufio.NewScanner(r)}
}

// formatGoInstallOutput formats the output of go install commands
func formatGoInstallOutput(line string) (bool, string) {
	if strings.Contains(line, "go: downloading") {
		return true, line
	}
	return false, ""
}

// extractVersion extracts version information from command output
func extractVersion(output string) string {
	// Common version patterns
	patterns := []string{
		`(?i)version\s+(v\d+\.\d+\.\d+)`, // matches version v1.2.3 (case insensitive)
		`(?i)amass\s+-\s+v\d+\.\d+\.\d+`, // matches amass - v1.2.3
		`v\d+\.\d+\.\d+`,                 // matches v1.2.3
		`\d+\.\d+\.\d+`,                  // matches 1.2.3
		`go\d+\.\d+\.\d+`,                // matches go1.2.3
		`Version: (v\d+\.\d+\.\d+)`,      // matches Version: v1.2.3
	}

	version := strings.TrimSpace(output)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if strings.Contains(pattern, "(") {
			// Handle patterns with capture groups
			if match := re.FindStringSubmatch(version); len(match) > 1 {
				return match[1]
			}
		} else {
			// Handle simple patterns
			if match := re.FindString(version); match != "" {
				// Clean up amass version format
				if strings.Contains(match, "amass - ") {
					return strings.TrimPrefix(strings.TrimPrefix(match, "amass - "), "v")
				}
				// Clean up go version format
				if strings.HasPrefix(match, "go") {
					return strings.TrimPrefix(match, "go")
				}
				return match
			}
		}
	}

	// If no version pattern matched, return first line
	if lines := strings.Split(version, "\n"); len(lines) > 0 {
		return lines[0]
	}

	return ""
}
