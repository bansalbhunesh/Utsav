import { NextRequest, NextResponse } from 'next/server'

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url)
  const slug = searchParams.get('slug')
  const API_URL = process.env.API_URL?.trim()

  if (!slug) {
    return NextResponse.json({ error: 'Slug is required' }, { status: 400 })
  }
  if (!API_URL) {
    return NextResponse.json({ error: 'API_URL is required' }, { status: 500 })
  }

  try {
    const res = await fetch(`${API_URL}/v1/public/events/check-slug?slug=${slug}`)
    if (!res.ok) throw new Error('API request failed')
    const data = await res.json()
    return NextResponse.json(data)
  } catch {
    return NextResponse.json({ error: 'Failed to check slug availability' }, { status: 500 })
  }
}
