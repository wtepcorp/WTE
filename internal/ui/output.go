package ui

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Colors
var (
	Red     = color.New(color.FgRed)
	Green   = color.New(color.FgGreen)
	Yellow  = color.New(color.FgYellow)
	Blue    = color.New(color.FgBlue)
	Cyan    = color.New(color.FgCyan)
	White   = color.New(color.FgWhite, color.Bold)
	Gray    = color.New(color.FgHiBlack)
	Magenta = color.New(color.FgMagenta)
)

// Symbols for output
const (
	SymbolSuccess = "✓"
	SymbolFailed  = "✗"
	SymbolWarning = "⚠"
	SymbolInfo    = "ℹ"
	SymbolArrow   = "→"
	SymbolBullet  = "•"
	SymbolCheck   = "✔"
	SymbolCross   = "✘"
)

// NoColor disables color output
var NoColor = false

// Quiet mode suppresses non-essential output
var Quiet = false

// Verbose mode enables additional output
var Verbose = false

// SetNoColor sets color mode
func SetNoColor(noColor bool) {
	NoColor = noColor
	color.NoColor = noColor
}

// SetQuiet sets quiet mode
func SetQuiet(quiet bool) {
	Quiet = quiet
}

// SetVerbose sets verbose mode
func SetVerbose(verbose bool) {
	Verbose = verbose
}

// Print outputs a message
func Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Println outputs a message with newline
func Println(args ...interface{}) {
	fmt.Println(args...)
}

// Printf outputs a formatted message
func Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	if Quiet {
		return
	}
	Green.Printf("  %s  ", SymbolSuccess)
	fmt.Printf(format+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	Red.Printf("  %s  ", SymbolFailed)
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	if Quiet {
		return
	}
	Yellow.Printf("  %s  ", SymbolWarning)
	fmt.Printf(format+"\n", args...)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	if Quiet {
		return
	}
	Blue.Printf("  %s  ", SymbolInfo)
	fmt.Printf(format+"\n", args...)
}

// Action prints an action message
func Action(format string, args ...interface{}) {
	if Quiet {
		return
	}
	Gray.Printf("  %s  ", SymbolArrow)
	fmt.Printf(format+"\n", args...)
}

// Detail prints a detail message (indented)
func Detail(format string, args ...interface{}) {
	if Quiet {
		return
	}
	Gray.Printf("     %s ", SymbolBullet)
	Gray.Printf(format+"\n", args...)
}

// Debug prints a debug message (only in verbose mode)
func Debug(format string, args ...interface{}) {
	if !Verbose {
		return
	}
	Magenta.Printf("  [DEBUG] ")
	fmt.Printf(format+"\n", args...)
}

// Header prints a section header
func Header(title string) {
	if Quiet {
		return
	}
	fmt.Println()
	Cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	White.Printf("  %s\n", title)
	Cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// Step prints a step indicator with progress
func Step(current, total int, title string) {
	if Quiet {
		return
	}
	percent := current * 100 / total

	// Build progress bar
	barWidth := 20
	filled := current * barWidth / total
	empty := barWidth - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	fmt.Println()
	Cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	White.Printf("  STEP %d/%d", current, total)
	Gray.Printf(" │ ")
	fmt.Printf("%s %d%% ", bar, percent)
	Gray.Printf("│ ")
	White.Printf("%s\n", title)
	Cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// Box prints a message in a box
func Box(title string, lines []string) {
	if Quiet {
		return
	}
	fmt.Println()
	Cyan.Printf("┌─ %s ", title)
	for i := len(title) + 3; i < 70; i++ {
		Cyan.Print("─")
	}
	Cyan.Println("┐")
	Cyan.Println("│")

	for _, line := range lines {
		Cyan.Print("│   ")
		fmt.Println(line)
	}

	Cyan.Println("│")
	Cyan.Print("└")
	for i := 0; i < 70; i++ {
		Cyan.Print("─")
	}
	Cyan.Println("┘")
}

// PrintBanner prints the application banner
func PrintBanner(version string) {
	if Quiet {
		return
	}
	fmt.Println()
	Cyan.Println("╔═══════════════════════════════════════════════════════════════════════════╗")
	Cyan.Println("║                                                                           ║")
	Cyan.Print("║   ")
	White.Print("██╗    ██╗████████╗███████╗")
	Cyan.Println("                                            ║")
	Cyan.Print("║   ")
	White.Print("██║    ██║╚══██╔══╝██╔════╝")
	Cyan.Println("                                            ║")
	Cyan.Print("║   ")
	White.Print("██║ █╗ ██║   ██║   █████╗  ")
	Cyan.Println("                                            ║")
	Cyan.Print("║   ")
	White.Print("██║███╗██║   ██║   ██╔══╝  ")
	Cyan.Println("                                            ║")
	Cyan.Print("║   ")
	White.Print("╚███╔███╔╝   ██║   ███████╗")
	Cyan.Println("                                            ║")
	Cyan.Print("║   ")
	White.Print(" ╚══╝╚══╝    ╚═╝   ╚══════╝")
	Cyan.Println("                                            ║")
	Cyan.Println("║                                                                           ║")
	Cyan.Print("║          ")
	Gray.Printf("WTE (Window to Europe) v%s", version)
	for i := len(version) + 28; i < 65; i++ {
		fmt.Print(" ")
	}
	Cyan.Println("║")
	Cyan.Println("║                                                                           ║")
	Cyan.Println("╚═══════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// PrintCredentialsBox prints credentials in a formatted box
func PrintCredentialsBox(title string, fields map[string]string) {
	fmt.Println()
	Cyan.Printf("┌─ %s ", title)
	for i := len(title) + 3; i < 70; i++ {
		Cyan.Print("─")
	}
	Cyan.Println("┐")
	Cyan.Println("│")

	for key, value := range fields {
		Cyan.Print("│   ")
		fmt.Printf("%-12s ", key+":")
		Green.Println(value)
	}

	Cyan.Println("│")
	Cyan.Print("└")
	for i := 0; i < 70; i++ {
		Cyan.Print("─")
	}
	Cyan.Println("┘")
}

// Fatal prints an error message and exits
func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}

// FatalErr prints an error and exits
func FatalErr(err error) {
	if err != nil {
		Fatal("%v", err)
	}
}

// Confirm asks for user confirmation
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
