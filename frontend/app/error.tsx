'use client'

import { useEffect } from 'react'
import * as Sentry from '@sentry/nextjs'
import { Button } from '@/components/ui/button'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error('Global app error:', error)
    Sentry.captureException(error)
  }, [error])

  return (
    <div className="min-h-screen flex items-center justify-center p-6 bg-zinc-50">
      <div className="max-w-md w-full rounded-3xl border border-zinc-200 bg-white p-8 text-center space-y-4">
        <p className="text-xs font-bold uppercase tracking-widest text-zinc-400">UTSAV System Notice</p>
        <h2 className="text-2xl font-bold text-zinc-900">Something went wrong</h2>
        <p className="text-sm text-zinc-500">
          We hit an unexpected issue. Please retry this action.
        </p>
        <Button onClick={reset} className="bg-orange-600 hover:bg-orange-700 text-white rounded-xl">
          Try Again
        </Button>
      </div>
    </div>
  )
}
