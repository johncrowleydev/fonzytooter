import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

export function useFeature() {
  useQueryClient()
  useQuery({ queryKey: ['feature'], queryFn: async () => 'feature' })
  return useMutation({ mutationFn: async () => 'feature' })
}
