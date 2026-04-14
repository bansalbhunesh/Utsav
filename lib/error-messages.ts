import { ApiError } from '@/lib/api'

export function getUserFacingError(err: unknown, fallback: string): string {
  if (err instanceof ApiError) {
    if (err.status === 401) return 'Your session expired. Please log in again.'
    if (err.status === 403) return 'You do not have permission to perform this action.'
    if (err.status === 404) return 'The requested resource could not be found.'
    if (err.status >= 500) return 'Server issue detected. Please try again shortly.'
    return err.message || fallback
  }

  if (err instanceof Error && err.message) return err.message
  return fallback
}
