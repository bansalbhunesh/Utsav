/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    const api = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";
    return [
      { source: "/v1/:path*", destination: `${api}/v1/:path*` },
    ];
  },
};

export default nextConfig;
