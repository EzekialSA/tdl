package prog

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/progress"

	"github.com/iyear/tdl/pkg/utils"
)

// SimpleWriter provides a non-interactive progress logger suitable for containers and non-TTY environments
// Instead of rendering progress bars, it logs important events (start, completion, errors)
type SimpleWriter struct {
	mu           sync.Mutex
	out          io.Writer
	trackers     map[*progress.Tracker]*simpleTracker
	expectedNum  int
	isRendering  bool
	isStopped    bool
	formatter    progress.UnitsFormatter
	overallStart time.Time
	completed    int
	totalBytes   int64
}

type simpleTracker struct {
	startTime time.Time
	total     int64
	message   string
	done      bool
	errored   bool
}

// NewSimple creates a simple non-interactive progress writer
func NewSimple(formatter progress.UnitsFormatter) progress.Writer {
	return &SimpleWriter{
		out:       os.Stdout,
		trackers:  make(map[*progress.Tracker]*simpleTracker),
		formatter: formatter,
	}
}

func (s *SimpleWriter) AppendTracker(tracker *progress.Tracker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.trackers[tracker] = &simpleTracker{
		startTime: time.Now(),
		total:     tracker.Total,
		message:   tracker.Message,
	}
}

func (s *SimpleWriter) AppendTrackers(trackers []*progress.Tracker) {
	for _, t := range trackers {
		s.AppendTracker(t)
	}
}

func (s *SimpleWriter) IsRenderInProgress() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRendering && !s.isStopped
}

func (s *SimpleWriter) Length() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.trackers)
}

func (s *SimpleWriter) LengthActive() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	active := 0
	for _, t := range s.trackers {
		if !t.done && !t.errored {
			active++
		}
	}
	return active
}

func (s *SimpleWriter) LengthDone() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	done := 0
	for _, t := range s.trackers {
		if t.done {
			done++
		}
	}
	return done
}

func (s *SimpleWriter) LengthInQueue() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := 0
	for pt, t := range s.trackers {
		if !t.done && !t.errored && pt.Value() == 0 {
			queue++
		}
	}
	return queue
}

func (s *SimpleWriter) Log(msg string, a ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(a) > 0 {
		fmt.Fprintf(s.out, msg+"\n", a...)
	} else {
		fmt.Fprintln(s.out, msg)
	}
}

func (s *SimpleWriter) Render() {
	s.mu.Lock()
	s.isRendering = true
	s.overallStart = time.Now()
	s.mu.Unlock()

	// Monitor trackers and log completion events
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		s.mu.Lock()
		if s.isStopped {
			s.mu.Unlock()
			return
		}

		// Check for completed/errored trackers
		for pt, st := range s.trackers {
			if !st.done && !st.errored {
				if pt.IsDone() {
					st.done = true
					s.completed++
					s.totalBytes += pt.Total
					elapsed := time.Since(st.startTime)
					speed := ""
					if elapsed.Seconds() > 0 {
						speed = fmt.Sprintf(" at %s/s", s.formatter(int64(float64(pt.Total)/elapsed.Seconds())))
					}
					msg := fmt.Sprintf("%s %s (%s in %s%s)",
						color.GreenString("✓"),
						st.message,
						s.formatter(pt.Total),
						elapsed.Round(time.Millisecond),
						speed)
					fmt.Fprintln(s.out, msg)
				} else if pt.IsErrored() {
					st.errored = true
					elapsed := time.Since(st.startTime)
					msg := fmt.Sprintf("%s %s (failed after %s)",
						color.RedString("✗"),
						st.message,
						elapsed.Round(time.Millisecond))
					fmt.Fprintln(s.out, msg)
				}
			}
		}
		s.mu.Unlock()

		<-ticker.C
	}
}

func (s *SimpleWriter) SetAutoStop(autoStop bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetMessageLength(length int) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetMessageWidth(width int) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetNumTrackersExpected(numTrackers int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expectedNum = numTrackers
	if numTrackers > 0 {
		fmt.Fprintf(s.out, "Starting download of %d file(s)...\n", numTrackers)
	}
}

func (s *SimpleWriter) SetOutputWriter(writer io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.out = writer
}

func (s *SimpleWriter) SetPinnedMessages(messages ...string) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetSortBy(sortBy progress.SortBy) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetStyle(style progress.Style) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetTrackerLength(length int) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetTrackerPosition(position progress.Position) {
	// No-op for simple writer
}

func (s *SimpleWriter) SetUpdateFrequency(frequency time.Duration) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowETA(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowOverall(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowOverallTracker(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowPercentage(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowPinned(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowTime(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowTracker(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) ShowValue(show bool) {
	// No-op for simple writer
}

func (s *SimpleWriter) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStopped {
		return
	}

	s.isStopped = true
	s.isRendering = false

	// Print summary
	elapsed := time.Since(s.overallStart)
	failed := 0
	for _, t := range s.trackers {
		if t.errored {
			failed++
		}
	}

	if s.completed > 0 || failed > 0 {
		summary := fmt.Sprintf("\nDownload complete: %d succeeded", s.completed)
		if failed > 0 {
			summary += fmt.Sprintf(", %d failed", failed)
		}
		summary += fmt.Sprintf(" in %s", elapsed.Round(time.Second))
		if s.completed > 0 && s.totalBytes > 0 {
			summary += fmt.Sprintf(" (%s total, %s/s avg)",
				utils.Byte.FormatBinaryBytes(s.totalBytes),
				utils.Byte.FormatBinaryBytes(int64(float64(s.totalBytes)/elapsed.Seconds())))
		}
		fmt.Fprintln(s.out, summary)
	}
}

func (s *SimpleWriter) Style() *progress.Style {
	return &progress.Style{} // Return empty style for simple writer
}
