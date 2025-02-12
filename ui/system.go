package ui

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mcuadros/go-octoprint"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

var systemPanelInstance *systemPanel

type systemPanel struct {
	CommonPanel

	list *gtk.Box
}

func SystemPanel(ui *UI, parent Panel) *systemPanel {
	if systemPanelInstance == nil {
		m := &systemPanel{CommonPanel: NewCommonPanel(ui, parent)}
		m.initialize()
		systemPanelInstance = m
	} else {
		systemPanelInstance.p = parent
	}

	return systemPanelInstance
}

func (m *systemPanel) initialize() {
	defer m.Initialize()

	m.Grid().Attach(m.createOctoPrintInfo(), 1, 0, 2, 1)
	m.Grid().Attach(m.createOctoScreenInfo(), 3, 0, 2, 1)
	m.Grid().Attach(m.createSystemInfo(), 1, 1, 3, 1)
}

func (m *systemPanel) createOctoPrintInfo() *gtk.Box {
	r, err := (&octoprint.VersionRequest{}).Do(m.UI.Printer)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	info := MustBox(gtk.ORIENTATION_VERTICAL, 0)

	info.SetHExpand(true)
	info.SetHAlign(gtk.ALIGN_CENTER)
	info.SetVExpand(true)
	info.SetVAlign(gtk.ALIGN_CENTER)
	logoWidth := m.Scaled(69)
	img := MustImageFromFileWithSize("logo-octoprint.png", logoWidth, int(float64(logoWidth)*1.25))
	info.Add(img)

	info.Add(MustLabel("\nOctoPrint Version: <b>%s (%s)</b>", r.Server, r.API))
	return info
}

func (m *systemPanel) createOctoScreenInfo() *gtk.Box {
	info := MustBox(gtk.ORIENTATION_VERTICAL, 0)

	info.SetHExpand(true)
	info.SetHAlign(gtk.ALIGN_CENTER)
	info.SetVExpand(true)
	info.SetVAlign(gtk.ALIGN_CENTER)

	logoWidth := m.Scaled(80)

	img := MustImageFromFileWithSize("logo-z-bolt.svg", logoWidth, int(float64(logoWidth)*0.8))
	info.Add(img)
	info.Add(MustLabel("UI Version: <b>%s (%s)</b>", Version, Build))
	return info
}

func (m *systemPanel) createSystemInfo() *gtk.Box {
	info := MustBox(gtk.ORIENTATION_VERTICAL, 0)

	info.SetVExpand(true)
	info.SetVAlign(gtk.ALIGN_CENTER)

	title := MustLabel("<b>System Information</b>")
	title.SetMarginBottom(5)
	title.SetMarginTop(15)
	info.Add(title)

	v, _ := mem.VirtualMemory()
	info.Add(MustLabel(fmt.Sprintf(
		"Memory Total / Free: <b>%s / %s</b>",
		humanize.Bytes(v.Total), humanize.Bytes(v.Free),
	)))

	l, _ := load.Avg()
	info.Add(MustLabel(fmt.Sprintf(
		"Load Average: <b>%.2f, %.2f, %.2f</b>",
		l.Load1, l.Load5, l.Load15,
	)))

	return info
}

func (m *systemPanel) createActionBar() gtk.IWidget {
	bar := MustBox(gtk.ORIENTATION_HORIZONTAL, 5)
	bar.SetHAlign(gtk.ALIGN_END)
	bar.SetHExpand(true)
	bar.SetMarginTop(5)
	bar.SetMarginBottom(5)
	bar.SetMarginEnd(5)

	if b := m.createRestartButton(); b != nil {
		bar.Add(b)
	}

	bar.Add(MustButton(MustImageFromFileWithSize("back.svg", m.Scaled(40), m.Scaled(40)), m.UI.GoHistory))

	return bar
}

func (m *systemPanel) createRestartButton() gtk.IWidget {
	r, err := (&octoprint.SystemCommandsRequest{}).Do(m.UI.Printer)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	var cmd *octoprint.CommandDefinition
	for _, c := range r.Core {
		if c.Action == "reboot" {
			cmd = c
		}
	}

	if cmd == nil {
		return nil
	}

	return m.doCreateButtonFromCommand(cmd)
}

func (m *systemPanel) doCreateButtonFromCommand(cmd *octoprint.CommandDefinition) gtk.IWidget {
	do := func() {
		r := &octoprint.SystemExecuteCommandRequest{
			Source: octoprint.Core,
			Action: cmd.Action,
		}

		if err := r.Do(m.UI.Printer); err != nil {
			Logger.Error(err)
			return
		}
	}

	cb := do
	if len(cmd.Confirm) != 0 {
		cb = MustConfirmDialog(m.UI.w, cmd.Confirm, do)
	}

	return MustButton(MustImageFromFileWithSize(cmd.Action+".svg", 40, 40), cb)
}
