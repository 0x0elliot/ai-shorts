"use client"

import { useState, useEffect } from "react";
import { siteConfig } from "@/app/siteConfig";
import { setCookie } from "nookies";

const headerStyle = {
    fontSize: "24px",
    fontWeight: "bold",
    marginBottom: "20px",
  };

export default function MagicLinkAuth() {
    const [accessToken, setAccessToken] = useState("");
    const [refreshToken, setRefreshToken] = useState("");

    const verifyToken = async (accessToken_: string, refreshToken_: string) => {
        try {
            const res = await fetch(`${siteConfig.baseApiUrl}/api/user/private/getinfo`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${accessToken_}`,
                },
            });

            const data = await res.json();
            console.log("Data:", data);
            if (data.error) {
                return false
            } else {
                if (accessToken_ === "" || refreshToken_ === "") {
                    console.log("Error: Access token or refresh token is empty");
                    return false
                }

                // Save tokens to cookies
                setCookie(null, "access_token", accessToken_, { path: "/" });
                setCookie(null, "refresh_token", refreshToken_, { path: "/" });

                localStorage.setItem("userinfo", JSON.stringify(data.user));
                return true
            }
        } catch (error) {
            console.error("Error:", error);
            return false
        }
    }

    // onload, check if "token" param from URL
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        let accessToken_ = urlParams.get("accessToken") || "";
        let refreshToken_ = urlParams.get("refreshToken") || "";

        if (!accessToken_ || !refreshToken_) {
            return;
        }

        setAccessToken(accessToken_);
        setRefreshToken(refreshToken_);

        verifyToken(accessToken_, refreshToken_).then((verified) => {
            if (verified) {
                // Redirect to dashboard
                window.location.href = "/";
            } else {
                // Redirect to login page
                window.location.href = "/login";
            }
        });
        }, []);

    return (
        <div>
            <header style={headerStyle}>
                Verifying Auth Token...
            </header>

            <p style={{ marginBottom: "10px" }}>
                Meanwhile, Feel free to check out how the product works <a href="https://www.youtube.com/watch?v=dQw4w9WgXcQ">here</a>
            </p>
        </div>
    )
}

