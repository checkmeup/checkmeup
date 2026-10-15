import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { projectsApi, type Project } from '@/api/projects'

export function useProjects() {
  return useQuery({ queryKey: ['projects'], queryFn: projectsApi.list })
}

// Creates a project and adds it to the cached list, for creating one from a
// monitor's form without refetching (ProjectField).
export function useAddProject() {
  const queryClient = useQueryClient()
  return async (name: string): Promise<Project> => {
    const p = await projectsApi.create(name)
    queryClient.setQueryData<Project[]>(['projects'], (old) =>
      [...(old ?? []), p].sort((a, b) => a.name.localeCompare(b.name)),
    )
    return p
  }
}
