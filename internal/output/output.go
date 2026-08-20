// Package output renders Loom's data for a human terminal: aligned
// columns and a subtle color accent per event kind. MCP tool results
// are a separate, deliberately plain-text path — an LLM reads those,
// not a terminal, so this package is CLI-only.
package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
)

// ColorEnabled is computed once at startup from whether stdout is a
// terminal and whether NO_COLOR is set (https://no-color.org). It's a
// var, not a const, so tests can override it.
var ColorEnabled = isatty.IsTerminal(os.Stdout.Fd()) && os.Getenv("NO_COLOR") == ""

const (
	colorDim    = "\x1b[2m"
	colorBold   = "\x1b[1m"
	colorCyan   = "\x1b[36m"
	colorYellow = "\x1b[33m"
	colorGreen  = "\x1b[32m"
	colorBlue   = "\x1b[34m"
	colorReset  = "\x1b[0m"
)

func paint(code, text string, color bool) string {
	if !color || text == "" {
		return text
	}
	return code + text + colorReset
}

func kindColor(kind string) string {
	switch kind {
	case "claim":
		return colorYellow
	case "release":
		return colorGreen
	default:
		return colorCyan
	}
}

// formatTime renders bare HH:MM for events from today, and a dated
// form for anything older, so a multi-day `loom show` stays unambiguous.
func formatTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}
	return t.Format("Jan _2 15:04")
}

func formatEvent(e storage.Event, color bool) string {
	kind := fmt.Sprintf("%-7s", e.Kind)
	return fmt.Sprintf(
		"%s  %s  %s  %s",
		paint(colorDim, formatTime(e.Timestamp), color),
		paint(kindColor(e.Kind), kind, color),
		paint(colorBold, e.Agent, color),
		e.Message,
	)
}

// Event renders a single event as a colored, column-aligned line.
func Event(e storage.Event) string {
	return formatEvent(e, ColorEnabled)
}

func formatEvents(events []storage.Event, color bool) string {
	if len(events) == 0 {
		return paint(colorDim, "no events", color)
	}

	lines := make([]string, len(events))
	for i, e := range events {
		lines[i] = formatEvent(e, color)
	}
	return strings.Join(lines, "\n")
}

// Events renders a list of events, most recent first, one per line.
func Events(events []storage.Event) string {
	return formatEvents(events, ColorEnabled)
}

func formatClaim(c storage.Claim, color bool) string {
	return fmt.Sprintf(
		"%s  %s  %s",
		paint(colorDim, formatTime(c.ClaimedAt), color),
		paint(colorBold, c.Agent, color),
		c.Path,
	)
}

// Claim renders a single active claim as a colored, column-aligned line.
func Claim(c storage.Claim) string {
	return formatClaim(c, ColorEnabled)
}

func formatClaims(claims []storage.Claim, color bool) string {
	if len(claims) == 0 {
		return paint(colorDim, "no active claims", color)
	}

	lines := make([]string, len(claims))
	for i, c := range claims {
		lines[i] = formatClaim(c, color)
	}
	return strings.Join(lines, "\n")
}

// Claims renders a list of active claims, one per line.
func Claims(claims []storage.Claim) string {
	return formatClaims(claims, ColorEnabled)
}

func formatConfigEntry(key, value string, color bool) string {
	return fmt.Sprintf("%s = %s", paint(colorDim, key, color), value)
}

// ConfigEntry renders a single "key = value" line for `loom config list`.
func ConfigEntry(key, value string) string {
	return formatConfigEntry(key, value, ColorEnabled)
}

func formatSuccess(message string, color bool) string {
	return fmt.Sprintf("%s %s", paint(colorBold+colorGreen, "✓", color), message)
}

// Success renders a short confirmation line for a write command.
func Success(message string) string {
	return formatSuccess(message, ColorEnabled)
}

func formatBold(text string, color bool) string {
	return paint(colorBold, text, color)
}

// Bold renders text bold, for section headers in multi-part output.
func Bold(text string) string {
	return formatBold(text, ColorEnabled)
}

func projectColumnWidth[T any](rows []T, project func(T) string) int {
	width := 0
	for _, r := range rows {
		if n := len(project(r)); n > width {
			width = n
		}
	}
	return width
}

func formatGlobalEvent(le project.LabeledEvent, projectWidth int, color bool) string {
	label := fmt.Sprintf("%-*s", projectWidth, le.Project)
	kind := fmt.Sprintf("%-7s", le.Event.Kind)
	return fmt.Sprintf(
		"%s  %s  %s  %s  %s",
		paint(colorBlue, label, color),
		paint(colorDim, formatTime(le.Event.Timestamp), color),
		paint(kindColor(le.Event.Kind), kind, color),
		paint(colorBold, le.Event.Agent, color),
		le.Event.Message,
	)
}

func formatGlobalEvents(events []project.LabeledEvent, color bool) string {
	if len(events) == 0 {
		return paint(colorDim, "no events", color)
	}

	width := projectColumnWidth(events, func(e project.LabeledEvent) string { return e.Project })

	lines := make([]string, len(events))
	for i, e := range events {
		lines[i] = formatGlobalEvent(e, width, color)
	}
	return strings.Join(lines, "\n")
}

// GlobalEvents renders events from more than one project, most recent
// first, each line prefixed with which project it came from.
func GlobalEvents(events []project.LabeledEvent) string {
	return formatGlobalEvents(events, ColorEnabled)
}

func formatGlobalClaim(lc project.LabeledClaim, projectWidth int, color bool) string {
	label := fmt.Sprintf("%-*s", projectWidth, lc.Project)
	return fmt.Sprintf(
		"%s  %s  %s  %s",
		paint(colorBlue, label, color),
		paint(colorDim, formatTime(lc.Claim.ClaimedAt), color),
		paint(colorBold, lc.Claim.Agent, color),
		lc.Claim.Path,
	)
}

func formatGlobalClaims(claims []project.LabeledClaim, color bool) string {
	if len(claims) == 0 {
		return paint(colorDim, "no active claims", color)
	}

	width := projectColumnWidth(claims, func(c project.LabeledClaim) string { return c.Project })

	lines := make([]string, len(claims))
	for i, c := range claims {
		lines[i] = formatGlobalClaim(c, width, color)
	}
	return strings.Join(lines, "\n")
}

// GlobalClaims renders active claims from more than one project, each
// line prefixed with which project it came from.
func GlobalClaims(claims []project.LabeledClaim) string {
	return formatGlobalClaims(claims, ColorEnabled)
}
