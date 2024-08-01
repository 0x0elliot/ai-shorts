"use client"
import { siteConfig } from "@/app/siteConfig"
import { cx, focusRing } from "@/lib/utils"
import {
  RiHome2Line,
  RiLinkM
} from "@remixicon/react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { useEffect, useState } from "react"
import MobileSidebar from "./MobileSidebar"
import {
  WorkspacesDropdownMobile
} from "./SidebarWorkspacesDropdown"
import { UserProfileDesktop, UserProfileMobile } from "./UserProfile"

const navigation = [
  { name: "Dashboard", href: siteConfig.baseLinks.dashboard, icon: RiHome2Line },
  { name: "Create", href: siteConfig.baseLinks.create, icon: RiLinkM },

  {
    name: "Discord",
    href: "https://discord.gg/y6yPVWjdD9",
    icon: RiLinkM,
  }
] as const

const shortcuts = [
  // {
  //   name: "Add new user",
  //   href: "#",
  //   icon: RiLinkM,
  // },
  // {
  //   name: "Workspace usage",
  //   href: "#",
  //   icon: RiLinkM,
  // },
  // {
  //   name: "Cost spend control",
  //   href: "#",
  //   icon: RiLinkM,
  // },
  // {
  //   name: "Overview – Rows written",
  //   href: "#",
  //   icon: RiLinkM,
  // },
  {
    name: "Create a new campaign",
    href: siteConfig.baseLinks.campaigns,
    icon: RiLinkM,
  },
  {
    name: "Campaign history",
    href: siteConfig.baseLinks.campaign_history,
    icon: RiLinkM,
  },

] as const

export function Sidebar() {
  const [userInfo, setUserInfo] = useState(null);

  const pathname = usePathname()
  const [workspaces, setWorkspaces] = useState([
    {
      value: "retail-analytics",
      name: "Your shop",
      initials: "SHOP",
      role: "Member",
      color: "bg-orange-600 dark:bg-orange-500",
    },
  ])

  useEffect(() => {
    setUserInfo(JSON.parse(localStorage.getItem('userinfo') || "{}" ));
  }, [])

  const isActive = (itemHref: string) => {
    if (itemHref === siteConfig.baseLinks.settings) {
      return pathname.startsWith("/settings")
    }
    return pathname === itemHref;
  }
  return (
    <>
      {/* sidebar (lg+) */}
      <nav className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
        <aside className="flex grow flex-col gap-y-6 overflow-y-auto border-r border-gray-200 bg-white p-4 dark:border-gray-800 dark:bg-gray-950">
          <>
          <nav
              aria-label="core navigation links"
              className="flex flex-1 flex-col space-y-10"
            >
              <ul role="list" className="space-y-0.5">
                {navigation.map((item) => (
                  <li key={item.name}>
                    <Link
                      href={item.href}
                      className={cx(
                        isActive(item.href)
                          ? "text-orange-600 dark:text-orange-400"
                          : "text-gray-700 hover:text-gray-900 dark:text-gray-400 hover:dark:text-gray-50",
                        "flex items-center gap-x-2.5 rounded-md px-2 py-1.5 text-sm font-medium transition hover:bg-gray-100 hover:dark:bg-gray-900",
                        focusRing
                      )}
                      target={item.name === "Discord" ? "_blank" : "_self"}
                    >
                      <item.icon className="size-4 shrink-0" aria-hidden="true" />
                      {item.name}
                    </Link>
                  </li>
                ))}
              </ul>
              {/* <div>
      <span className="text-xs font-medium leading-6 text-gray-500">
        Shortcuts
      </span>
      <ul aria-label="shortcuts" role="list" className="space-y-0.5">
        {shortcuts.map((item) => (
          <li key={item.name}>
            <Link
              href={item.href}
              className={cx(
                pathname === item.href || pathname.startsWith(item.href)
                  ? "text-orange-600 dark:text-orange-400"
                  : "text-gray-700 hover:text-gray-900 dark:text-gray-400 hover:dark:text-gray-50",
                "flex items-center gap-x-2.5 rounded-md px-2 py-1.5 text-sm font-medium transition hover:bg-gray-100 hover:dark:bg-gray-900",
                focusRing,
              )}
            >
              <item.icon
                className="size-4 shrink-0"
                aria-hidden="true"
              />
              {item.name}
            </Link>
          </li>
        ))}
      </ul>
    </div> */}
            </nav><div className="mt-auto">
                <UserProfileDesktop />
              </div></>
        </aside>
      </nav>
      {/* top navbar (xs-lg) */}
      <div className="sticky top-0 z-40 flex h-16 shrink-0 items-center justify-between border-b border-gray-200 bg-white px-2 shadow-sm sm:gap-x-6 sm:px-4 lg:hidden dark:border-gray-800 dark:bg-gray-950">
        <WorkspacesDropdownMobile workspaces={workspaces} />
        <div className="flex items-center gap-1 sm:gap-2">
          <UserProfileMobile />
          <MobileSidebar />
        </div>
      </div>
    </>
  )
}
