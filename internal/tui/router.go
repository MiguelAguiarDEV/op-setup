package tui

// Screen identifies a TUI screen.
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDetection
	ScreenAgents
	ScreenComponents
	ScreenReview
	ScreenInstalling
	ScreenComplete
)

// NextScreen returns the next screen in the linear flow.
func NextScreen(s Screen) (Screen, bool) {
	if s >= ScreenComplete {
		return s, false
	}
	return s + 1, true
}

// PreviousScreen returns the previous screen in the linear flow.
func PreviousScreen(s Screen) (Screen, bool) {
	if s <= ScreenWelcome {
		return s, false
	}
	// Can't go back from Installing or Complete.
	if s >= ScreenInstalling {
		return s, false
	}
	return s - 1, true
}
