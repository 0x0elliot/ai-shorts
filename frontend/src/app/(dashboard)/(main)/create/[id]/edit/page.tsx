"use client"
import { useState, useRef, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Card, CardContent } from "@/components/ui/card"
import { useToast } from "@/components/ui/use-toast"
import { parseCookies } from 'nookies'
import { siteConfig } from '@/app/siteConfig'
import { motion } from 'framer-motion'
import confetti from 'canvas-confetti'
import { Button } from '@/components/ui/button'

export default function EditCreate() {
    const { toast } = useToast()
    const { id } = useParams()
    
    const [progress, setProgress] = useState(0)
    const [error, setError] = useState(null)
    const [status, setStatus] = useState('Brewing your video magic...')
    const [accessToken, setAccessToken] = useState("")
    const intervalRef = useRef(null)

    const handleRetry = () => {
        fetch(`${siteConfig.baseApiUrl}/api/video/private/recreate/${id}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        })
        .then(res => {
            if (!res.ok) throw new Error('Failed to recreate video')
            return res.json()
        })
        .then(data => {
            setError(null)
            setProgress(0)
            setStatus('Restarting the magic! Hang tight...')
            toast({
                title: "Success",
                description: "Video recreation started. The show must go on!",
            })

            // reload page after 2 seconds
            setTimeout(() => {
                window.location.reload()
            }, 2000)
        })
        .catch(err => {
            console.error('Error recreating video:', err)
            toast({
                variant: 'destructive',
                title: 'Error',
                description: 'Failed to restart the video creation. Our wand needs new batteries!'
            })
        })
    }

    const determineProgress = (video) => {
        if (video.error) {
            setError(video.error)
            setProgress(0)
            setStatus('Oops! Our magic wand misfired!')
            return
        }
        if (video.videoUploaded) {
            setProgress(100)
            setStatus('Ta-da! Your video masterpiece is ready!')
            confetti({
                particleCount: 100,
                spread: 70,
                origin: { y: 0.6 }
            })
        } else if (video.dalleGenerated) { setProgress(60); setStatus('Summoning the AI art genies...') }
        else if (video.srtGenerated) { setProgress(40); setStatus('Crafting subtitles for your masterpiece...') }
        else if (video.ttsGenerated) { setProgress(30); setStatus('Teaching robots to talk like humans...') }
        else if (video.dallePromptGenerated) { setProgress(45); setStatus('Preparing to summon the AI art genies...') }
        else if (video.scriptGenerated) { setProgress(10); setStatus('Crafting a blockbuster script...') }
        else { setProgress(0); setStatus('Warming up our creative engines...') }
    }

    useEffect(() => {
        const cookies = parseCookies()
        const access_token = cookies.access_token
        setAccessToken(access_token)
        const fetchProgress = () => {
            fetch(`${siteConfig.baseApiUrl}/api/video/private/${id}`, {
                headers: {
                    'Authorization': `Bearer ${access_token}`
                }
            })
            .then(res => {
                if (!res.ok) throw new Error('Failed to fetch video progress')
                return res.json()
            })
            .then(data => {
                determineProgress(data.video)
            })
            .catch(err => {
                console.error('Error fetching video progress:', err)
                toast({
                    variant: 'destructive',
                    title: 'Error',
                    description: 'Our crystal ball is foggy. Retry in a bit!'
                })
            })
        }
        fetchProgress() // Initial fetch
        intervalRef.current = setInterval(fetchProgress, 5000)
        return () => {
            if (intervalRef.current) clearInterval(intervalRef.current)
        }
    }, [id, toast])

    return (
        <div className="max-w-4xl mx-auto p-6 space-y-8">
            <motion.section 
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="text-center mb-12"
            >
                <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-50 mb-4">
                    🎬 Lights, Camera, AI-ction!
                </h1>
                <p className="text-xl text-gray-600 dark:text-gray-300">
                    Your video is in the making. Grab some popcorn!
                </p>
            </motion.section>
            <Card>
                <CardContent className="p-6">
                    {error ? (
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            className="text-center"
                        >
                            <p className="text-red-500 text-lg mb-4">Whoops! We hit a snag in our magic trick:</p>
                            <p className="text-red-400">{error}</p>
                            <p className="mt-4 text-gray-600">Don't worry, our team of wizard debuggers is on it!</p>
                            <Button 
                                onClick={handleRetry}
                                className="mt-6 bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded"
                            >
                                🔄 Wave the Magic Wand Again
                            </Button>
                        </motion.div>
                    ) : (
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            className="text-center"
                        >
                            <p className="text-lg mb-4 font-semibold text-blue-600 dark:text-blue-400">{status}</p>
                            <div className="w-full bg-gray-200 rounded-full h-4 dark:bg-gray-700 mb-4">
                                <motion.div 
                                    className="bg-blue-600 h-4 rounded-full transition-all duration-500" 
                                    style={{width: `${progress}%`}}
                                    initial={{ width: 0 }}
                                    animate={{ width: `${progress}%` }}
                                />
                            </div>
                            <p className="text-2xl font-bold text-gray-800 dark:text-gray-200">{progress}% Complete</p>
                            {progress === 100 && (
                                <motion.p
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ delay: 0.5 }}
                                    className="mt-4 text-green-500 font-semibold"
                                >
                                    🎉 Bravo! Your video is ready for its debut!
                                </motion.p>
                            )}
                        </motion.div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}