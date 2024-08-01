"use client"

import React, { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button"
import Link from 'next/link';

import cookies from 'nookies';
import { siteConfig } from '@/app/siteConfig';

export default function WebPushNotificationThemeSettings(props) {
    const [notifications, setNotifications] = useState([]);
    const [notificationEnabled, setNotificationEnabled] = useState(false);
    const storeName = props.storeName;
    const [accessToken, setAccessToken] = useState(null);

    useEffect(() => {
        setAccessToken(cookies.get(null).access_token);
        let accessToken_ = cookies.get(null).access_token;

        if (storeName === null || storeName === undefined) {
            return;
        }

        fetch(`${siteConfig.baseApiUrl}/api/notification/private/notifications?shop_identifier=${storeName}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                "Authorization": `Bearer ${accessToken_}`
            }
        })
            .then(response => response.json())
            .then(data => {
                setNotifications(data.notifications);
                if (data.notifications.length > 1) {
                    // this means the user has enabled push notifications for their store
                    // later, we filter by "push" to check if the user has enabled push notifications
                    setNotificationEnabled(true);
                }
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }, [storeName]);

    const handleEnableNotifications = () => {

        fetch(`${siteConfig.baseApiUrl}/api/notification/private/enable/push-notifications`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                "Authorization": `Bearer ${accessToken}`
            },
            body: JSON.stringify({
                shop_identifier: storeName
            })
        })
            .then(response => response.json())
            .then(data => {
                setNotificationEnabled(true);
            });
    };

    const handleConfigure = () => {
        let storeName_;
        storeName_ = storeName;
        if (storeName?.includes('.myshopify.com')) {
            storeName_ = storeName.split('.myshopify.com')[0];
        }

        return `https://${storeName_}.myshopify.com/admin/themes/current/editor?context=apps&activateAppId=1900f709-83df-43f3-b52f-4ec0ad27798a%2Ffloating_logic`
    }

    return (
        <div className="flex bg-white dark:bg-gray-900">
            <div className="flex-1 p-4">
                <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-50">Web Push Notifications</h1>
                <p className="text-sm text-gray-600 dark:text-gray-400">Configure your web push notifications</p>

                <div className="flex justify-center p-4 mt-10 space-x-4">
                    <Button
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                    >
                        {!notificationEnabled ? (
                            <Link href={handleConfigure()} target="_blank" onClick={handleEnableNotifications}>
                                Configure On Shopify
                            </Link>
                        ) : (
                            // Come back to this when you can auto detect if shopify store has been set up or not
                            <Link href={handleConfigure()} target="_blank">
                                Configure on Shopify
                            </Link>
                        )}
                    </Button>
                </div>

                {/* Add subtitle context below */}
                <div className="justify-center p-4 margin-top-10">
                    <p className="text-xs text-gray-600 dark:text-gray-400">
                        Make sure our shopify app is installed in your store.
                    </p>

                    <p className="text-xs text-gray-600 dark:text-gray-400">
                        just click on the above button to verify if you're not sure :)
                    </p>
                </div>
            </div>
        </div>
    )
}
