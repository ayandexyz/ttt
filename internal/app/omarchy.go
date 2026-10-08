package app

import (
	"github.com/eugenioenko/ttt/internal/config"
	"github.com/eugenioenko/ttt/internal/watcher"

	"github.com/gdamore/tcell/v3"
)

// OmarchyThemeChangedResult is posted to the event loop when the active
// Omarchy desktop theme changes.
type OmarchyThemeChangedResult struct{}

// StartOmarchyWatcher follows Omarchy theme switches. omarchy-theme-set swaps
// the whole theme/ directory by rename, so the state directory's listing is
// what changes, not colors.toml in place.
func (a *App) StartOmarchyWatcher() {
	if !config.OmarchyAvailable() {
		return
	}
	dir := config.OmarchyStateDir()
	w, err := watcher.New(nil, func(string) {
		if a.Screen != nil {
			a.Screen.PostEvent(tcell.NewEventInterrupt(&OmarchyThemeChangedResult{}))
		}
	})
	if err != nil {
		return
	}
	w.SyncDirs([]string{dir})
	a.OmarchyWatcher = w
}

func (a *App) HandleOmarchyThemeChanged() {
	if a.Screen == nil || config.EffectiveThemeName(a.Settings.Theme) != config.OmarchyThemeName {
		return
	}
	theme, err := config.LoadOmarchyTheme()
	if err != nil {
		return
	}
	borders := a.applyThemeConfig(theme, a.Settings.Editor.TransparentBackground)
	a.applyBorderStyle(&borders)
}
