/** @type {import('next').NextConfig} */
const baseApiUrl = process.env.BASE_API_URL;
const useMock = process.env.NEXT_PUBLIC_USE_MOCK === "true";
const nextConfig = {
  distDir: "build",
  reactStrictMode: true,
  productionBrowserSourceMaps: false,
  ...(baseApiUrl ? { env: { NEXT_PUBLIC_API_URL: baseApiUrl } } : {}),
  async rewrites() {
    if (!useMock) return [];
    return [
      {
        source: "/osm/api/workflows",
        destination: "/api/mock/api/workflows-list",
      },
      {
        source: "/osm/api/:path*",
        destination: "/api/mock/api/:path*",
      },
    ];
  },
  ...(process.env.NEXT_EXPORT === "true"
    ? {
        output: "export",
        images: { unoptimized: true },
      }
    : {}),
};

module.exports = nextConfig;
