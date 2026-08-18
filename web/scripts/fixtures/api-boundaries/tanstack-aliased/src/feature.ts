import { useMutation as useServerMutation, useQuery as useServerQuery } from '@tanstack/react-query'

export function useFeature() {
  useServerQuery({ queryKey: ['feature'], queryFn: async () => 'feature' })
  return useServerMutation({ mutationFn: async () => 'feature' })
}
