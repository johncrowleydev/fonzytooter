type Daypart = 'morning' | 'afternoon' | 'evening'

export function formatDashboardDate(now: Date, timeZone: string): string {
  return new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    timeZone,
  }).format(now)
}

export function getDashboardDaypart(now: Date, timeZone: string): Daypart {
  const hour = Number(
    new Intl.DateTimeFormat(undefined, {
      hour: 'numeric',
      hourCycle: 'h23',
      timeZone,
    })
      .formatToParts(now)
      .find((part) => part.type === 'hour')?.value,
  )

  if (hour < 12) return 'morning'
  if (hour < 17) return 'afternoon'
  return 'evening'
}

export function formatDashboardGreeting(now: Date, timeZone: string): string {
  return `Good ${getDashboardDaypart(now, timeZone)}, Fonzy.`
}
