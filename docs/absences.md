# Absences

With a licence key that includes `absences` ([licence keys](licence-keys.md)),
everyone on the instance can see who is away: on holiday, sick, working
remotely or away for another reason. This needs a server built with the
`ee` features.

## Who enters what

- Members and admins enter their own absences: the first and last day
  (both included), the kind, an optional note of up to 200 characters.
- Admins enter and change absences for anyone, for example for someone who
  called in sick.
- One absence spans at most a year.
- Deleting an account deletes its absences.

## Who sees what

Everyone signed in sees who is away on which days, as do guests and kiosk
displays whose shared tab shows the absences.

An absence marked **private** shows others only *away*, without the kind
and the note. The person and admins still see everything. Use it for sick
days: the reason can be health data. The audit log records that an absence
was entered, changed or deleted, never its kind or note.

With an expired licence key or too few seats, absences stay readable, but
nothing can be entered.

## API

| Method | Path | |
| --- | --- | --- |
| GET | `/api/absences?from=YYYY-MM-DD&to=YYYY-MM-DD` | Absences touching these days, at most 62 days at once. Each has `name`, `kind` (`holiday`, `sick`, `remote`, `other`, or `away` for someone else's private absence), `note`, `private`, `editable`. |
| POST | `/api/absences` | `{from, to, kind, note, private}`; admins can add `accountId` |
| PUT | `/api/absences/{id}` | Same fields, without `accountId` |
| DELETE | `/api/absences/{id}` | |
