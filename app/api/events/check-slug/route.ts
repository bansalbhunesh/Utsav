import { supabase } from '@/lib/supabase/client'
import { NextRequest, NextResponse } from 'next/server'

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url)
  const slug = searchParams.get('slug')

  if (!slug) {
    return NextResponse.json({ error: 'Slug is required' }, { status: 400 })
  }

  // Real-time check against events table
  const { data, error } = await supabase
    .from('events')
    .select('id')
    .eq('slug', slug)
    .single()

  if (error && error.code !== 'PGRST116') { // PGRST116 means zero rows found
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  const isAvailable = !data

  return NextResponse.json({ available: isAvailable })
}
