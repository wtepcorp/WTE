package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps schollz/progressbar
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar
func NewProgressBar(max int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	if !Quiet {
		bar = progressbar.NewOptions64(
			max,
			progressbar.OptionSetDescription(description),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(30),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	return &ProgressBar{bar: bar}
}

// NewSpinner creates a new spinner
func NewSpinner(description string) *ProgressBar {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
	)

	if !Quiet {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
		)
	}

	return &ProgressBar{bar: bar}
}

// Add adds to the progress bar
func (p *ProgressBar) Add(n int) {
	_ = p.bar.Add(n)
}

// Add64 adds to the progress bar (int64)
func (p *ProgressBar) Add64(n int64) {
	_ = p.bar.Add64(n)
}

// Set sets the current progress
func (p *ProgressBar) Set(n int) {
	_ = p.bar.Set(n)
}

// Set64 sets the current progress (int64)
func (p *ProgressBar) Set64(n int64) {
	_ = p.bar.Set64(n)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	_ = p.bar.Finish()
}

// Clear clears the progress bar
func (p *ProgressBar) Clear() {
	_ = p.bar.Clear()
}

// Describe updates the description
func (p *ProgressBar) Describe(description string) {
	p.bar.Describe(description)
}

// Writer returns an io.Writer for the progress bar
func (p *ProgressBar) Writer() io.Writer {
	return p.bar
}

// DownloadProgressBar creates a progress bar suitable for downloads
func DownloadProgressBar(size int64, filename string) *ProgressBar {
	description := fmt.Sprintf("Downloading %s", filename)
	return NewProgressBar(size, description)
}
