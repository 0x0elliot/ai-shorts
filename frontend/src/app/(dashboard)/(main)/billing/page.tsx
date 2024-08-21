"use client"

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useToast } from "@/components/ui/use-toast"
import { parseCookies } from 'nookies'
import { siteConfig } from '@/app/siteConfig'
import { motion } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { CreditCard, Download, Check, ArrowRight, Calendar } from 'lucide-react'

const pricingTiers = [
    {
        name: "Free",
        monthlyPlanId: null,
        yearlyPlanId: null,
        monthlyPrice: 0,
        yearlyPrice: 0,
        features: ["Only 1 video allowed", "Basic features", "Limited usage"]
    },
    {
        name: "Basic",
        monthlyPlanId: "336427",
        yearlyPlanId: "336436",
        monthlyPrice: 10,
        yearlyPrice: 102,
        features: ["5 reels per month", "All Free features", "Access to all new features", "Priority support"]
    },
    {
        name: "Standard",
        monthlyPlanId: "336421",
        yearlyPlanId: "336437",
        monthlyPrice: 19,
        yearlyPrice: 193.80,
        features: ["20 reels per month", "All Basic features", "Access to all new features", "Priority support"]
    },
    {
        name: "Pro",
        monthlyPlanId: "336428",
        yearlyPlanId: "336438",
        monthlyPrice: 39,
        yearlyPrice: 397.80,
        features: ["40 reels per month", "All Standard features", "Access to all new features", "Priority support"]
    },
    {
        name: "Premium",
        monthlyPlanId: "336432",
        yearlyPlanId: "336439",
        monthlyPrice: 69,
        yearlyPrice: 703.80,
        features: ["100 reels per month", "All Pro features", "24/7 premium support", "Access to all new features", "Priority support"]
    },
    {
        name: "Enterprise",
        monthlyPlanId: null,
        yearlyPlanId: null,
        monthlyPrice: 0,
        yearlyPrice: 0,
        features: ["Unlimited reels per month", "API support", "All Premium features", "Custom pricing", "Dedicated account manager", "24/7 premium support", "Access to all new features", "Priority support"]
    }
]

