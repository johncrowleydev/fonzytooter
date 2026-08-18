import * as ReactQuery from '@tanstack/react-query'

export function useFeature() {
  return ReactQuery.useQuery({ queryKey: ['feature'], queryFn: async () => 'feature' })
}
