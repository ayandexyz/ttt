package config

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// OmarchyThemeName selects the theme generated from the active Omarchy
// desktop theme. It is not a file, so no user theme can shadow it.
const OmarchyThemeName = "omarchy"

// OverrideOmarchyStateDir replaces the Omarchy state directory, so tests do
// not pick up the theme of the machine they run on.
var OverrideOmarchyStateDir string

// OmarchyStateDir is where Omarchy keeps the active theme. omarchy-theme-set
// replaces its theme/ subdirectory wholesale and then rewrites theme.name, so
// watchers must watch this directory rather than the colors file itself.
func OmarchyStateDir() string {
	if OverrideOmarchyStateDir != "" {
		return OverrideOmarchyStateDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "state", "omarchy", "current")
}

func omarchyColorsPaths() []string {
	if OverrideOmarchyStateDir != "" {
		return []string{filepath.Join(OverrideOmarchyStateDir, "theme", "colors.toml")}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(OmarchyStateDir(), "theme", "colors.toml"),
		// Omarchy 3 kept the active theme under ~/.config.
		filepath.Join(home, ".config", "omarchy", "current", "theme", "colors.toml"),
	}
}

// OmarchyAvailable reports whether an active Omarchy theme can be read.
func OmarchyAvailable() bool {
	_, ok := readOmarchyColors()
	return ok
}

// EffectiveThemeName resolves an unset theme to the Omarchy theme when
// running on Omarchy, so the editor follows the desktop by default. An
// overridden config dir (tests, TTT_CONFIG_DIR scripted sessions) opts out, so
// those runs do not depend on the host's desktop theme.
func EffectiveThemeName(name string) string {
	if name == "" && OverrideConfigDir == "" && OmarchyAvailable() {
		return OmarchyThemeName
	}
	return name
}

func readOmarchyColors() (map[string]string, bool) {
	for _, path := range omarchyColorsPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		colors := parseOmarchyColors(data)
		if colors["background"] != "" || colors["color0"] != "" {
			return colors, true
		}
	}
	return nil, false
}

// parseOmarchyColors reads the flat `key = "value"` subset of TOML that
// colors.toml uses.
func parseOmarchyColors(data []byte) map[string]string {
	colors := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, `"`) {
			if end := strings.Index(value[1:], `"`); end >= 0 {
				value = value[1 : end+1]
			}
		} else if i := strings.Index(value, "#"); i > 0 {
			value = strings.TrimSpace(value[:i])
		}
		colors[strings.TrimSpace(key)] = strings.ToLower(value)
	}
	return colors
}

type omarchyPalette map[string]string

// get returns the first defined key, so semantic names can fall back to the
// legacy color0..color15 names older themes use.
func (p omarchyPalette) get(keys ...string) string {
	for _, k := range keys {
		if v := p[k]; v != "" {
			if _, ok := parseThemeRGB(v); ok {
				return v
			}
		}
	}
	return ""
}

func mixThemeColor(from, to string, amount float64) string {
	a, aOK := parseThemeRGB(from)
	b, bOK := parseThemeRGB(to)
	if !aOK || !bOK {
		return from
	}
	return formatThemeRGB(themeRGB{
		r: a.r + (b.r-a.r)*amount,
		g: a.g + (b.g-a.g)*amount,
		b: a.b + (b.b-a.b)*amount,
	})
}

// LoadOmarchyTheme builds a theme from the active Omarchy colors.toml.
func LoadOmarchyTheme() (ThemeConfig, error) {
	theme := DefaultTheme()
	colors, ok := readOmarchyColors()
	if !ok {
		return theme, os.ErrNotExist
	}
	applyOmarchyPalette(&theme, omarchyPalette(colors))
	theme.ResolveColors()
	return theme, nil
}

