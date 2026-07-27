/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    const apiUrl = (process.env.RAG_API_URL ?? "http://localhost:8080").replace(/\/$/, "");

    return [{ source: "/backend/:path*", destination: `${apiUrl}/:path*` }];
  },
};

module.exports = nextConfig;
