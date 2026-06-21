import { useQuery } from '@tanstack/vue-query'
import { notificationChannelsApi } from '@/api/notificationChannels'

export function useNotificationChannels() {
  return useQuery({
    queryKey: ['notification-channels'],
    queryFn: notificationChannelsApi.list,
  })
}
