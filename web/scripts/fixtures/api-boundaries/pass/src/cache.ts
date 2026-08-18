import { useQueryClient } from '@tanstack/react-query'

export function useCache() {
  return useQueryClient()
}
