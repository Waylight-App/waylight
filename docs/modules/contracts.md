# Contracts

**What it does:** Household contracts with their cost and cancellation deadline, and a reminder through your notification channels before each deadline.

## Using it

Add the *Contracts* widget and enter a contract with **+ New contract**:

- **Cost** and how often it is due (monthly, quarterly or yearly). The widget
  adds everything up to a monthly total.
- **End of term**: the last day of the current term, as it says on the
  contract.
- **Notice period** in days. The cancellation deadline is the end of the term
  minus the notice period. Three months are entered as 90 days.
- **Renews by** in months. A contract that is not cancelled usually runs on
  for another 12 months, or one month for many mobile and streaming plans.
  Once a deadline has passed, Waylight moves the term on by this period and
  shows the next deadline. Enter 0 for a contract that simply ends.
- **Remind** a number of days before the deadline; 0 sends no reminder.

The widget lists the contracts by their next deadline and marks the ones
inside their reminder period. The compact view names the next deadline and
asks for attention once it is close. The currency the amounts are shown in is
a widget setting.

## Reminders

Once a day after the reminder period starts, Waylight sends one reminder per
deadline through the [notification channels](../notifications.md) of the
instance, with the deadline, the end of the term and what happens otherwise.
Changing the term or the notice period moves the deadline, which then gets a
reminder of its own.

Reminders cover the contracts the admin sees: their own and those shared
with them. Members' private contracts are never sent to the instance's
channels.

## Who sees what

New contracts are private for members and admins, like notes, and can be
shared with a group or everyone. Only the person who entered a contract can
change or remove it.

## Configuration

No environment variables. `MODULES_DISABLED=contracts` switches the module
off.

## Outbound requests

None; reminders leave through the notification channels you set up.

## API

`/api/contracts` (list, create, update, delete) and `/api/contracts/compact`
– see `api/modules/contracts.yaml`.
