import { Sidebar } from '@/components/ui/navigation/sidebar'


export default function Layout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
      <div className="mx-auto max-w-screen-1xl">
        <Sidebar />
        <div  className="lg:pl-72">
          <main className="p-4 sm:px-6 sm:pb-10 sm:pt-10 lg:px-10 lg:pt-7">
            {children}
          </main>
        </div>
      </div>
  )
}
