package main

import (
	ui "github.com/gizak/termui"
	"github.com/jimmidyson/wurzel/node"
)

func main() {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	header := ui.NewPar(":PRESS q TO QUIT")
	header.Height = 3
	header.Width = 50
	header.TextFgColor = ui.ColorWhite
	header.BorderFg = ui.ColorCyan

	memGauge := ui.NewGauge()
	memGauge.BorderLabel = "Memory"
	memGauge.Height = 3
	memGauge.BorderFg = ui.ColorWhite
	memGauge.BorderLabelFg = ui.ColorCyan

	swapGauge := ui.NewGauge()
	swapGauge.BorderLabel = "Swap"
	swapGauge.Height = 3
	swapGauge.BorderFg = ui.ColorWhite
	swapGauge.BorderLabelFg = ui.ColorCyan

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, header),
		),
		ui.NewRow(
			ui.NewCol(6, 0, memGauge),
			ui.NewCol(6, 0, swapGauge),
		),
	)

	ui.Body.Align()
	ui.Render(ui.Body)

	update := func() {
		mem, _ := node.Memory()
		memGauge.Percent = int(mem.UsedPercent)
		b, f := thresholdColour(70, 90, int(mem.UsedPercent))
		memGauge.BarColor = b
		memGauge.PercentColor = f

		swap, _ := node.Swap()
		swapGauge.Percent = int(swap.UsedPercent)
		b, f = thresholdColour(70, 90, int(swap.UsedPercent))
		swapGauge.BarColor = b
		swapGauge.PercentColor = f

		ui.Render(ui.Body)
	}

	update()

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Handle("/timer/1s", func(e ui.Event) {
		update()
	})

	ui.Loop()
}

func thresholdColour(warningLevel, errorLevel, actualLevel int) (ui.Attribute, ui.Attribute) {
	if actualLevel < warningLevel {
		return ui.ColorGreen, ui.ColorWhite
	}
	if actualLevel < errorLevel {
		return ui.ColorYellow, ui.ColorWhite
	}
	return ui.ColorRed, ui.ColorWhite
}
