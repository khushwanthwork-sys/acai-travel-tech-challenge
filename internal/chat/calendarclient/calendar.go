package calendarclient

import (
	"context"
	"fmt"
	"log/slog"

	ics "github.com/arran4/golang-ical"
)

// LoadCalendar fetches and parses an iCalendar from a URL
func LoadCalendar(ctx context.Context, link string) ([]*ics.VEvent, error) {
	slog.InfoContext(ctx, "Loading calendar", "link", link)

	cal, err := ics.ParseCalendarFromUrl(link, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return cal.Events(), nil
}
