"use client"

import * as React from "react"
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts"

import confetti from 'canvas-confetti';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart"

import cookies from 'nookies';
import axios from "axios"
import { siteConfig } from "@/app/siteConfig"
import Link from "next/link"

const chartConfig = {
  views: {
    label: "Page Views",
  },
  desktop: {
    label: "Desktop",
    color: "hsl(var(--chart-1))",
  },
  mobile: {
    label: "Mobile",
    color: "hsl(var(--chart-2))",
  },
} satisfies ChartConfig

export function DashboardStatistics() {
  const [chartData, setChartData] = React.useState<{
    date: string
    desktop: number
    mobile: number
  }[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    let accessToken = cookies.get(null).access_token;

    let userinfoStr = localStorage.getItem('userinfo');
    let userinfo;

    try {
      userinfo = JSON.parse(userinfoStr);
    } catch (error) {
      console.error(error)
      // window.location.href = "/logout";
    }

    console.log(userinfo);

    let shopId = userinfo?.current_shop_id

    setLoading(true)
    axios.get(`${siteConfig.baseApiUrl}/api/tracking/private/stats/devices?shop_id=${shopId}`, {
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${accessToken}`,
      },
    })
      .then((response) => {
        setChartData(response.data.stats);
      })
      .catch((error) => {
        console.error(error);
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  const [activeChart, setActiveChart] =
    React.useState<keyof typeof chartConfig>("desktop")

  const total = React.useMemo(
    () => ({
      desktop: chartData.reduce((acc, curr) => acc + curr?.desktop, 0),
      mobile: chartData.reduce((acc, curr) => acc + curr?.mobile, 0),
    }),
    [chartData]
  )

  const NoDataDisplay = () => {
    const [mouseOver, setMouseOver] = React.useState(false)

    return (
      <div className="relative flex flex-col items-center justify-center h-[250px] text-center p-6 bg-gradient-to-r from-purple-400 via-pink-500 to-red-500 rounded-lg shadow-lg overflow-hidden">
        <h3 className="text-2xl font-bold text-white mb-2">No Data? No Problem!</h3>
        <p className="text-white mb-6">Time to make some magic happen!</p>
        <button
          onMouseEnter={() => setMouseOver(true)}
          onMouseLeave={() => setMouseOver(false)}
          onClick={() => window.location.href = '/campaigns'}
          className="px-6 py-3 bg-yellow-400 text-gray-900 rounded-full font-bold text-lg transform transition duration-200 hover:scale-105 hover:bg-yellow-300 focus:outline-none focus:ring-2 focus:ring-yellow-600 focus:ring-opacity-50 z-10"
        >
          {mouseOver ? "Abracadabra! 🎩✨" : "Start Your First Campaign!"}
        </button>
      </div>
    )
  }


  return (
    <Card>
      <CardHeader className="flex flex-col items-stretch space-y-0 border-b p-0 sm:flex-row">
        <div className="flex flex-1 flex-col justify-center gap-1 px-6 py-5 sm:py-6">
          <CardTitle>Click Statistics</CardTitle>
          <CardDescription>
            Showing total visitors for the last 3 months
          </CardDescription>
        </div>
        <div className="flex">
          {["desktop", "mobile"].map((key) => {
            const chart = key as keyof typeof chartConfig
            return (
              <button
                key={chart}
                data-active={activeChart === chart}
                className="relative z-30 flex flex-1 flex-col justify-center gap-1 border-t px-6 py-4 text-left even:border-l data-[active=true]:bg-muted/50 sm:border-l sm:border-t-0 sm:px-8 sm:py-6"
                onClick={() => setActiveChart(chart)}
              >
                <span className="text-xs text-muted-foreground">
                  {chartConfig[chart].label}
                </span>
                <span className="text-lg font-bold leading-none sm:text-3xl">
                  {total[key as keyof typeof total].toLocaleString()}
                </span>
              </button>
            )
          })}
        </div>
      </CardHeader>
      <CardContent className="px-2 sm:p-6">
        {loading ? (
          <div className="flex items-center justify-center h-[250px]">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
          </div>
        ) : chartData.length === 0 ? (
          <NoDataDisplay />
        ) : (
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[250px] w-full"
          >
            <BarChart
              accessibilityLayer
              data={chartData}
              margin={{
                left: 12,
                right: 12,
              }}
            >
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const date = new Date(value)
                  return date.toLocaleDateString("en-US", {
                    month: "short",
                    day: "numeric",
                  })
                }}
              />
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    className="w-[150px]"
                    nameKey="views"
                    labelFormatter={(value) => {
                      return new Date(value).toLocaleDateString("en-US", {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      })
                    }}
                  />
                }
              />
              <Bar dataKey={activeChart} fill={`#FE5E34`} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}