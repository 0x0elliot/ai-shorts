import { create } from "domain"

export const siteConfig = {
  name: "zappush",
  url: "https://dashboard.tremor.so",
  description: "The only dashboard you will ever need.",
  baseLinks: {
    home: "/",
    dashboard: "/dashboard",
    create  : "/create",
  },
  externalLink: {
    blocks: "https://blocks.tremor.so/templates#dashboard",
  },
  baseApiUrl: "http://localhost:5002",
}

export type siteConfig = typeof siteConfig
