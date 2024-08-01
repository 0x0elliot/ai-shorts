"use client"

import { useEffect } from "react";
import { destroyCookie } from "nookies";
import { siteConfig } from "@/app/siteConfig";

const centerStyle = {
  display: "flex",
  flexDirection: "column",
  justifyContent: "center",
  alignItems: "center",
  minHeight: "100vh", // Adjust height as needed
  maxWidth: "400px", // Adjust maximum width as needed
  margin: "auto", // Center horizontally and vertically
};

const headerStyle = {
  fontSize: "24px",
  fontWeight: "bold",
  marginBottom: "20px",
};

export default function Logout() {
  useEffect(() => {
    fetch(`${siteConfig.baseApiUrl}/api/user/logout`, {
      method: "POST",
      // credentials: "include",
    });
    try {
      localStorage.removeItem("userinfo");
      destroyCookie(null, "access_token");
      destroyCookie(null, "refresh_token");
      localStorage.removeItem("notification_configuration");
    } catch (e) {
      console.error(e);
    }

    window.location.href = "/login";
  }, []);

  return (
    <div style={centerStyle}>
      <header style={headerStyle}>
        Logging you out...
      </header>
    </div>
  );
}