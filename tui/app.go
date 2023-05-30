package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/data"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type App struct {
	sub *data.Subscription
	ui  *tea.Program
}

func NewApp(p *data.Pipeline) *App {
	sub := p.SubscribeAcceleration("tui")

	model := InitialModel()
	ui := tea.NewProgram(model)

	return &App{
		sub: sub,
		ui:  ui,
	}
}

func (a *App) Run() (err error) {
	go func() {
		lastUpdate := time.Time{}
		var lastAcceleration iim42652.Acceleration
		for {
			select {
			case acceleration := <-a.sub.IncomingAcceleration:
				if lastUpdate == (time.Time{}) {
					lastAcceleration = *acceleration
					lastUpdate = time.Now()
					continue
				}

				timeSinceLastUpdate := time.Since(lastUpdate)

				speed := computeSpeed(timeSinceLastUpdate.Seconds(), lastAcceleration.CamX())
				avgMagnitude := averageMagnitudeForce(lastAcceleration.CamX(), lastAcceleration.CamY())

				lastAcceleration = *acceleration

				motionModel := &MotionModelMsg{
					Acceleration:     acceleration,
					speed:            &speed,
					averageMagnitude: &avgMagnitude,
				}

				a.ui.Send(motionModel)
			}
		}
	}()

	if _, err = a.ui.Run(); err != nil {
		if err != tea.ErrProgramKilled {
			return err
		}
	}
	return nil
}
