package handler

import (
	"context"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
)

// attachDefaultNotificationChannels attaches every enabled channel the org
// currently has to a newly created monitor — a new monitor defaults to all
// of the org's enabled channels (US-2802), matching the pre-EP-28 implicit
// behavior of every monitor alerting on every org-level channel.
func (h *MonitorHandler) attachDefaultNotificationChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID) {
	channels, err := h.queries.ListEnabledNotificationChannels(ctx, orgID)
	if err != nil {
		return
	}
	for _, c := range channels {
		_ = h.queries.InsertMonitorNotificationChannel(ctx, db.InsertMonitorNotificationChannelParams{
			ChannelID: c.ID, MonitorType: monitorType, MonitorID: monitorID,
		})
	}
}

// setMonitorNotificationChannels replaces a monitor's attached channels with
// channelIDs, dropping any ID that doesn't resolve to a channel owned by
// orgID — same ownership-scoping approach as resolveMonitorName.
func (h *MonitorHandler) setMonitorNotificationChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID, channelIDs []string) error {
	owned, err := h.queries.ListNotificationChannels(ctx, orgID)
	if err != nil {
		return err
	}
	ownedSet := make(map[uuid.UUID]bool, len(owned))
	for _, c := range owned {
		ownedSet[c.ID] = true
	}

	if err := h.queries.DeleteMonitorNotificationChannels(ctx, db.DeleteMonitorNotificationChannelsParams{
		MonitorType: monitorType, MonitorID: monitorID,
	}); err != nil {
		return err
	}
	for _, idStr := range channelIDs {
		id, err := uuid.Parse(idStr)
		if err != nil || !ownedSet[id] {
			continue
		}
		if err := h.queries.InsertMonitorNotificationChannel(ctx, db.InsertMonitorNotificationChannelParams{
			ChannelID: id, MonitorType: monitorType, MonitorID: monitorID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// attachMonitorChannels attaches the requested channels to a newly created
// monitor, or falls back to every enabled org channel when none were
// explicitly selected (US-2802).
func (h *MonitorHandler) attachMonitorChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID, channelIDs []string) error {
	if len(channelIDs) > 0 {
		return h.setMonitorNotificationChannels(ctx, orgID, monitorType, monitorID, channelIDs)
	}
	h.attachDefaultNotificationChannels(ctx, orgID, monitorType, monitorID)
	return nil
}

// loadNotificationChannelIDs returns the channel IDs attached to a monitor,
// for inclusion in its API response (edit form pre-selection).
func (h *MonitorHandler) loadNotificationChannelIDs(ctx context.Context, monitorType string, monitorID uuid.UUID) []string {
	ids, err := h.queries.ListMonitorNotificationChannelIDs(ctx, db.ListMonitorNotificationChannelIDsParams{
		MonitorType: monitorType, MonitorID: monitorID,
	})
	if err != nil {
		return []string{}
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
