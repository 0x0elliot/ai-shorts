"use client"
import { useState, useRef, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { Card, CardContent } from "@/components/ui/card"
import { useToast } from "@/components/ui/use-toast"
import { parseCookies } from 'nookies'
import { siteConfig } from '@/app/siteConfig'

export default function EditCreate() {
    const { toast } = useToast()
    const { id } = useParams()
    
    const [progress, setProgress] = useState(0)
    const [error, setError] = useState(null)
    const intervalRef = useRef(null)

    const determineProgress = (video) => {
        if (video.error) {
            setError(video.error)
            setProgress(0)
            return
        }
        if (video.videoUploaded) setProgress(100)
        else if (video.videoGenerated) setProgress(90)
        else if (video.videoStitched) setProgress(80)
        else if (video.ttsGenerated) setProgress(60)
        else if (video.dallePromptGenerated) setProgress(40)
        else if (video.scriptGenerated) setProgress(20)
        else if (video.script) setProgress(15)
        else setProgress(0)
    }

    useEffect(() => {
        const cookies = parseCookies()
        const access_token = cookies.access_token

        const fetchProgress = () => {
            // fetch(`/api/video/private/${id}`, {
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
                    description: 'Failed to get video progress'
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
            <section className="text-center mb-12">
                <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-50 mb-4">
                    Edit Your Video Magic
                </h1>
                <p className="text-xl text-gray-600 dark:text-gray-300">
                    Make changes to your video schedule
                </p>
            </section>
            <Card>
                <CardContent className="p-6">
                    {error ? (
                        <p className="text-red-500">Error: {error}</p>
                    ) : (
                        <div>
                            <p className="text-lg mb-2">Video Progress:</p>
                            <div className="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                                <div 
                                    className="bg-blue-600 h-2.5 rounded-full transition-all duration-500" 
                                    style={{width: `${progress}%`}}
                                ></div>
                            </div>
                            <p className="mt-2">{progress}% Complete</p>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}