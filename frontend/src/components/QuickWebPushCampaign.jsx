"use client";

import React, { useEffect } from "react";
import { useState } from "react";
import { Button } from "@/components/ui/button"

import axios from "axios";

import cookies from 'nookies';
import { siteConfig } from "@/app/siteConfig";
import ConfigureWebPushCampaign from "./ConfigureWebPushCampaign";
import { toast } from "./ui/use-toast";

export default function QuickWebPushCampaign() {
    const [segmentSelectionModalOpen, setSegmentSelectionModalOpen] = useState(false);
    const [configureCampaignModalOpen, setConfigureCampaignModalOpen] = useState(false);

    const [webPushSubscriberCount, setWebPushSubscriberCount] = useState(0);
    const [notificationsToAll, setNotificationsToAll] = useState(false);

    const [notifConfig, setNotifConfig] = useState({});
    const [notifConfigurations, setNotifConfigurations] = useState([]);

    const [userinfo, setUserinfo] = useState({});

    const [accessToken, setAccessToken] = useState(null);

    const handleConfigure = () => {
        configureCampaignModalOpen ? setConfigureCampaignModalOpen(false) : setConfigureCampaignModalOpen(true);
        localStorage.removeItem("notification_configuration");
    }

    const closeModalCampaign = () => {
        setConfigureCampaignModalOpen(false);
    };

    useEffect(() => {
        let accessToken = cookies.get(null).access_token;
        let current_shop_id = userinfo.current_shop_id;
        if (!current_shop_id) {
            let userInfoString = localStorage.getItem("userinfo") || "{}";
            let userInfo = JSON.parse(userInfoString);
            current_shop_id = userInfo.current_shop_id;

            // if still not found, then fetch it from the API
            if (!current_shop_id) {
                axios.get(`${siteConfig.baseApiUrl}/api/user/private/getinfo`, {
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${accessToken}`,
                    },
                }).then((response) => {
                    localStorage.setItem("userinfo", JSON.stringify(response.data));
                    setUserinfo(response.data);
                }
                ).catch((error) => {
                    console.error(error);
                    toast({
                        title: "Error",
                        description: "Failed to fetch user information",
                        variant: "destructive",
                    });
                });
            }
        }

        // check if the user has the notification configuration in the local storage
        // if not, then fetch it from the API
        let notificationConfiguration = localStorage.getItem("notification_configuration");
        axios.get(`${siteConfig.baseApiUrl}/api/notification/private/configurations?shop_id=${current_shop_id}`, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${accessToken}`,
            },
        }).then((response) => {
            setNotifConfigurations(response.data.configurations);
            if (notificationsToAll) {
                for (let i = 0; i < response.data?.configurations?.length; i++) {
                    if (response.data?.configurations[i]?.id === notificationConfiguration) {
                        setNotifConfig(response.data.configurations[i]);
                    }
                }
            }
        }).catch((error) => {
            console.error(error);
            toast({
                title: "Error",
                description: "Failed to fetch notification configuration",
                variant: "destructive",
            })
        });
    }, [localStorage.getItem("notification_configuration")]);

    useEffect(() => {
        const token = cookies.get(null).access_token;
        setAccessToken(token);

        const userInfoString = localStorage.getItem("userinfo") || "{}";
        try {
            // setUserinfo(JSON.parse(userInfoString));
            json = JSON.parse(userInfoString);
            if (json.id) {
                set(Userinfojson);
            }
        } catch (error) {
            setUserinfo({});
        }
    }, []);

    useEffect(() => {
        if (segmentSelectionModalOpen) {
            getWebPushSegments();
        }
    }, [segmentSelectionModalOpen]);

    const handleWebPushToAll = (e) => {
        console.log(e.target.checked);
        setNotificationsToAll(e.target.checked);
    }

    const handleLaunchCampaign = () => {
        axios.post(`${siteConfig.baseApiUrl}/api/notification/private/launch`, {
            all: notificationsToAll,
            shop_id: userinfo.current_shop_id,
            notification_configuration_id: notifConfig.id,
        }, {
            headers: {
                contentType: 'application/json',
                Authorization: `Bearer ${accessToken}`,
            },
        }).then((response) => {
            toast({
                title: "Success",
                description: "Campaign launched successfully",
                variant: "success",
            });
        }
        ).catch((error) => {
            console.error(error);
            toast({
                title: "Error",
                description: "Failed to launch the campaign",
                variant: "destructive",
            });
        });
    }

    const getWebPushSegments = async () => {
        // check local storage, userinfo. if it has a "CurrentShop" key, then use that shop identifier
        // if not, then ask API again.
        let localUserinfo = userinfo

        if (!localUserinfo.current_shop_id) {
            const userInfoResponse = await fetch(`${siteConfig.baseApiUrl}/api/user/private/getinfo`, {
                method: 'GET',
                headers: {
                    "Authorization": `Bearer ${accessToken}`,
                },
            });

            const userInfoData = await userInfoResponse.json();
            if (userInfoResponse.status !== 200) {
                console.error(userInfoData.message);
                return;
            }

            localStorage.setItem('userinfo', JSON.stringify(userInfoData));
            setUserinfo(userInfoData);
            localUserinfo = userInfoData;
        }

        if (!localUserinfo.current_shop_id) {
            console.error("No shop identifier found. Please select a shop first.");
            return;
        }

        const response = await fetch(`${siteConfig.baseApiUrl}/api/notification/private/push-subscribers?shop_identifier=${localUserinfo.current_shop.shop_identifier}&count_only=true`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${accessToken}`
            },
        });

        const data = await response.json();
        if (response.status !== 200) {
            console.error(data.message);
            return;
        }

        setWebPushSubscriberCount(data.count);
    }

    return (
        <div className="flex bg-white dark:bg-gray-900">
            <div className="flex-1 p-4">
                <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-50">Launch a quick campaign</h1>
                <p style={{ marginBottom: 10 }} className="text-sm text-gray-600 dark:text-gray-400">Select from your segment</p>
                {/* <div className="flex justify-center p-4 mt-10 space-x-4" style={{ marginTop: "2rem" }}> */}
                <div className="flex flex-col space-y-4">
                    <Button
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50 "
                        onClick={() => segmentSelectionModalOpen ? setSegmentSelectionModalOpen(false) : setSegmentSelectionModalOpen(true)}
                    >
                        Select a segment
                    </Button>

                    <Button
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                        disabled={!notificationsToAll}
                        onClick={() => handleConfigure()}
                    >
                        Configure
                    </Button>

                    <Button
                        className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                        disabled={!notifConfig?.id}
                        onClick={() => handleLaunchCampaign()}
                    >
                        Launch
                    </Button>
                </div>
            </div>


            {(configureCampaignModalOpen) && (
                <div className="fixed inset-0 bg-opacity-50 z-50 flex items-center justify-center">
                    <ConfigureWebPushCampaign onClose={closeModalCampaign} existingConfigurations={notifConfigurations} />
                </div>
            )}

            {(segmentSelectionModalOpen) && (
                <div className="fixed inset-0 bg-opacity-50 z-50 flex items-center justify-center">
                    <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-lg">
                        <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-50">Select a segment</h1>
                        <p className="text-sm text-gray-600 dark:text-gray-400">Select a segment to send the campaign to</p>
                        <div className="flex flex-col items-center overflow-y-auto p-4 rounded-lg">
                            {/* Iterate over all the subscribers and show them here as a checklist */}
                            <div className="flex flex-col space-y-2">
                                <div className="flex items-center space-x-2">
                                    <input type="checkbox" id="all" name="all" value="all" onChange={(e) => handleWebPushToAll(e)} />
                                    <label htmlFor="all">All ({webPushSubscriberCount}) Web Push Subscribers + Your test subscription</label>
                                </div>
                            </div>
                        </div>

                        <div className="flex justify-center">
                            <Button
                                className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50"
                                onClick={() => setSegmentSelectionModalOpen(false)}
                            >
                                Select
                            </Button>

                            <Button
                                className="text-white bg-black hover:bg-gray-800 focus:ring-2 focus:ring-gray-500 font-medium text-sm px-5 py-2.5 text-center dark:bg-white dark:text-black dark:hover:bg-gray-200 dark:focus:ring-gray-400 disabled:opacity-50 ml-2"
                                onClick={() => setSegmentSelectionModalOpen(false)}
                            >
                                Close
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>



    )
}
