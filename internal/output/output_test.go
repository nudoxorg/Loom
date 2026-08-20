package output

import (
	"fmt"
	"testing"
	"time"

	"github.com/nudoxorg/loom/internal/project"
	"github.com/nudoxorg/loom/internal/storage"
)

func TestFormatEventTodayUsesBareTime(t *testing.T) {
	e := storage.Event{
		Agent:     "claude-code",
		Kind:      "log",
		Message:   "picking postgres over mysql",
		Timestamp: time.Now(),
	}

	got := formatEvent(e, false)
	want := formatTime(e.Timestamp) + "  log      claude-code  picking postgres over mysql"
	if got != want {
		t.Fatalf("formatEvent() = %q, want %q", got, want)
	}
}

func TestFormatEventOlderIncludesDate(t *testing.T) {
	old := time.Now().AddDate(0, 0, -3)
	e := storage.Event{Agent: "manual", Kind: "claim", Message: "claimed internal/auth", Timestamp: old}

	got := formatEvent(e, false)
	if got != old.Format("Jan _2 15:04")+"  claim    manual  claimed internal/auth" {
		t.Fatalf("formatEvent() for older event = %q", got)
	}
}

func TestFormatEventsEmpty(t *testing.T) {
	if got := formatEvents(nil, false); got != "no events" {
		t.Fatalf("formatEvents(nil) = %q, want %q", got, "no events")
	}
}

func TestFormatClaimShape(t *testing.T) {
	c := storage.Claim{Agent: "cursor", Path: "internal/auth", ClaimedAt: time.Now()}

	got := formatClaim(c, false)
	want := formatTime(c.ClaimedAt) + "  cursor  internal/auth"
	if got != want {
		t.Fatalf("formatClaim() = %q, want %q", got, want)
	}
}

func TestFormatClaimsEmpty(t *testing.T) {
	if got := formatClaims(nil, false); got != "no active claims" {
		t.Fatalf("formatClaims(nil) = %q, want %q", got, "no active claims")
	}
}

func TestFormatConfigEntry(t *testing.T) {
	if got := formatConfigEntry("default_agent", "manual", false); got != "default_agent = manual" {
		t.Fatalf("formatConfigEntry() = %q", got)
	}
}

func TestFormatSuccess(t *testing.T) {
	if got := formatSuccess("logged", false); got != "✓ logged" {
		t.Fatalf("formatSuccess() = %q", got)
	}
}

func TestPaintDisabledReturnsPlainText(t *testing.T) {
	if got := paint(colorCyan, "log", false); got != "log" {
		t.Fatalf("paint(color=false) = %q, want plain text unchanged", got)
	}
}

func TestPaintEnabledWrapsWithColorCodes(t *testing.T) {
	got := paint(colorCyan, "log", true)
	if got != colorCyan+"log"+colorReset {
		t.Fatalf("paint(color=true) = %q, want wrapped in ANSI codes", got)
	}
}

func TestFormatBold(t *testing.T) {
	if got := formatBold("RECENT EVENTS", false); got != "RECENT EVENTS" {
		t.Fatalf("formatBold(color=false) = %q, want plain text unchanged", got)
	}
	if got := formatBold("RECENT EVENTS", true); got != colorBold+"RECENT EVENTS"+colorReset {
		t.Fatalf("formatBold(color=true) = %q, want wrapped in bold", got)
	}
}

func TestFormatGlobalEventsEmpty(t *testing.T) {
	if got := formatGlobalEvents(nil, false); got != "no events" {
		t.Fatalf("formatGlobalEvents(nil) = %q, want %q", got, "no events")
	}
}

func TestFormatGlobalClaimsEmpty(t *testing.T) {
	if got := formatGlobalClaims(nil, false); got != "no active claims" {
		t.Fatalf("formatGlobalClaims(nil) = %q, want %q", got, "no active claims")
	}
}

func TestFormatGlobalEventsPadsToLongestProjectLabel(t *testing.T) {
	now := time.Now()
	events := []project.LabeledEvent{
		{Project: "loom", Event: storage.Event{Kind: "log", Agent: "claude-code", Message: "short label", Timestamp: now}},
		{Project: "api-service", Event: storage.Event{Kind: "claim", Agent: "cursor", Message: "long label", Timestamp: now}},
	}

	got := formatGlobalEvents(events, false)

	// width should be len("api-service") == 11, applied to every row.
	want := fmt.Sprintf("%-11s  %s  %-7s  %s  %s", "loom", formatTime(now), "log", "claude-code", "short label") +
		"\n" +
		fmt.Sprintf("%-11s  %s  %-7s  %s  %s", "api-service", formatTime(now), "claim", "cursor", "long label")
	if got != want {
		t.Fatalf("formatGlobalEvents() =\n%q\nwant\n%q", got, want)
	}
}

func TestFormatGlobalClaimsIncludesProjectLabel(t *testing.T) {
	now := time.Now()
	claims := []project.LabeledClaim{
		{Project: "loom", Claim: storage.Claim{Agent: "claude-code", Path: "internal/mcpserver", ClaimedAt: now}},
	}

	got := formatGlobalClaims(claims, false)
	want := fmt.Sprintf("%-4s  %s  %s  %s", "loom", formatTime(now), "claude-code", "internal/mcpserver")
	if got != want {
		t.Fatalf("formatGlobalClaims() = %q, want %q", got, want)
	}
}
