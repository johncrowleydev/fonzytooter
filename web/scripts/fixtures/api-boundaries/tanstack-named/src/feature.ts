import { useQuery } from '@tanstack/react-query'

export function useFeature() {
  return useQuery({ queryKey: ['feature'], queryFn: async () => 'feature' })
}
