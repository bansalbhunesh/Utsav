import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
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
  };
}
