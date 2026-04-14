import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { jwtVerify } from 'jose'

const protectedPaths = ['/dashboard', '/organiser', '/billing', '/events']
const textEncoder = new TextEncoder()

function isProtectedPath(pathname: string): boolean {
  return protectedPaths.some((path) => pathname.startsWith(path))
}

async function verifyAccessToken(token: string): Promise<boolean> {
  const secret = process.env.JWT_SECRET
  if (!secret) return false
  try {
    const { payload } = await jwtVerify(token, textEncoder.encode(secret), {
      algorithms: ['HS256'],
    })
    return typeof payload.sub === 'string' && payload.sub.length > 0
  } catch {
    return false
  }
}

export async function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname
  const token = request.cookies.get('utsav_access_token')?.value
  const isProtected = isProtectedPath(pathname)

  // If already authenticated, avoid showing login screen.
  if (pathname === '/login' && token) {
    const valid = await verifyAccessToken(token)
    if (valid) {
      const url = request.nextUrl.clone()
      url.pathname = '/dashboard'
      return NextResponse.redirect(url)
    }
  }

  if (isProtected && !token) {
    const url = request.nextUrl.clone()
    url.pathname = '/login'
    return NextResponse.redirect(url)
  }

  if (isProtected && token) {
    const valid = await verifyAccessToken(token)
    if (!valid) {
      const url = request.nextUrl.clone()
      url.pathname = '/login'
      const response = NextResponse.redirect(url)
      response.cookies.delete('utsav_access_token')
      return response
    }
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
