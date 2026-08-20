export function downloadPdf(blob: Blob, disposition: string | undefined, fallbackFilename: string) {
  const filename = filenameFromContentDisposition(disposition) ?? fallbackFilename
  const objectURL = URL.createObjectURL(blob)
  try {
    const link = document.createElement('a')
    link.href = objectURL
    link.download = filename
    link.click()
  } finally {
    window.setTimeout(() => URL.revokeObjectURL(objectURL), 0)
  }
}

function filenameFromContentDisposition(disposition: string | undefined) {
  const match = disposition?.match(/filename="([^"]+)"/i)
  return match?.[1]
}

export function pdfDownloadErrorMessage(error: unknown) {
  if (error instanceof Error && 'status' in error && error.status === 503) {
    return 'PDF downloads are temporarily unavailable because the rendering tools are not installed.'
  }
  return 'The PDF could not be downloaded. Please try again.'
}
