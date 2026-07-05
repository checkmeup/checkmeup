# EP-32: Multi-region checking

Today every check runs from the single Hetzner CX23 ([ADR-006](../decisions/006-infrastructure-hetzner-kamal-traefik.md)) — a network blip between that one VPS and a target can register as a false "down" that has nothing to do with the target's actual health. Multi-region checking runs the same check from more than one geographic location and only alerts when a majority of regions agree, closing UptimeRobot/Cronitor's biggest false-positive advantage (see `docs/proposals/bucket-list.md`).

**Needs an infra decision before implementation** (add to [decision backlog](../decisions/backlog.md) / write an ADR): the current architecture is one VPS with no external queue or broker ([ADR-001](../decisions/001-worker-model.md)) and no managed scaling ([ADR-006](../decisions/006-infrastructure-hetzner-kamal-traefik.md)). Multi-region checking needs compute in more than one geographic location, which this stack doesn't have today. Options include additional small Hetzner (or other provider) VPS instances in other regions reporting results back to the single Postgres instance — keeps ADR-001's no-broker constraint, but adds servers to provision, patch, and pay for — versus a check-execution-as-a-service provider. Pick an approach, including how many regions and where, before US-3201 starts.

---

### US-3201: Choose check regions for a monitor

**As a** user, **I want** to pick which regions check my uptime monitor **so that** I control the tradeoff between coverage and noise.

**Estimate:** 2 h (excludes the infra provisioning work covered by the decision above)

**Acceptance criteria:**

- [ ] Multi-select of available regions per uptime monitor (region list depends on the infra decision above)
- [ ] Default: today's single existing region — opt-in, so all current monitors behave exactly as before
- [ ] Number of selectable regions gated per plan tier, same enforcement pattern as other plan limits ([ADR-019](../decisions/019-plan-limits.md))

---

### US-3202: Run the check from each selected region independently

**As a** platform, **I want** each region to execute the check on its own schedule **so that** regions don't need to coordinate with each other.

**Estimate:** 3 h

**Acceptance criteria:**

- [ ] Each selected region runs the monitor's existing check (and any EP-31 assertions) on the monitor's configured interval, independently
- [ ] Per-region result (status, response time, error) recorded separately, not merged at write time
- [ ] A region becoming unavailable (e.g. that regional worker is down) doesn't block or delay checks from the other regions

---

### US-3203: Quorum-based down determination

**As a** user, **I want** a monitor to only go "down" when most regions agree **so that** I'm not alerted for a problem that's actually on checkmeup's side, not mine.

**Estimate:** 2.5 h

**Acceptance criteria:**

- [ ] Monitor transitions to "down" only when a majority (or configurable threshold — "any region" vs "all regions") of selected regions report failure within the same check window
- [ ] A single region's failure while others are healthy is shown as a regional anomaly, not a down state
- [ ] Single-region monitors (the default, US-3201) keep today's behavior exactly — no quorum question when there's only one region

---

### US-3204: View per-region status

**As a** user, **I want** to see which regions are failing **so that** I can tell a real outage from a regional network issue.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Monitor detail shows a per-region breakdown: status, response time, last checked, for each selected region
- [ ] Alert message names which regions failed vs passed, not just an aggregate "down"
