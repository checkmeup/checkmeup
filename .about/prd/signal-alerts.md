# Signal alerts

Status: proposed

## Problem

Some privacy-minded users want alerts on Signal rather than Telegram or WhatsApp.

## Users

Users who prefer Signal; demand is unconfirmed.

## Success signal

Not set until there is a go decision backed by real demand.

## Scope

- Signal notification channels, several numbers per org.
- Down and recovery alerts sent through a self-hosted `signal-cli` service.
- Delivery failures shown in Settings.

## Non-goals

- Anything that depends on an official Signal bot API, which doesn't exist.

## Acceptance criteria

- "Send test alert" reports whether the message reached the number.
- A down alert names the monitor, its type, the reason and the time; a recovery alert names the monitor and the downtime.
- When `signal-cli` is unreachable, the failure shows in Settings ("Last delivery: failed, connection refused") and the monitor's other channels still deliver.

## Open questions

- Whether to build it at all: it needs an always-on `signal-cli` process and a bot phone number Signal may ban (see the Inbox).
