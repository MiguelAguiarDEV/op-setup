package tui

import "testing"

func TestNextScreen(t *testing.T) {
	tests := []struct {
		from Screen
		want Screen
		ok   bool
	}{
		{ScreenWelcome, ScreenProfile, true},
		{ScreenProfile, ScreenDetection, true},
		{ScreenDetection, ScreenAgents, true},
		{ScreenAgents, ScreenComponents, true},
		{ScreenComponents, ScreenReview, true},
		{ScreenReview, ScreenInstalling, true},
		{ScreenInstalling, ScreenComplete, true},
		{ScreenComplete, ScreenComplete, false},
	}

	for _, tt := range tests {
		got, ok := NextScreen(tt.from)
		if got != tt.want || ok != tt.ok {
			t.Errorf("NextScreen(%d) = (%d, %v), want (%d, %v)", tt.from, got, ok, tt.want, tt.ok)
		}
	}
}

func TestPreviousScreen(t *testing.T) {
	tests := []struct {
		from Screen
		want Screen
		ok   bool
	}{
		{ScreenWelcome, ScreenWelcome, false},
		{ScreenProfile, ScreenWelcome, true},
		{ScreenDetection, ScreenProfile, true},
		{ScreenAgents, ScreenDetection, true},
		{ScreenComponents, ScreenAgents, true},
		{ScreenReview, ScreenComponents, true},
		{ScreenInstalling, ScreenInstalling, false}, // Can't go back
		{ScreenComplete, ScreenComplete, false},     // Can't go back
	}

	for _, tt := range tests {
		got, ok := PreviousScreen(tt.from)
		if got != tt.want || ok != tt.ok {
			t.Errorf("PreviousScreen(%d) = (%d, %v), want (%d, %v)", tt.from, got, ok, tt.want, tt.ok)
		}
	}
}
