export function safeReturnTo(value: string | null | undefined, fallback = '/') {
  if (!value || !value.startsWith('/') || value.startsWith('//') || value.includes('\\')) {
    return fallback
  }

  try {
    const origin = window.location.origin
    const destination = new URL(value, origin)
    if (destination.origin !== origin) return fallback
    return `${destination.pathname}${destination.search}${destination.hash}`
  } catch {
    return fallback
  }
}
