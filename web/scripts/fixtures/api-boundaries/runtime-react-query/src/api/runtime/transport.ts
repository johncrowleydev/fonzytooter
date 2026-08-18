import * as query from '@tanstack/react-query'

export function useTransport() {
  return query.useQuery({ queryKey: ['transport'] })
}
