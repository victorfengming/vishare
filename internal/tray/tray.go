package tray

import (
	"fyne.io/systray"
	"github.com/rs/zerolog/log"
	"github.com/victorfengming/vishare/internal/status"
)

// Run starts the systray. Must be called from the main goroutine.
// iconConnected and iconDisconnected are PNG icon bytes.
func Run(statusCh <-chan status.Msg, iconConnected, iconDisconnected []byte, quit func()) {
	systray.Run(func() { onReady(statusCh, iconConnected, iconDisconnected, quit) }, onExit)
}

func onReady(statusCh <-chan status.Msg, iconConnected, iconDisconnected []byte, quit func()) {
	systray.SetIcon(iconDisconnected)
	systray.SetTitle("ViShare")
	systray.SetTooltip("ViShare — disconnected")

	menuStatus := systray.AddMenuItem("Disconnected", "Connection status")
	menuStatus.Disable()
	systray.AddSeparator()
	menuQuit := systray.AddMenuItem("Quit", "Quit ViShare")

	go func() {
		for {
			select {
			case msg, ok := <-statusCh:
				if !ok {
					return
				}
				if msg.Connected {
					systray.SetIcon(iconConnected)
					label := "Connected"
					if msg.ClientName != "" {
						label = "Connected — " + msg.ClientName
					}
					menuStatus.SetTitle(label)
					systray.SetTooltip("ViShare — " + label)
				} else {
					systray.SetIcon(iconDisconnected)
					menuStatus.SetTitle("Disconnected")
					systray.SetTooltip("ViShare — disconnected")
				}
			case <-menuQuit.ClickedCh:
				log.Info().Msg("quit via tray")
				quit()
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {}
