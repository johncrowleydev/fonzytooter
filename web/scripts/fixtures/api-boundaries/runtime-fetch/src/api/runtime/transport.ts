export async function stream(request: RequestInfo | URL, init?: RequestInit) {
  return fetch(request, init)
}
