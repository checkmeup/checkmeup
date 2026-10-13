# Frontend structure

Date: 2026-10-13

The `use*Monitors.ts` composables are shallow on purpose: they keep the TanStack Query keys in one place. An architecture review should not flag them as pass-throughs to inline.
