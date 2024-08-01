"use client"

import React, { useEffect, useState } from "react"
import axios from "axios"
import { siteConfig } from "@/app/siteConfig"
import { useToast } from "@/components/ui/use-toast"
import cookies from "nookies"

export default function ConfigureWebPushCampaign({ onClose, existingConfigurations }) {
    const { toast } = useToast()

    const [title, setTitle] = useState("")
    const [message, setMessage] = useState("")
    const [url, setUrl] = useState("")
    const [icon, setIcon] = useState("")
    const [badge, setBadge] = useState("")
    const [showTestNotification, setShowTestNotification] = useState(false)

    const [accessToken, setAccessToken] = useState("")
    const [shopID, setShopId] = useState("")

    const [selectedConfiguration, setSelectedConfiguration] = useState("")

    const handleClose = () => {
        onClose()
    }

    useEffect(() => {
        let access_token = cookies.get(null).access_token
        if (access_token) {
            setAccessToken(access_token)
        }

        let userinfo = localStorage.getItem("userinfo")
        if (userinfo) {
            userinfo = JSON.parse(userinfo)
            setShopId(userinfo?.current_shop_id)
        }
    }, [])

    const handleConfigurationSelect = (e) => {
        const configId = e.target.value
        setSelectedConfiguration(configId)

        // Populate form with selected configuration data
        const config = existingConfigurations.find(c => c.id === configId)
        if (config) {
            setTitle(config.title || "")
            setMessage(config.message || "")
            setUrl(config.url || "")
            setIcon(config.icon || "")
            setBadge(config.badge || "")
        }
    }

    const convertToBase64 = (file) => {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result);
            reader.onerror = (error) => reject(error);
        });
    }

    const handleImageUpload = async (e, setImageFunction) => {
        const file = e.target.files[0];
        if (file) {
            try {
                const base64 = await convertToBase64(file);
                setImageFunction(base64);
            } catch (error) {
                console.error("Error converting image to base64:", error);
            }
        }
    }

    const handleCreateNewCopy = () => {
        toast({
            title: "Uploading new notification configuration...",
            description: "Please wait and do not click anything",
            variant: "info"
        })

        axios.post(`${siteConfig.baseApiUrl}/api/notification/private/notification-configuration`, {
            title,
            message,
            url,
            icon,
            badge,
            shop_id: shopID
        },
        {
            headers: {
                Authorization: `Bearer ${accessToken}`
            }
        }
        ).then(response => {
            console.log("New notification configuration created:", response.data);

            if (!response.data?.notification?.id) {
                toast({
                    title: "Error",
                    description: "Notification configuration ID not found in response",
                    variant: "destructive"
                })
                return;
            }

            localStorage.setItem("notification_configuration", response.data.notification.id);

            toast({
                title: "Success",
                description: "New notification configuration created successfully",
                variant: "success"
            })
            
            handleClose();
        }).catch(error => {
            let msg = "Error creating new notification configuration";

            console.error("Error creating new notification configuration:", error)

            if (error.response.data?.message) {
                msg += ": " + error.response.data.message;
            }

            toast({
                title: "Error",
                description: msg,
                variant: "destructive"
            })
        });
    }

    const handleUseThisOne = () => {
        if (selectedConfiguration) {
            localStorage.setItem("notification_configuration", selectedConfiguration);
            toast({
                title: "Success",
                description: "Selected notification configuration set for use",
                variant: "success"
            })
            handleClose();
        } else {
            toast({
                title: "Error",
                description: "Please select a configuration first",
                variant: "destructive"
            })
        }
    }

    const handleTest = () => {
        if ('Notification' in window) {
            Notification.requestPermission().then(permission => {
                if (permission === 'granted') {
                    const notification = new Notification(title || "", {
                        body: message || "",
                        icon: icon || undefined,
                        badge: badge || undefined,
                    });

                    if (url) {
                        console.log("Notification URL:", url);
                        notification.onclick = function (event) {
                            event.preventDefault(); // Prevent the browser from focusing the Notification's tab
                            window.open(url, '_blank');
                        }
                    }
                }
            });
        }
        setShowTestNotification(true)
        setTimeout(() => setShowTestNotification(false), 5000) // Hide after 5 seconds
    }

    return (
        <div className="max-w-md mx-auto mt-10 p-6 bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-gray-50">Configure Web Push Notification</h2>

            <div className="space-y-4">
                <select
                    value={selectedConfiguration}
                    onChange={handleConfigurationSelect}
                    className="w-full p-2 border rounded text-gray-900 dark:text-gray-50 bg-white dark:bg-gray-700"
                >
                    <option value="">Select existing configuration</option>
                    {existingConfigurations.map(config => (
                        <option key={config.id} value={config.id}>
                            {`${config.title} ${config.id}` || `Configuration ${config.id}`}
                        </option>
                    ))}
                </select>

                <input
                    type="text"
                    placeholder="Title"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    className="w-full p-2 border rounded text-gray-900 dark:text-gray-50 bg-white dark:bg-gray-700"
                />
                <textarea
                    placeholder="Message"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    className="w-full p-2 border rounded text-gray-900 dark:text-gray-50 bg-white dark:bg-gray-700"
                    rows="3"
                />
                <input
                    type="url"
                    placeholder="Redirect URL"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    className="w-full p-2 border rounded text-gray-900 dark:text-gray-50 bg-white dark:bg-gray-700"
                />
                <div>
                    <label className="block text-sm font-medium text-gray-900 dark:text-gray-50 mb-1">
                        Icon
                    </label>
                    <div className="flex items-center space-x-2">
                        <input
                            type="file"
                            accept="image/*"
                            onChange={(e) => handleImageUpload(e, setIcon)}
                            className="text-sm text-gray-900 dark:text-gray-50 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-gray-50 file:text-gray-700 hover:file:bg-gray-100"
                        />
                        {icon && <img src={icon} alt="Icon preview" className="w-10 h-10 object-cover" />}
                    </div>
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-900 dark:text-gray-50 mb-1">
                        Badge
                    </label>
                    <div className="flex items-center space-x-2">
                        <input
                            type="file"
                            accept="image/*"
                            onChange={(e) => handleImageUpload(e, setBadge)}
                            className="text-sm text-gray-900 dark:text-gray-50 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-gray-50 file:text-gray-700 hover:file:bg-gray-100"
                        />
                        {badge && <img src={badge} alt="Badge preview" className="w-10 h-10 object-cover" />}
                    </div>
                </div>
            </div>

            <div className="mt-6 space-x-2">
                <button
                    onClick={handleTest}
                    className="px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-900 dark:text-gray-50 rounded hover:bg-gray-300 dark:hover:bg-gray-500 transition"
                >
                    Test
                </button>

                <button
                    className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
                    onClick={handleCreateNewCopy}
                >
                    {/* Create a new copy */}
                    {selectedConfiguration ? "Create a new copy" : "Create new"}
                </button>

                {selectedConfiguration && (
                    <button
                        className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 transition"
                        onClick={handleUseThisOne}
                    >
                        Use
                    </button>
                )}

                <button
                    onClick={handleClose}
                    className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition"
                >
                    Close
                </button>
            </div>

            {showTestNotification && (
                <div className="fixed bottom-4 right-4 max-w-sm bg-white dark:bg-gray-800 border rounded-lg shadow-lg p-4 transition-opacity duration-500">
                    <div className="flex items-center">
                        {icon && <img src={icon} alt="Icon" className="w-10 h-10 mr-3 object-cover" />}
                        <div>
                            <h3 className="font-bold text-gray-900 dark:text-gray-50">{title || "Notification Title"}</h3>
                            <p className="text-sm text-gray-900 dark:text-gray-50">{message || "Notification message"}</p>
                        </div>
                    </div>
                    {url && (
                        <a href={url} target="_blank" rel="noopener noreferrer" className="text-blue-500 dark:text-blue-400 text-sm mt-2 block">
                            Click to open link
                        </a>
                    )}
                </div>
            )}
        </div>
    )
}