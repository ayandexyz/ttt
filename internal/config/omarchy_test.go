package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeOmarchyColors(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "theme")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "colors.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	OverrideOmarchyStateDir = dir
	t.Cleanup(func() { OverrideOmarchyStateDir = "" })
}

func TestLoadOmarchyThemeMapsPalette(t *testing.T) {
	writeOmarchyColors(t, `mode = "dark"
accent = "#F38D70" # inline comment
background = "#2c2525"
foreground = "#e6d9db"
muted = "#72696a"
selection = "#403e41"
red = "#fd6883"
green = "#adda78"
bright_red = "#ff8297"
`)

	if !OmarchyAvailable() {
		t.Fatal("expected Omarchy theme to be available")
	}
	theme, err := LoadTheme(OmarchyThemeName)
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string][2]string{
		"default.fg":      {theme.Default.Fg, "#e6d9db"},
		"default.bg":      {theme.Default.Bg, "#2c2525"},
		"tabs.active":     {theme.Tabs.Active.Fg, "#f38d70"},
		"selection":       {theme.Editor.Selection.Bg, "#403e41"},
		"danger":          {theme.Danger.Fg, "#fd6883"},
		"terminal.bright": {theme.Terminal.BrightRed, "#ff8297"},
		"syntax.string":   {theme.Syntax.String.Fg, "#adda78"},
		"muted":           {theme.Muted.Fg, "#72696a"},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %q, want %q", name, c[0], c[1])
		}
	}
}

func TestLoadOmarchyThemeLegacyColorNames(t *testing.T) {
	writeOmarchyColors(t, `color0 = "#1a1b26"
color1 = "#f7768e"
color7 = "#c0caf5"
`)

	theme, err := LoadOmarchyTheme()
	if err != nil {
		t.Fatal(err)
	}
	if theme.Default.Bg != "#1a1b26" || theme.Default.Fg != "#c0caf5" || theme.Danger.Fg != "#f7768e" {
		t.Errorf("legacy names not mapped: default=%+v danger=%q", theme.Default, theme.Danger.Fg)
	}
}

func TestEffectiveThemeName(t *testing.T) {
	writeOmarchyColors(t, `background = "#000000"
foreground = "#ffffff"
`)
	if got := EffectiveThemeName(""); got != OmarchyThemeName {
		t.Errorf("unset theme on Omarchy = %q, want %q", got, OmarchyThemeName)
	}
	if got := EffectiveThemeName("dracula"); got != "dracula" {
		t.Errorf("explicit theme = %q, want dracula", got)
	}

	OverrideConfigDir = t.TempDir()
	t.Cleanup(func() { OverrideConfigDir = "" })
	if got := EffectiveThemeName(""); got != "" {
		t.Errorf("unset theme with overridden config dir = %q, want empty", got)
	}
}

func TestOmarchyUnavailable(t *testing.T) {
	OverrideOmarchyStateDir = t.TempDir()
	t.Cleanup(func() { OverrideOmarchyStateDir = "" })

	if OmarchyAvailable() {
		t.Fatal("expected no Omarchy theme")
	}
	if got := EffectiveThemeName(""); got != "" {
		t.Errorf("unset theme off Omarchy = %q, want empty", got)
	}
	for _, name := range ListThemes() {
		if name == OmarchyThemeName {
			t.Error("omarchy listed without an Omarchy theme")
		}
	}
}
