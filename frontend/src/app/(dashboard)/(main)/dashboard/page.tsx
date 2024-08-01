"use client"

import { siteConfig } from "@/app/siteConfig";
import React from "react"
import { useEffect } from "react"
import cookies from 'nookies';

import axios from 'axios';

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { DashboardStatistics } from "@/components/ui/DashboardStatistics";
import { DashboardClickerOS } from "@/components/ui/DashboardClickerOS";

export default function Dashboard() {
  const [userinfo, setUserinfo] = React.useState({});

  const GetRedirectUrl = async (shopName: string) => {
    const response = await fetch(`${siteConfig.baseApiUrl}/api/user/shopify-oauth?shop=${shopName}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    const data = await response.json();
    return data;
  }

  const handleAddShop = (shopName: string) => {
    // window.location.href = `${siteConfig.baseApiUrl}/api/user/shopify-oauth?shop=${shopName}`;
    GetRedirectUrl(shopName).then((data) => {
      window.location.href = data.auth_url;
    });
  };

  // make a request to API + /api/shop/priv/all
  // get the response, check if data.shops is empty.
  // if it is, then show the "Add a shop" button
  useEffect(() => {
    let accessToken = cookies.get(null).access_token;

    // use axios to make a request to the API
    axios.get(`${siteConfig.baseApiUrl}/api/user/private/getinfo`, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${accessToken}`,
      },
    }).then((response) => {
      localStorage.setItem('userinfo', JSON.stringify(response.data.user));
      setUserinfo(response.data.user);
    }).catch((error) => {
      if (error.response.status === 401) {
        window.location.href = '/logout';
      }
    });
  }, []);

  return (
    <>
      <section aria-labelledby="flows-title">
        <h1
          id="overall-title"
          className="scroll-mt-10 text-lg font-semibold text-gray-900 sm:text-xl dark:text-gray-50"
        >
          Videos
        </h1>
      </section>

      <section aria-labelledby="flows-description" className="mb-4">
        <p
          id="overall-description"
          className="text-sm text-gray-500 dark:text-gray-400"
        >
          View your previous videos here.
        </p>
      </section>
    </>
  )
}
