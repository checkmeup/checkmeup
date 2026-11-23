-- +goose Up

-- PRD alert-delivery: a monitor has at most one open incident. Older code
-- could leave more (two processes checking one monitor during a deploy) and
-- leave one open on a monitor that isn't down (a failed resolve, a ping racing
-- the worker), so those are resolved first: on a monitor not down, every open
-- one; elsewhere, all but the newest.
UPDATE uptime_incidents i SET resolved_at = NOW() FROM uptime_monitors m
WHERE m.id = i.monitor_id AND i.resolved_at IS NULL AND m.status <> 'down';
UPDATE port_incidents i SET resolved_at = NOW() FROM port_monitors m
WHERE m.id = i.monitor_id AND i.resolved_at IS NULL AND m.status <> 'down';
UPDATE dns_incidents i SET resolved_at = NOW() FROM dns_monitors m
WHERE m.id = i.monitor_id AND i.resolved_at IS NULL AND m.status <> 'down';
UPDATE cron_incidents i SET resolved_at = NOW() FROM cron_monitors m
WHERE m.id = i.monitor_id AND i.resolved_at IS NULL AND m.status <> 'down';

UPDATE uptime_incidents i SET resolved_at = NOW()
WHERE i.resolved_at IS NULL AND EXISTS (
    SELECT 1 FROM uptime_incidents j WHERE j.monitor_id = i.monitor_id AND j.resolved_at IS NULL
      AND (j.started_at, j.id) > (i.started_at, i.id));
UPDATE port_incidents i SET resolved_at = NOW()
WHERE i.resolved_at IS NULL AND EXISTS (
    SELECT 1 FROM port_incidents j WHERE j.monitor_id = i.monitor_id AND j.resolved_at IS NULL
      AND (j.started_at, j.id) > (i.started_at, i.id));
UPDATE dns_incidents i SET resolved_at = NOW()
WHERE i.resolved_at IS NULL AND EXISTS (
    SELECT 1 FROM dns_incidents j WHERE j.monitor_id = i.monitor_id AND j.resolved_at IS NULL
      AND (j.started_at, j.id) > (i.started_at, i.id));
UPDATE cron_incidents i SET resolved_at = NOW()
WHERE i.resolved_at IS NULL AND EXISTS (
    SELECT 1 FROM cron_incidents j WHERE j.monitor_id = i.monitor_id AND j.resolved_at IS NULL
      AND (j.started_at, j.id) > (i.started_at, i.id));

CREATE UNIQUE INDEX uptime_incidents_one_open ON uptime_incidents (monitor_id) WHERE resolved_at IS NULL;
CREATE UNIQUE INDEX port_incidents_one_open ON port_incidents (monitor_id) WHERE resolved_at IS NULL;
CREATE UNIQUE INDEX dns_incidents_one_open ON dns_incidents (monitor_id) WHERE resolved_at IS NULL;
CREATE UNIQUE INDEX cron_incidents_one_open ON cron_incidents (monitor_id) WHERE resolved_at IS NULL;

-- +goose Down
DROP INDEX cron_incidents_one_open;
DROP INDEX dns_incidents_one_open;
DROP INDEX port_incidents_one_open;
DROP INDEX uptime_incidents_one_open;