func applyOmarchyPalette(t *ThemeConfig, p omarchyPalette) {
	bg := p.get("background", "color0")
	fg := p.get("foreground", "color7", "color15")
	if bg == "" || fg == "" {
		return
	}
	accent := p.get("accent", "blue", "color4")
	muted := p.get("muted", "dark_foreground", "color8")
	if muted == "" {
		muted = mixThemeColor(fg, bg, 0.45)
	}
	selection := p.get("selection", "selection_background")
	if selection == "" {
		selection = mixThemeColor(bg, fg, 0.2)
	}
	lighter := p.get("lighter_background")
	if lighter == "" {
		lighter = mixThemeColor(bg, fg, 0.08)
	}
	darker := p.get("dark_background")
	if darker == "" {
		darker = mixThemeColor(bg, "#000000", 0.2)
	}

	red := p.get("red", "color1")
	green := p.get("green", "color2")
	yellow := p.get("yellow", "color3")
	blue := p.get("blue", "color4")
	magenta := p.get("magenta", "purple", "color5")
	cyan := p.get("cyan", "color6")
	orange := p.get("orange", "bright_yellow", "color11")
	if orange == "" {
		orange = yellow
	}

	t.Terminal = TerminalColors{
		Foreground:    fg,
		Background:    bg,
		Black:         p.get("color0", "background"),
		Red:           red,
		Green:         green,
		Yellow:        yellow,
		Blue:          blue,
		Magenta:       magenta,
		Cyan:          cyan,
		White:         p.get("color7", "foreground"),
		BrightBlack:   p.get("color8", "muted", "dark_foreground"),
		BrightRed:     p.get("bright_red", "color9", "red"),
		BrightGreen:   p.get("bright_green", "color10", "green"),
		BrightYellow:  p.get("bright_yellow", "color11", "yellow"),
		BrightBlue:    p.get("bright_blue", "color12", "blue"),
		BrightMagenta: p.get("bright_magenta", "bright_purple", "color13", "magenta"),
		BrightCyan:    p.get("bright_cyan", "color14", "cyan"),
		BrightWhite:   p.get("bright_foreground", "color15", "foreground"),
		Selection:     selection,
	}

	t.Default = StyleDef{Fg: fg, Bg: bg}
	t.Muted = StyleDef{Fg: muted}
	t.Success = StyleDef{Fg: green}
	t.Danger = StyleDef{Fg: red}
	t.Warning = StyleDef{Fg: yellow}
	t.Conflict = StyleDef{Fg: magenta}
	t.StatusBar = StyleDef{Fg: fg}
	t.Border = StyleDef{Fg: mixThemeColor(muted, bg, 0.3)}
	t.BorderActive = StyleDef{Fg: accent}

	t.Tabs = TabStyles{
		Active:   StyleDef{Fg: accent, Bold: true},
		Inactive: StyleDef{Fg: muted},
	}
	t.Sidebar = SidebarStyles{
		Header:   StyleDef{Fg: accent, Bold: true},
		Item:     StyleDef{Fg: fg},
		Selected: StyleDef{Fg: fg, Bg: selection},
	}
	t.Dialog = DialogStyles{
		Item:     StyleDef{Fg: fg},
		Selected: StyleDef{Fg: fg, Bg: selection},
		Muted:    StyleDef{Fg: muted},
	}
	t.Menu = MenuStyles{
		Item:   StyleDef{Fg: fg},
		Active: StyleDef{Fg: bg, Bg: accent, Bold: true},
	}
	t.Button = ButtonStyles{
		Focused: StyleDef{Fg: bg, Bg: accent},
	}
	t.Input = InputStyles{
		Item: StyleDef{Fg: fg, Bg: darker},
	}
	t.Scrollbar = StyleDef{Fg: muted, Bg: lighter}

	t.Editor = EditorStyles{
		LineNumber:    StyleDef{Fg: muted},
		ActiveLine:    StyleDef{Bg: lighter},
		Selection:     StyleDef{Bg: selection},
		SearchMatch:   StyleDef{Fg: bg, Bg: yellow},
		SearchActive:  StyleDef{Fg: bg, Bg: orange},
		BracketMatch:  StyleDef{Bg: selection},
		BracketColors: []string{"yellow", "magenta", "blue"},
		Diagnostics: DiagnosticStyles{
			Error:   StyleDef{Fg: red},
			Warning: StyleDef{Fg: yellow},
			Info:    StyleDef{Fg: cyan},
			Hint:    StyleDef{Fg: muted},
		},
	}

	t.Diff = DiffStyles{
		Added:             StyleDef{Bg: mixThemeColor(bg, green, 0.15)},
		Deleted:           StyleDef{Bg: mixThemeColor(bg, red, 0.15)},
		Modified:          StyleDef{Bg: mixThemeColor(bg, yellow, 0.15)},
		CollapsedEmphasis: StyleDef{Bold: true},
		GutterAdded:       StyleDef{Fg: green},
		GutterDeleted:     StyleDef{Fg: red},
		GutterModified:    StyleDef{Fg: yellow},
		GutterBookmark:    StyleDef{Fg: accent},
	}

	t.Syntax = SyntaxStyles{
		Comment:     StyleDef{Fg: muted, Italic: true},
		String:      StyleDef{Fg: green},
		Keyword:     StyleDef{Fg: magenta},
		Number:      StyleDef{Fg: orange},
		Operator:    StyleDef{Fg: cyan},
		Function:    StyleDef{Fg: blue},
		Type:        StyleDef{Fg: yellow},
		Builtin:     StyleDef{Fg: cyan},
		Variable:    StyleDef{Fg: fg},
		Punctuation: StyleDef{Fg: p.get("light_foreground", "foreground")},
		Tag:         StyleDef{Fg: red},
		Attribute:   StyleDef{Fg: yellow},
		Constant:    StyleDef{Fg: orange},
		Heading:     StyleDef{Fg: accent, Bold: true},
		Link:        StyleDef{Fg: cyan},
	}
	t.TokenColors = nil
}
