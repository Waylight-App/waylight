# Team calendar

With a licence key that includes `absences` ([licence keys](licence-keys.md)),
which covers absences and the team calendar, members and admins share
their calendar subscriptions with a group or with everyone. This needs a
server built with the `ee` features.

## Share a calendar

A calendar subscription (an ICS address under the calendar settings) can be
shared like a note: only you, everyone, or one of your groups (admins can
pick any group). Everyone it is shared with sees its events in their
calendar widget, with the calendar's name and colour. They cannot change it
and never see its address, which may contain a private token.

A calendar connected through a Google or Microsoft account can be shared
the same way: sharing the account shares the calendars you enabled for your
agenda. Its events are read with your connection; when its token needs
renewing, it is renewed and stored as yours, whoever opened the calendar,
so your connection keeps working.

## What others see

- Shared calendars are added to each person's agenda when it is shown, with
  that person's groups at that moment: leaving the group or a removed key
  hides them at once.
- Guests and kiosk displays keep seeing the owner's calendars, nothing
  shared.
- The feed is fetched once per refresh interval of the subscription, however
  many people see it.

## API

| Method | Path | |
| --- | --- | --- |
| GET | `/api/teamcalendar/subscriptions` | Your own subscriptions with `id`, `name`, `visibility` |
| PUT | `/api/teamcalendar/subscriptions/{id}` | `{visibility}`: `private`, `everyone` or `group:<id>` |
| GET | `/api/teamcalendar/accounts` | Your own connected accounts with `id`, `name`, `visibility` |
| PUT | `/api/teamcalendar/accounts/{id}` | `{visibility}`, as above |
