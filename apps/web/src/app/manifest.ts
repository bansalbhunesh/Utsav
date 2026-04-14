import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "UTSAV",
    short_name: "UTSAV",
    start_url: "/",
    display: "standalone",
    background_color: "#09090b",
    theme_color: "#d4a853",
    icons: [],
  };
}
