"use client"

import { Input } from "@/components/ui/input";
import { useState } from "react";
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

export default function Login() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState(false);
  const [msg, setMsg] = useState("");

  return (
    <div style={centerStyle}>
      <header style={headerStyle}>
        Magic Link Login with Email
      </header>

      <Input
        type="email"
        placeholder="Email"
        style={{ marginBottom: "10px" }}
        onChange={(e) => setEmail(e.target.value)}
      />

      <p style={{ marginBottom: "10px" }}>
        {error ? <span style={{ color: "red" }}>{msg}</span> : msg}
      </p>

      <button
        style={{
          backgroundColor: "chocolate",
          color: "white",
          padding: "10px",
          width: "100%",
          border: "none",
          cursor: "pointer",
        }}
        onClick={() => {
          console.log("Email:", email);

          // Validate email format
          const emailPattern = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}$/;
          if (!emailPattern.test(email)) {
            setError(true);
            setMsg('Error: Invalid email format');
            return;
          }

          setError(false);
          setMsg("Sending magic link...");

          // Send a GET request to the backend to generate a magic link
          fetch(`${siteConfig.baseApiUrl}/api/user/passwordless-login`, {
            method: "POST",
            body: JSON.stringify({
              email: email,
            }),
            headers: {
              "Content-Type": "application/json",
            },
          }).then((res) => {
            // Handle the response from the backend
            if (res.status === 200) {
              res.json().then((data) => {
                let message = "Magic link sent to your email";
                if (data.message) {
                  message = data.message;
                }
                setMsg(message);
                setError(false);
              });
            } else {
              setError(true);
              setMsg("Error: Incorrect email address");
              res.json().then((data) => {
                let message = "Error: Please try again later";
                if (data.message) {
                  message = "Error: " + data.message;
                }
                setMsg(message);
              });
            }
          });
        }}
      >
        Login
      </button>

    </div>
  );
}