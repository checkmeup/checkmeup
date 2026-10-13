# WhatsApp alerts

Status: proposed

## Problem

In many countries WhatsApp is where people actually look; an alert there is seen faster than an email.

## Users

Anyone who wants down and recovery alerts on WhatsApp.

## Success signal

A user verifies a number and receives the next down alert on WhatsApp within one check cycle.

## Scope

- WhatsApp notification channels: several numbers per org, each confirmed by an opt-in step before alerts start.
- Down and recovery alerts through Meta-approved message templates.
- A test alert before relying on it.

## Non-goals

- Two-way conversations or acknowledging alerts from WhatsApp.

## Acceptance criteria

- A number starts receiving alerts only after its owner completes the opt-in step.
- "Send test alert" reports whether the message was delivered.
- A down alert uses the approved template with the monitor's name, type, reason and time; a recovery alert uses its own template with the downtime.
- Removing a channel deletes the stored number; adding it again needs the opt-in again.
- A failed WhatsApp delivery doesn't stop the monitor's other channels.

## Open questions

- Which provider: Meta's Cloud API directly or a BSP (see the Inbox).
- Template wording, and Meta's approval lead time.
- How the per-message cost is bounded: SMS-style credits (ADR billing) or something else.
