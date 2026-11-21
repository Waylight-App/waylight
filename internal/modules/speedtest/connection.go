package speedtest

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/modules/speedtest/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Speedtest Tracker connection (#610). The
// thresholds stay settings of the module; the check reads the results like
// SaveSettings.
func connectionSpecs(source usecase.SpeedtestSource, clock kernel.Clock) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "speedtest", Title: "Speedtest Tracker",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"], 0, 0)
			if err != nil {
				return err
			}
			if _, err := source.Results(ctx, s, clock.Now().Add(-usecase.HistoryRange)); err != nil {
				return kernel.SetupFailure("Speedtest Tracker", err)
			}
			return nil
		},
	}}
}
