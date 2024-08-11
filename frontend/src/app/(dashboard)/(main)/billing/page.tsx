"use client"
import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useToast } from "@/components/ui/use-toast"
import { parseCookies } from 'nookies'
import { siteConfig } from '@/app/siteConfig'
import { motion } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { CreditCard, Download, Check, ArrowRight, Calendar } from 'lucide-react'
import Script from 'next/script'

const pricingTiers = [
    { 
        name: "Free", 
        monthlyPrice: 0, 
        yearlyPrice: 0, 
        features: [ "Only 1 video allowed" , "Basic features", "Limited usage"]
    },
    { 
        name: "Starter", 
        monthlyPrice: 19, 
        yearlyPrice: 171, // 19 * 12 * 0.75 (25% discount)
        features: ["20 reels per month" ,"All Free features", "Access to all new features", "Priority support"]
    },
    { 
        name: "Pro", 
        monthlyPrice: 39, 
        yearlyPrice: 351, // 39 * 12 * 0.75
        features: ["40 reels per month", "All Starter features","Access to all new features", "Priority support"]
    },
    { 
        name: "Hustler", 
        monthlyPrice: 68, 
        yearlyPrice: 612, // 68 * 12 * 0.75
        features: ["70 reels per month", "All Pro features", "Dedicated account manager", "Custom integrations", "Access to all new features", "Priority support"]
    },
    { 
        name: "Big Player",
        monthlyPrice: 88, 
        yearlyPrice: 792, // 88 * 12 * 0.75
        features: ["100 reels per month", "All Business features", "24/7 premium support", "Access to all new features", "Priority support"]
    },
    {
        name: "Enterprise",
        monthlyPrice: 0,
        yearlyPrice: 0,
        features: ["Unlimited reels per month", "API support", "All Big Player features", "Custom pricing", "Dedicated account manager", "24/7 premium support", "Access to all new features", "Priority support"]
    }
]

export default function BillingPage() {
    const { toast } = useToast()
    const [accessToken, setAccessToken] = useState("")
    const [billingInfo, setBillingInfo] = useState(null)
    const [loading, setLoading] = useState(true)
    const [selectedTier, setSelectedTier] = useState(null)
    const [billingCycle, setBillingCycle] = useState('monthly')

    useEffect(() => {
        const cookies = parseCookies()
        const access_token = cookies.access_token
        setAccessToken(access_token)

        const fetchBillingInfo = async () => {
            try {
                const response = await fetch(`${siteConfig.baseApiUrl}/api/billing`, {
                    headers: {
                        'Authorization': `Bearer ${access_token}`
                    }
                })
                if (!response.ok) throw new Error('Failed to fetch billing info')
                const data = await response.json()
                setBillingInfo(data)
                setSelectedTier(pricingTiers.find(tier => tier.name === data.currentPlan) || pricingTiers[0])
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

    const handleUpgrade = async () => {
        if (!selectedTier || selectedTier.monthlyPrice === 0) return;

        try {
            const response = await fetch(`${siteConfig.baseApiUrl}/api/billing/create-order`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    tierName: selectedTier.name,
                    billingCycle: billingCycle
                })
            })
            if (!response.ok) throw new Error('Failed to create order')
            const data = await response.json()
            
            const options = {
                key: data.razorpayKeyId,
                amount: data.amount,
                currency: data.currency,
                name: "Your Company Name",
                description: `Upgrade to ${selectedTier.name} Plan (${billingCycle})`,
                order_id: data.orderId,
                handler: async function (response) {
                    try {
                        const verifyResponse = await fetch(`${siteConfig.baseApiUrl}/api/billing/verify-payment`, {
                            method: 'POST',
                            headers: {
                                'Authorization': `Bearer ${accessToken}`,
                                'Content-Type': 'application/json'
                            },
                            body: JSON.stringify({
                                razorpay_payment_id: response.razorpay_payment_id,
                                razorpay_order_id: response.razorpay_order_id,
                                razorpay_signature: response.razorpay_signature
                            })
                        })
                        if (!verifyResponse.ok) throw new Error('Failed to verify payment')
                        const verifyData = await verifyResponse.json()
                        toast({
                            title: "Success",
                            description: `Your plan has been upgraded to ${selectedTier.name} (${billingCycle})!`,
                        })
                        setBillingInfo(verifyData.billingInfo)
                    } catch (error) {
                        console.error('Error verifying payment:', error)
                        toast({
                            variant: 'destructive',
                            title: 'Error',
                            description: 'Failed to verify payment. Please contact support.'
                        })
                    }
                },
                prefill: {
                    name: billingInfo?.name,
                    email: billingInfo?.email,
                    contact: billingInfo?.phone
                },
                theme: {
                    color: "#3399cc"
                }
            };
            const paymentObject = new window.Razorpay(options)
            paymentObject.open()
        } catch (error) {
            console.error('Error creating order:', error)
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

    if (loading) {
        return <div className="flex justify-center items-center h-screen">Loading...</div>
    }

    return (
        <>
            <Script src="https://checkout.razorpay.com/v1/checkout.js" />
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
                                Yearly (25% off)
                            </Button>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {pricingTiers.map((tier, index) => (
                                <Card key={index} className={`p-6 ${selectedTier === tier ? 'border-blue-500 border-2' : ''}`}>
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
                                        <><Button
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
                                            </Button></>
                                    ) : (
                                        <Button 
                                            onClick={() => setSelectedTier(tier)} 
                                            className={`w-full ${selectedTier === tier ? 'bg-blue-500' : 'bg-gray-200 text-gray-800'}`}
                                        >
                                            {selectedTier === tier ? 'Selected' : 'Select'}
                                        </Button>
                                    )}
                                </Card>
                            ))}
                        </div>
                    </CardContent>
                </Card>

                {selectedTier && selectedTier.monthlyPrice > 0 && (
                    <Card>
                        <CardHeader>
                            <CardTitle>Upgrade to {selectedTier.name}</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-center">
                                <p className="text-2xl font-bold mb-2">
                                    ${billingCycle === 'monthly' ? selectedTier.monthlyPrice : (selectedTier.yearlyPrice / 12).toFixed(2)}/month
                                </p>
                                {billingCycle === 'yearly' && (
                                    <p className="text-green-500 mb-4">Billed annually at ${selectedTier.yearlyPrice}/year (Save 25%)</p>
                                )}
                                <Button onClick={handleUpgrade} className="mt-4 bg-green-500 hover:bg-green-600">
                                    Upgrade to {selectedTier.name} ({billingCycle}) <ArrowRight className="ml-2" size={16} />
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                )}

                <Card>
                    <CardHeader>
                        <CardTitle>Current Plan: {billingInfo?.currentPlan}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xl mb-4">Your current price: ${billingInfo?.currentPrice}/{billingInfo?.billingCycle}</p>
                        <h3 className="font-semibold mb-2">Features:</h3>
                        <ul className="list-disc list-inside mb-4">
                            {billingInfo?.features.map((feature, index) => (
                                <li key={index} className="flex items-center">
                                    <Check className="mr-2 text-green-500" size={16} />
                                    {feature}
                                </li>
                            ))}
                        </ul>
                    </CardContent>
                </Card>

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
                                {billingInfo?.invoices.map((invoice) => (
                                    <tr key={invoice.id}>
                                        <td>{new Date(invoice.date).toLocaleDateString()}</td>
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
            </div>
        </>
    )
}