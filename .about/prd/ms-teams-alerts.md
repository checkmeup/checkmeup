# Microsoft Teams alerts

Status: proposed

## Problem

Companies on Microsoft 365 want alerts in a Teams channel.

## Users

Teams using Microsoft Teams.

## Success signal

A user pastes a workflow webhook URL and the next alert appears in the Teams channel as a readable card.

## Scope

- Teams notification channels from Power Automate workflow webhook URLs, several per org.
- Down and recovery alerts as Adaptive Cards.
- Delivery results shown in Settings, without retries.

## Non-goals

- The retired Office 365 "Incoming Webhook" connector.
- A Teams bot or app.

## Acceptance criteria

- Only a URL matching the Power Automate workflow webhook pattern is accepted.
- "Send test message" reports whether Teams accepted it.
- A down card shows the monitor's name, type, reason and time; a recovery card shows the downtime.
- A delivery's failure is recorded on the channel and shown in Settings, and doesn't stop the monitor's other channels.

## Open questions

- Microsoft's current recommended webhook mechanism, checked right before building (see the Inbox).
- A Microsoft 365 work or school tenant to build and test against, such as the free Developer Program.
