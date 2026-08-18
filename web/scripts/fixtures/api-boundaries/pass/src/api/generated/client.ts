export async function generatedClient() {
  const response = await fetch('/api/health')
  return response
}