export default function BillingPage() {
    const { toast } = useToast()
    const [accessToken, setAccessToken] = useState("")
    const [billingInfo, setBillingInfo] = useState(null)
    const [loading, setLoading] = useState(true)
    const [billingCycle, setBillingCycle] = useState('monthly')

    useEffect(() => {
        const cookies = parseCookies()
        const access_token = cookies.access_token
        setAccessToken(access_token)

        const fetchBillingInfo = async () => {
            try {
                const response = await fetch(`${siteConfig.baseApiUrl}/api/billing/private/current-plan`, {
                    headers: {
                        'Authorization': `Bearer ${access_token}`
                    }
                })
                if (!response.ok) throw new Error('Failed to fetch billing info')
                const data = await response.json()
                setBillingInfo(data.subscription)
            } catch (error) {
                console.error('Error fetching billing info:', error)
                toast({
                    variant: 'destructive',
                    title: 'Error',
                    description: 'Failed to load billing information. Please try again.'
                })
            } finally {
                setLoading(false)
            }
        }

        fetchBillingInfo()
    }, [toast])

    const handleEnterpriseClick = () => {
        window.open('https://your-calendar-booking-link.com', '_blank')
    }

    const handleUpgrade = async (tier) => {
        if (tier.monthlyPrice === 0) return;

        const planId = billingCycle === 'monthly' ? tier.monthlyPlanId : tier.yearlyPlanId;

        try {
            const response = await fetch(`${siteConfig.baseApiUrl}/api/billing/private/create-checkout`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    plan_id: planId,
                    email: billingInfo?.email
                })
            })
            if (!response.ok) throw new Error('Failed to create checkout')
            const data = await response.json()
            
            // Redirect to LemonSqueezy checkout
            window.location.href = data.data.checkout_url
        } catch (error) {
            console.error('Error creating checkout:', error)
            toast({
                variant: 'destructive',
                title: 'Error',
                description: 'Failed to initiate upgrade. Please try again.'
            })
        }
    }

    const handleDownloadInvoice = async (invoiceId) => {
        try {
            const response = await fetch(`${siteConfig.baseApiUrl}/api/billing/invoice/${invoiceId}`, {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            })
            if (!response.ok) throw new Error('Failed to download invoice')
            const blob = await response.blob()
            const url = window.URL.createObjectURL(blob)
            const link = document.createElement('a')
            link.href = url
            link.download = `invoice_${invoiceId}.pdf`
            document.body.appendChild(link)
            link.click()
            window.URL.revokeObjectURL(url)
            document.body.removeChild(link)
        } catch (error) {
            console.error('Error downloading invoice:', error)
            toast({
                variant: 'destructive',
                title: 'Error',
                description: 'Failed to download invoice. Please try again.'
            })
        }
    }

    const getCurrentPlanDetails = () => {
        if (!billingInfo?.lemon_squeezy_id) return null;
        
        const currentPlan = pricingTiers.find(tier => 
            tier.monthlyPlanId === billingInfo.lemon_squeezy_id || 
            tier.yearlyPlanId === billingInfo.lemon_squeezy_id
        );

        if (!currentPlan) return null;

        const isYearly = currentPlan.yearlyPlanId === billingInfo.lemon_squeezy_id;
        const price = isYearly ? currentPlan.yearlyPrice : currentPlan.monthlyPrice;
        const cycle = isYearly ? 'year' : 'month';

        return {
            name: currentPlan.name,
            price: price,
            cycle: cycle,
            features: currentPlan.features
        };
    }

    if (loading) {
        return <div className="flex justify-center items-center h-screen">Loading...</div>
    }

    const currentPlan = getCurrentPlanDetails();

    return (
        <div className="max-w-6xl mx-auto p-6 space-y-8">
            <motion.section 
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="text-center mb-12"
            >
                <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-50 mb-4">
                    💳 Billing & Subscription
                </h1>
                <p className="text-xl text-gray-600 dark:text-gray-300">
                    Choose the perfect plan for your needs
                </p>
            </motion.section>

            <Card>
                <CardHeader>
                    <CardTitle>Select Your Plan</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex justify-center mb-6">
                        <Button
                            onClick={() => setBillingCycle('monthly')}
                            className={`mr-2 ${billingCycle === 'monthly' ? 'bg-blue-500' : 'bg-gray-200 text-gray-800'}`}
                        >
                            Monthly
                        </Button>
                        <Button
                            onClick={() => setBillingCycle('yearly')}
                            className={`${billingCycle === 'yearly' ? 'bg-blue-500' : 'bg-gray-200 text-gray-800'}`}
                        >
                            Yearly (15% off)
                        </Button>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {pricingTiers.map((tier, index) => (
                            <Card key={index} className="p-6">
                                <h3 className="text-2xl font-bold mb-2">{tier.name}</h3>
                                {tier.name.toLowerCase() === "enterprise" ? (
                                    <p className="text-3xl font-bold mb-4">Get in touch</p>
                                ) : (
                                    <p className="text-3xl font-bold mb-4">
                                        ${billingCycle === 'monthly' ? tier.monthlyPrice : (tier.yearlyPrice / 12).toFixed(2)}
                                        <span className="text-sm font-normal">/month</span>
                                    </p>
                                )}
                                {billingCycle === 'yearly' && tier.name.toLowerCase() !== "enterprise" && (
                                    <p className="text-green-500 mb-4">Billed annually at ${tier.yearlyPrice}/year</p>
                                )}
                                <ul className="mb-4">
                                    {tier.features.map((feature, i) => (
                                        <li key={i} className="flex items-center mb-2">
                                            <Check className="mr-2 text-green-500" size={16} />
                                            {feature}
                                        </li>
                                    ))}
                                </ul>
                                {tier.name.toLowerCase() === "enterprise" ? (
                                    <>
                                        <Button
                                            onClick={handleEnterpriseClick}
                                            className="w-full bg-blue-500 hover:bg-blue-600"
                                        >
                                            <Calendar className="mr-2" size={16} />
                                            Schedule a Call
                                        </Button>
                                        <Button
                                            onClick={() => window.open('mailto:support@reelrocket.ai')}
                                            className="w-full bg-blue-500 hover:bg-blue-600 mt-2"
                                        >
                                            Email Us
                                        </Button>
                                    </>
                                ) : (
                                    <Button 
                                        onClick={() => handleUpgrade(tier)} 
                                        className="w-full bg-green-500 hover:bg-green-600"
                                        disabled={currentPlan?.name === tier.name || tier.monthlyPrice === 0}
                                    >
                                        {currentPlan?.name === tier.name ? 'Current Plan' : `Upgrade to ${tier.name}`}
                                        {currentPlan?.name !== tier.name && <ArrowRight className="ml-2" size={16} />}
                                    </Button>
                                )}
                            </Card>
                        ))}
                    </div>
                </CardContent>
            </Card>

            {currentPlan && (
                <Card>
                    <CardHeader>
                        <CardTitle>Current Plan: {currentPlan.name}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xl mb-4">Your current price: ${currentPlan.price}/{currentPlan.cycle}</p>
                        <h3 className="font-semibold mb-2">Features:</h3>
                        <ul className="list-disc list-inside mb-4">
                            {currentPlan.features.map((feature, index) => (
                                <li key={index} className="flex items-center">
                                    <Check className="mr-2 text-green-500" size={16} />
                                    {feature}
                                </li>
                            ))}
                        </ul>
                    </CardContent>
                </Card>
            )}

            {billingInfo && billingInfo.invoices && billingInfo.invoices.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle>Billing History</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <table className="w-full">
                            <thead>
                                <tr>
                                    <th className="text-left">Date</th>
                                    <th className="text-left">Amount</th>
                                    <th className="text-left">Status</th>
                                    <th className="text-left">Invoice</th>
                                </tr>
                            </thead>
                            <tbody>
                                {billingInfo.invoices.map((invoice) => (
                                    <tr key={invoice.id}>
                                        <td>{new Date(invoice.paid_at).toLocaleDateString()}</td>
                                        <td>${invoice.amount}</td>
                                        <td>{invoice.status}</td>
                                        <td>
                                            <Button onClick={() => handleDownloadInvoice(invoice.id)} variant="outline" size="sm">
                                                <Download size={16} className="mr-2" />
                                                Download
                                            </Button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </CardContent>
                </Card>
            )}
        </div>
    )
}