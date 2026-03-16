package xdg

import "testing"

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got := ConfigDir("/home/user")
	want := "/home/user/.config"
	if got != want {
		t.Fatalf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_XDGSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := ConfigDir("/home/user")
	want := "/custom/config"
	if got != want {
		t.Fatalf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_XDGOverridesHomeDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/opt/myconfig")
	got := ConfigDir("/home/different")
	if got != "/opt/myconfig" {
		t.Fatalf("ConfigDir() = %q, want %q (XDG should override homeDir)", got, "/opt/myconfig")
	}
}
