const localThing = {
  useMutation() {
    return 'local mutation state'
  },
  useQuery() {
    return 'local query state'
  },
}

export function useFeature() {
  return [localThing.useQuery(), localThing.useMutation()]
}
