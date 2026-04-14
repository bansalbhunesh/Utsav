import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
<<<<<<< HEAD
    name: "UTSAV - India's Event OS",
    short_name: "UTSAV",
    description: "Replacing WhatsApp chaos with digital elegance for Indian celebrations.",
    start_url: "/",
    display: "standalone",
    background_color: "#ffffff",
    theme_color: "#EA580C",
    icons: [
      {
        src: "/favicon.ico",
        sizes: "any",
        type: "image/x-icon",
      },
    ],
=======
    name: "UTSAV",
    short_name: "UTSAV",
    start_url: "/",
    display: "standalone",
    background_color: "#09090b",
    theme_color: "#d4a853",
    icons: [],
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
  };
}
