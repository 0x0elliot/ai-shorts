"use client"

import React, { useState, useEffect } from "react";
import { Button } from "@/components/ui/button"
import { siteConfig } from "@/app/siteConfig";

import cookies from 'nookies';

export default function WebPushNotificationsCampaignSettings() {
    const [testNotification, setTestNotification] = useState(false);
    const [permission, setPermission] = useState(null);
    const [accessToken, setAccessToken] = useState("");
    const [requestShopSetup, setRequestShopSetup] = useState(true);

    useEffect(() => {
        setAccessToken(cookies.get(null).access_token);

        if (typeof Notification !== 'undefined') {
            setPermission(Notification.permission);
        }

        // get userinfo from localsotrage
        try {
            let userinfoJSON = JSON.parse(localStorage.getItem('userinfo') || '{}');

            if (userinfoJSON.current_shop_id?.length > 1) {
                setRequestShopSetup(false);
            }
        } catch (error) {
            console.error("Error getting userinfo from localstorage:", error);
        }
    }, [])

    useEffect(() => {
        if (typeof Notification !== 'undefined') {
            if (testNotification && Notification.permission === 'granted') {
                // fetch("/api/send-test-notification", {
                fetch(`${siteConfig.baseApiUrl}/api/notification/private/push`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        "Authorization": `Bearer ${accessToken}`,
                    },
                    body: JSON.stringify({
                        "test": true,
                    }),
                })
                    .then(response => response.json())
                    .then(data => {
                        console.log("Test notification sent:", data);
                        setTestNotification(false);
                    })
                    .catch(error => {
                        console.error("Error sending test notification:", error);
                        setTestNotification(false);
                    });
            }
        }
    }, [testNotification, permission]);

    const requestPermission = async () => {
        if (requestShopSetup) {
            window.location.href = "/dashboard";
            return
        }

        const result = await Notification.requestPermission();
        console.log("Permission:", result);
        // setPermission(result);
        // if (result === 'granted') {
        //     await registerServiceWorker();
        // }

        await registerServiceWorker();
    };

    const registerServiceWorker = async () => {
        try {
            // const registration = await navigator.serviceWorker.register('/service-worker.js');
            const registration = await navigator.serviceWorker.register(`${siteConfig.baseApiUrl}/public/service-worker.js`);
            const subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: 'BCv7WgVIIGsZfgamKaruQEach2j6a8Us5en7Y2FIuC7PUt9aQxd2Nl2d5XIj80cfgs37DA6OE3TS1GOebJs0UTo'
            });

            await fetch(`${siteConfig.baseApiUrl}/api/notification/private/subscribe`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${accessToken}`,
                },
                body: JSON.stringify(subscription),
            }).then(response => response.json())
            .then(data => {
                setPermission('granted');
            })
        } catch (error) {
            console.error("Error registering service worker:", error);
        }
    };

    return (
        <div className="flex bg-white dark:bg-gray-900">
            <div className="bg-white dark:bg-gray-900 p-8 h-64 rounded-lg max-w-md w-full">
                <div className="text-sm font-medium text-gray-900 dark:text-gray-50 mb-6">
                    <p>Click the button below to allow web push notifications</p>
                </div>

                <div className="flex flex-col space-y-4">
                    <Button 
                        onClick={requestPermission} 
                        disabled={permission === 'granted' && !requestShopSetup}
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                    >
                        {/* {permission === 'granted' ? 'Permission Granted' : 'Allow Web Push Notifications'} */}
                        {requestShopSetup ? 'Setup Shop First' : permission === 'granted' ? 'Permission Granted' : 'Allow Web Push Notifications'}
                    </Button>

                    <Button 
                        onClick={() => setTestNotification(true)} 
                        disabled={permission !== 'granted'}
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                    >
                        Test Notification
                    </Button>

                    {testNotification && <p className="text-sm text-gray-600 dark:text-gray-400">Sending test notification...</p>}
                </div>
            </div>
        </div>
    );
}
