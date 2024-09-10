package ui

import (
	"vas/general"

	"fyne.io/fyne"
	"fyne.io/fyne/driver/desktop"
)

func (g *game) Addkeyshortcuts() {
	key_s := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_s, func(shortcut fyne.Shortcut) {
		g.StartMeasurement()
	})
	key_e := desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_e, func(shortcut fyne.Shortcut) {
		g.StopMeasurement()
	})
	key_q := desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_q, func(shortcut fyne.Shortcut) {
		g.closeapp()
	})
	key_u := desktop.CustomShortcut{KeyName: fyne.KeyU, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_u, func(shortcut fyne.Shortcut) {
		general.Checkforupdate()
	})
}
