export function safeExternalUrl(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : null
  } catch {
    return null
  }
}

/**
 * The publisher of a link, for display alongside a citation or video title.
 *
 * This exists because those rows used to show the resource's curriculum ID, which told a learner
 * nothing. "arxiv.org" answers the question the ID was sitting in the space of.
 */
export function externalHost(value: string) {
  const url = safeExternalUrl(value)

  if (!url) {
    return undefined
  }

  return new URL(url).hostname.replace(/^www\./, '')
}
