export function youtubeEmbedUrl(youtubeId: string) {
  return `https://www.youtube-nocookie.com/embed/${encodeURIComponent(youtubeId)}?rel=0`
}

export function youtubeWatchUrl(youtubeId: string) {
  return `https://www.youtube.com/watch?v=${encodeURIComponent(youtubeId)}`
}

export function youtubeThumbnailUrl(youtubeId: string) {
  return `https://i.ytimg.com/vi/${encodeURIComponent(youtubeId)}/hqdefault.jpg`
}
