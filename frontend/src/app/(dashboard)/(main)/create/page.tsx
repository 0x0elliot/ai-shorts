"use client"
import { useState, useRef } from 'react'
import { siteConfig } from "@/app/siteConfig"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { Card, CardContent } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Button } from "@/components/ui/button"
import { Icon } from '@iconify/react'
import { toast } from "@/components/ui/use-toast"

export default function Create() {
    const [topic, setTopic] = useState('')
    const [description, setDescription] = useState('')
    const [voice, setVoice] = useState('p251') // Edward as default
    const [videoStyle, setVideoStyle] = useState('default')
    const [postingSchedule, setPostingSchedule] = useState(['email']) // Email me as default
    const [isOneTime, setIsOneTime] = useState(true) // One time video on by default
    const [selectedStyle, setSelectedStyle] = useState('default')
    const audioRef = useRef(null)

    const videoStyles = [
        { name: 'Default', description: 'Clean, modern look with vibrant colors' },
        { name: 'Anime', description: 'Japanese animation style with vibrant colors' },
        { name: 'Watercolor', description: 'Soft, dreamy look with gentle color blends' },
        { name: 'Cartoon', description: 'Fun and playful animated style' },
    ]

    const narrators = [
        { name: "Edward", value: "p230", description: "Friendly and approachable" },
        { name: "Elena", value: "p248", description: "Warm and professional" },
        { name: "Oliver", value: "p251", description: "Charming British accent" },
        { name: "James", value: "p254", description: "Deep and authoritative" },
        { name: "William", value: "p256", description: "Refined British tone" },
        { name: "Charlotte", value: "p260", description: "Elegant British accent" },
        { name: "Sophia", value: "p263", description: "Energetic and engaging" },
        { name: "Michael", value: "p264", description: "Rich and resonant" },
        { name: "Daniel", value: "p267", description: "Serious and professional" },
        { name: "Emily", value: "p273", description: "Polished British accent" },
        { name: "Thomas", value: "p282", description: "Soothing and calming" },
        { name: "Priya", value: "p345", description: "Warm with a subtle accent" }
    ]

    const scheduleOptions = [
        { id: 'email', label: 'Email me', icon: 'ph:envelope-simple' },
        { id: 'youtube', label: 'Post as YouTube Short', icon: 'ph:youtube-logo' },
        { id: 'tiktok', label: 'Post as TikTok Short', icon: 'ph:tiktok-logo' },
        { id: 'instagram', label: 'Post as Instagram Short', icon: 'ph:instagram-logo' },
    ]

    const handleScheduleChange = (id) => {
        setPostingSchedule(prev =>
            prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id]
        )
    }

    const playVoiceDemo = () => {
        if (audioRef.current) {
            audioRef.current.pause()
            audioRef.current.currentTime = 0
            const selectedNarrator = narrators.find(n => n.value === voice)
            if (selectedNarrator) {
                const audioPath = `/audio/${selectedNarrator.name}.wav`
                audioRef.current.src = audioPath
                audioRef.current.play()
            }
        }
    }


    const handleSubmit = () => {
        if (!topic || !description || !voice || !videoStyle || postingSchedule.length === 0) {
            toast({
                title: "Error",
                description: "Please fill in all fields before submitting.",
                variant: "destructive",
            })
            return
        }

        const requestData = {
            topic,
            description,
            narrator: voice.toLowerCase(),
            videoStyle: videoStyle.toLowerCase(),
            postingSchedule,
            isOneTime
        }

        


        console.log('Request Data:', requestData)
        // Here you would typically send this data to your API
        toast({
            title: "Success",
            description: "Your video schedule has been created!",
        })
    }

    return (
        <div className="max-w-4xl mx-auto p-6 space-y-8">
            <section className="text-center mb-12">
                <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-50 mb-4">
                    Create Your Video Magic
                </h1>
                <p className="text-xl text-gray-600 dark:text-gray-300">
                    Transform your ideas into engaging social media content in just a few clicks!
                </p>
            </section>

            <Card className="shadow-lg">
                <CardContent className="p-8 space-y-8">
                    <div className="space-y-4">
                        <Label htmlFor="topic" className="text-xl font-semibold flex items-center gap-2">
                            <Icon icon="ph:lightbulb" className="w-6 h-6" />
                            What's your brilliant idea?
                        </Label>
                        <Input
                            id="topic"
                            value={topic}
                            onChange={(e) => setTopic(e.target.value)}
                            placeholder="Enter your video topic or idea"
                            className="text-lg p-4"
                            required
                        />
                    </div>

                    <div className="space-y-4">
                        <Label htmlFor="description" className="text-xl font-semibold flex items-center gap-2">
                            <Icon icon="ph:text-t" className="w-6 h-6" />
                            Describe your vision (max 5000 words)
                        </Label>
                        <Textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            placeholder="Paint a picture with words..."
                            maxLength={5000}
                            className="text-lg p-4 h-40"
                            required
                        />
                    </div>

                    <div className="space-y-4">
                        <Label htmlFor="voice" className="text-xl font-semibold flex items-center gap-2">
                            <Icon icon="ph:microphone" className="w-6 h-6" />
                            Choose your narrator
                        </Label>
                        <div className="flex items-center gap-4">
                            <Select onValueChange={setVoice} value={voice}>
                                <SelectTrigger className="text-lg p-4 flex-grow">
                                    <SelectValue placeholder="Select a voice" />
                                </SelectTrigger>
                                <SelectContent>
                                    {narrators.map((narrator) => (
                                        <SelectItem key={narrator.value} value={narrator.value}>
                                            {narrator.name} - {narrator.description}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <Button onClick={playVoiceDemo} disabled={!voice} className="whitespace-nowrap">
                                <Icon icon="ph:play" className="w-6 h-6 mr-2" />
                                Play Sample
                            </Button>
                        </div>
                        <audio ref={audioRef} />
                    </div>


                    <div className="space-y-4">
                        <Label className="text-xl font-semibold flex items-center gap-2">
                            <Icon icon="ph:palette" className="w-6 h-6" />
                            Pick your style of background imagery
                        </Label>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            {videoStyles.map((style) => (
                                <div
                                    key={style.name.toLowerCase()}
                                    className={`relative p-4 border-2 rounded-lg cursor-pointer transition-all ${selectedStyle === style.name.toLowerCase()
                                            ? 'border-blue-500 bg-blue-100 dark:bg-blue-900'
                                            : 'border-gray-200 bg-white dark:bg-gray-800 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
                                        }`}
                                    onClick={() => {
                                        setSelectedStyle(style.name.toLowerCase())
                                        setVideoStyle(style.name.toLowerCase())
                                    }}
                                >
                                    <div className="flex flex-col items-center justify-center text-center h-full">
                                        <span className={`font-medium text-lg mb-2 ${selectedStyle === style.name.toLowerCase()
                                                ? 'text-blue-600 dark:text-blue-400'
                                                : ''
                                            }`}>
                                            {style.name}
                                        </span>
                                        <span className={`text-sm ${selectedStyle === style.name.toLowerCase()
                                                ? 'text-blue-600 dark:text-blue-400'
                                                : 'text-gray-500 dark:text-gray-400'
                                            }`}>
                                            {style.description}
                                        </span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="space-y-4">
                        <Label className="text-xl font-semibold flex items-center gap-2">
                            <Icon icon="ph:share-network" className="w-6 h-6" />
                            Where should we share your masterpiece?
                        </Label>
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                            {scheduleOptions.map((option) => (
                                <div key={option.id} className="flex items-center space-x-2 bg-white dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700">
                                    <Checkbox
                                        id={option.id}
                                        checked={postingSchedule.includes(option.id)}
                                        onCheckedChange={() => handleScheduleChange(option.id)}
                                        className="h-5 w-5"
                                    />
                                    <Label htmlFor={option.id} className="flex items-center gap-2 cursor-pointer">
                                        <Icon icon={option.icon} className="w-6 h-6" />
                                        {option.label}
                                    </Label>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="flex items-center space-x-2">
                        <Checkbox
                            id="one-time"
                            checked={isOneTime}
                            onCheckedChange={setIsOneTime}
                            className="h-5 w-5"
                        />
                        <Label htmlFor="one-time" className="text-lg cursor-pointer">
                            One-time video (not a recurring schedule)
                        </Label>
                    </div>

                    <Button className="w-full text-lg py-6" size="lg" onClick={handleSubmit}>
                        Create My Video Schedule
                    </Button>
                </CardContent>
            </Card>
        </div>
    )
}