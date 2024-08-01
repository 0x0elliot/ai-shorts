"use client"

import * as React from "react"
import { TrendingUp } from "lucide-react"
import { Label, Pie, PieChart, Cell } from "recharts"
import { useToast } from "@/components/ui/use-toast"
import cookies from 'nookies';
import Link from 'next/link';
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import {
    ChartConfig,
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
} from "@/components/ui/chart"
import axios from "axios";
import { siteConfig } from "@/app/siteConfig";

const chartConfig = {
    visitors: { label: "Visitors" },
    linux: { label: "Linux" },
    macOS: { label: "Mac" },
    windows: { label: "Windows" },
    other: { label: "Other" },
} satisfies ChartConfig

export function DashboardClickerOS() {
    const { toast } = useToast()
    const [notificationCampaignData, setNotificationCampaignData] = React.useState([])
    const [chartData, setChartData] = React.useState([])
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
    
        let shopId = userinfo?.current_shop_id
        
        const fetchData = async () => {
            try {
                const [campaignResponse, osResponse] = await Promise.all([
                    axios.get(`${siteConfig.baseApiUrl}/api/notification/private/statistics/campaigns?shop_id=${shopId}`, {
                        headers: {
                            "Content-Type": "application/json",
                            "Authorization": `Bearer ${accessToken}`,
                        },
                    }),
                    axios.get(`${siteConfig.baseApiUrl}/api/tracking/private/stats/os?shop_id=${shopId}`, {
                        headers: {
                            "Content-Type": "application/json",
                            "Authorization": `Bearer ${accessToken}`,
                        },
                    })
                ]);

                setNotificationCampaignData(campaignResponse.data.stats);
                setChartData(osResponse.data.stats);
            } catch (error) {
                toast({
                    title: "Error",
                    description: error.response?.data?.message || "An error occurred",
                    variant: "destructive",
                })
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [])

    const totalVisitors = React.useMemo(() => {
        return chartData.reduce((acc, curr) => acc + curr.visitors, 0)
    }, [chartData])

    const NoDataDisplay = ({ message }) => (
        <div className="flex items-center justify-center h-[250px] rounded-lg mt-4 mb-4">
            <p className="text-gray-500 dark:text-gray-400">{message}</p>
        </div>
    )

    return (
        <Card className="flex flex-col">
            <CardHeader className="items-center pb-0">
                <CardTitle>About your clickers..</CardTitle>
                <CardDescription>All time data</CardDescription>
            </CardHeader>
            <div className="flex flex-row gap-4">
                <CardContent className="flex-1 pb-0">
                    {loading ? (
                        <div className="flex items-center justify-center h-[250px]">
                            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-gray-900"></div>
                        </div>
                    ) : chartData.length === 0 ? (
                        <NoDataDisplay message="No OS data available" />
                    ) : (
                        <ChartContainer
                            config={chartConfig}
                            className="mx-auto aspect-square max-h-[250px]"
                        >
                            <PieChart>
                                <ChartTooltip
                                    cursor={false}
                                    content={<ChartTooltipContent hideLabel />}
                                />
                                <Pie
                                    data={chartData}
                                    dataKey="visitors"
                                    nameKey="os"
                                    innerRadius={60}
                                    strokeWidth={5}
                                >
                                    <Label
                                        content={({ viewBox }) => {
                                            if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                                                return (
                                                    <text
                                                        x={viewBox.cx}
                                                        y={viewBox.cy}
                                                        textAnchor="middle"
                                                        dominantBaseline="middle"
                                                    >
                                                        <tspan
                                                            x={viewBox.cx}
                                                            y={viewBox.cy}
                                                            className="fill-foreground text-3xl font-bold"
                                                        >
                                                            {totalVisitors.toLocaleString()}
                                                        </tspan>
                                                        <tspan
                                                            x={viewBox.cx}
                                                            y={(viewBox.cy || 0) + 24}
                                                            className="fill-muted-foreground"
                                                        >
                                                            Visitors
                                                        </tspan>
                                                    </text>
                                                )
                                            }
                                        }}
                                    />
                                </Pie>
                            </PieChart>
                        </ChartContainer>
                    )}
                </CardContent>

                <CardContent className="flex-1 pb-0">
                    {loading ? (
                        <div className="flex items-center justify-center h-[250px]">
                            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-gray-900"></div>
                        </div>
                    ) : notificationCampaignData.length === 0 ? (
                        <NoDataDisplay message="No campaign data available" />
                    ) : (
                        <ChartContainer
                            config={chartConfig}
                            className="mx-auto aspect-square max-h-[250px]"
                        >
                            <PieChart>
                                <ChartTooltip
                                    cursor={false}
                                    content={<ChartTooltipContent hideLabel />}
                                />
                                <Pie
                                    data={notificationCampaignData}
                                    dataKey="visitors"
                                    nameKey="url"
                                    innerRadius={60}
                                    strokeWidth={5}
                                    onClick={(data) => {
                                        if (data && data.notification_campaign_id) {
                                            window.open(`/campaigns/history/${data.notification_campaign_id}/info`)
                                        }
                                    }}
                                >
                                    {notificationCampaignData.map((entry, index) => (
                                        <Cell key={`cell-${index}`} style={{ cursor: 'pointer' }} />
                                    ))}
                                    <Label
                                        content={({ viewBox }) => {
                                            if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                                                const totalCampaignVisitors = notificationCampaignData.reduce((sum, item) => sum + item.visitors, 0);
                                                return (
                                                    <text
                                                        x={viewBox.cx}
                                                        y={viewBox.cy}
                                                        textAnchor="middle"
                                                        dominantBaseline="middle"
                                                    >
                                                        <tspan
                                                            x={viewBox.cx}
                                                            y={viewBox.cy}
                                                            className="fill-foreground text-3xl font-bold"
                                                        >
                                                            {totalCampaignVisitors.toLocaleString()}
                                                        </tspan>
                                                        <tspan
                                                            x={viewBox.cx}
                                                            y={(viewBox.cy || 0) + 24}
                                                            className="fill-muted-foreground"
                                                        >
                                                            Campaign Visitors
                                                        </tspan>
                                                    </text>
                                                )
                                            }
                                        }}
                                    />
                                </Pie>
                            </PieChart>
                        </ChartContainer>
                    )}
                </CardContent>
            </div>
            <CardFooter className="flex-col gap-2 text-sm">
                <Link
                    href="https://calendly.com/aditya-zappush"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 font-medium leading-none text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 transition-colors"
                >
                    <div className="flex items-center gap-2">
                        Trend up more by talking to us <TrendingUp className="h-4 w-4" /> We probably can help
                    </div>
                </Link>
                <div className="leading-none text-muted-foreground">
                    Showing total visitors for the last 6 months
                </div>
            </CardFooter>
        </Card>
    )
}