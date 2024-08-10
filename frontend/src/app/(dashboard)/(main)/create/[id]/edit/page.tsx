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
import { CheckCircle2 } from 'lucide-react'
import { userInfo } from 'os'

const progressSteps = [
  { key: 'scriptGenerated', label: 'Script', slogan: 'Crafting a blockbuster script...' },
  { key: 'dallePromptGenerated', label: 'Image magic', slogan: 'Preparing to summon the AI art genies...' },
  { key: 'dalleGenerated', label: 'Images themselves', slogan: 'Summoning the AI art genies...' },
  { key: 'ttsGenerated', label: 'Fixing Timing', slogan: 'Teaching robots to talk like humans...' },
  { key: 'srtGenerated', label: 'Subtitles', slogan: 'Crafting subtitles for your masterpiece...' },
  { key: 'videoGenerated', label: 'Video Generated', slogan: 'Bringing your creation to life...' },
  { key: 'videoUploaded', label: 'Video Upload', slogan: 'Preparing your masterpiece for its debut...' },
]

const creativeSlogans = [
  "Lights, camera, AI-ction!",
  "Cooking up a visual feast...",
  "Sprinkling digital stardust...",
  "Painting with pixels and dreams...",
  "Turning 1s and 0s into pure magic...",
  "Mixing imagination and algorithms...",
  "Crafting tomorrow's memories today...",
  "Weaving a tapestry of digital wonder...",
]

export default function EditCreate() {
    const { toast } = useToast()
    const { id } = useParams()
    
    const [progress, setProgress] = useState(0)
    const [error, setError] = useState(null)
    const [status, setStatus] = useState('Brewing your video magic...')
    const [accessToken, setAccessToken] = useState("")
    const [canRecreate, setCanRecreate] = useState(false)
    const [createdAt, setCreatedAt] = useState(null)
    const [completedSteps, setCompletedSteps] = useState({})
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
            // setCanRecreate(false)
            setCompletedSteps({})
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
    
        const completed = {}
        progressSteps.forEach(step => {
            if (video[step.key]) {
                completed[step.key] = true
            }
        })
        setCompletedSteps(completed)
    
        // Use the progress directly from the backend
        setProgress(video.progress)
    
        if (video.progress === 100) {
            setStatus('Ta-da! Your video masterpiece is ready!')
            confetti({
                particleCount: 100,
                spread: 70,
                origin: { y: 0.6 }
            })
        } else {
            setStatus(creativeSlogans[Math.floor(Math.random() * creativeSlogans.length)])
        }
    
        // Check if video can be recreated
        const creationTime = new Date(video.CreatedAt)
        const now = new Date()
        const hoursSinceCreation = (now - creationTime) / (1000 * 60 * 60)
        setCanRecreate(hoursSinceCreation > 2 && video.progress < 100)
        setCreatedAt(video.CreatedAt)
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
                            <p className="text-2xl font-bold text-gray-800 dark:text-gray-200 mb-6">{progress}% Complete</p>
                            
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                                {progressSteps.map((step, index) => (
                                    <div key={index} className="flex flex-col items-center">
                                        <div className={`w-12 h-12 rounded-full flex items-center justify-center ${completedSteps[step.key] ? 'bg-green-500' : 'bg-gray-300'}`}>
                                            {completedSteps[step.key] ? <CheckCircle2 className="text-white" /> : (index + 1)}
                                        </div>
                                        <p className="mt-2 text-sm">{step.label}</p>
                                    </div>
                                ))}
                            </div>

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
                            {canRecreate && (
                                <motion.div
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    transition={{ delay: 0.5 }}
                                    className="mt-6"
                                >
                                    <p className="text-yellow-500 mb-2">Looks like our magic is taking a bit long. Want to try again?</p>
                                    <Button 
                                        onClick={handleRetry}
                                        className="bg-yellow-500 hover:bg-yellow-600 text-white font-bold py-2 px-4 rounded"
                                    >
                                        🔄 Recreate Video
                                    </Button>
                                </motion.div>
                            )}
                            {createdAt && (
                                <p className="mt-4 text-sm text-gray-500">
                                    Created at: {new Date(createdAt).toLocaleString()}
                                </p>
                            )}
                        </motion.div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}