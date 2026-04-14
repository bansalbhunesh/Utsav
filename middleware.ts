import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('utsav_access_token') || request.headers.get('Authorization')

  // List of protected routes
  const protectedPaths = ['/dashboard', '/organiser', '/events/create', '/management']
  
  const isProtected = protectedPaths.some(path => request.nextUrl.pathname.startsWith(path))

  // Note: For now, we only check for token existence. 
  // In a full production app, you'd verify the JWT here.
  if (isProtected && !token) {
    // Redirect to login if accessing protected page without token
    const url = request.nextUrl.clone()
    url.pathname = '/login'
    return NextResponse.redirect(url)
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}
