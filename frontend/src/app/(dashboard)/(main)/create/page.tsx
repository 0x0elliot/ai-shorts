"use client"

import { siteConfig } from "@/app/siteConfig"

export default function Create() {
    return (
        <>
            <section aria-labelledby="flows-title">
                <h1
                    id="overall-title"
                    className="scroll-mt-10 text-lg font-semibold text-gray-900 sm:text-xl dark:text-gray-50"
                >
                    Create new video schedules
                </h1>
            </section><section aria-labelledby="flows-description" className="mb-4">
                <p
                    id="overall-description"
                    className="text-sm text-gray-500 dark:text-gray-400"
                >
                    Create new video schedules for your social media accounts.
                </p>
            </section>
        </>
    )
}
