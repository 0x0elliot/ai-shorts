"use client"

import { ThemeProvider } from "next-themes"
import { Inter } from "next/font/google"
import { Toaster } from "@/components/ui/toaster"
import "./globals.css"

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
})

import { Sidebar } from "@/components/ui/navigation/sidebar"
import { siteConfig } from "./siteConfig"
import { isLastDayOfMonth } from "date-fns";

// export const metadata: Metadata = {
//   metadataBase: new URL("https://yoururl.com"),
//   title: siteConfig.name,
//   description: siteConfig.description,
//   keywords: [],
//   authors: [
//     {
//       name: "yourname",
//       url: "",
//     },
//   ],
//   creator: "yourname",
//   openGraph: {
//     type: "website",
//     locale: "en_US",
//     url: siteConfig.url,
//     title: siteConfig.name,
//     description: siteConfig.description,
//     siteName: siteConfig.name,
//   },
//   icons: {
//     icon: "/favicon.ico",
//   },
// }



export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Bungee&display=swap" rel="stylesheet" />

        <link rel="icon" href="/favicon.ico" />
        <link href="/public/apple-touch-icon.png" rel="apple-touch-icon" />
        <link rel="favicon-32x32" href="/public/favicon-32x32.png" />
        <link rel="favicon-16x16" href="/public/favicon-16x16.png" />
        <link href="public/android-chrome-192x192.png" rel="icon" sizes="192x192" />
        <link href="public/android-chrome-512x512.png" rel="icon" sizes="512x512" />
        <link rel="manifest" href="/public/site.webmanifest" />

        <link href="https://fonts.googleapis.com/css2?family=Bungee&family=Roboto:ital,wght@0,100;0,300;0,400;0,500;0,700;0,900;1,100;1,300;1,400;1,500;1,700;1,900&display=swap" rel="stylesheet" />
      </head>

      <body
        className={`${inter.className} overflow-y-scroll scroll-auto antialiased selection:bg-orange-100 selection:text-orange-700 dark:bg-gray-950`}
        suppressHydrationWarning
      >
        <ThemeProvider defaultTheme="system" attribute="class">
            <main>{children}</main>
            <Toaster />
        </ThemeProvider>
      </body>
    </html>
  )
}
