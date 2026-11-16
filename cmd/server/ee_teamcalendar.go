//go:build ee

package main

import (
	"context"

	"trisse-home/ee/teamcalendar"
	"trisse-home/ee/teamcalendar/usecase"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/calendar"
	"trisse-home/internal/shared/kernel"
)

// calendarSharingPort adapts the calendar's Sharing to the team calendar.
type calendarSharingPort struct{ s *calendar.Sharing }

func (c calendarSharingPort) OwnSubscriptions(ctx context.Context) ([]usecase.Subscription, error) {
	list, err := c.s.OwnSubscriptions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]usecase.Subscription, 0, len(list))
	for _, x := range list {
		out = append(out, usecase.Subscription{ID: x.ID, Name: x.Name, Visibility: x.Visibility})
	}
	return out, nil
}

func (c calendarSharingPort) Share(ctx context.Context, id string, v kernel.Visibility) error {
	return c.s.Share(ctx, id, v)
}

// Team calendar (#299), part of the ee/ features: calendars others shared
// show in a person's agenda only while the licence covers it.
func init() {
	registerModule("teamcalendar", func(d moduleDeps) modules.Module {
		m := teamcalendar.New(calendarSharingPort{d.calendarSharing}, licenceModule(d).Licensing())
		d.calendarSharing.Enable(m.SharedOn)
		return m
	})
}
