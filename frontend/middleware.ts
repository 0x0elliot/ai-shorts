import { NextRequest, NextResponse } from 'next/server'
 
export function middleware(request: NextRequest) {
  const accessToken = (request.cookies.get('access_token')?.value || "").length > 0
  const refreshToken = (request.cookies.get('refresh_token')?.value || "").length > 0

  const cookiesPresent = accessToken && refreshToken
    if (!cookiesPresent && !(request.nextUrl.pathname.startsWith('/magic-link-auth') || request.nextUrl.pathname.startsWith('/login') || request.nextUrl.pathname.startsWith('/logout'))) {
        console.log("Redirecting to login")
        const url = new URL('/login', request.url)
        return NextResponse.redirect(url, { status: 302 })
    }
}
 
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|.*\\.png$).*)'],
}