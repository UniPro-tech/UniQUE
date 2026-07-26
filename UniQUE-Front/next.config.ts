import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	experimental: {
		webpackMemoryOptimizations: true,
		authInterrupts: true,
	},
	output: "standalone",
	poweredByHeader: false,
	async headers() {
		const connectSrc = [
			"'self'",
			process.env.NEXT_PUBLIC_AUTH_API_URL,
			process.env.NEXT_PUBLIC_RESOURCE_API_URL,
		]
			.filter(Boolean)
			.join(" ");
		const imgSrc = [
			"'self'",
			process.env.NEXT_PUBLIC_RESOURCE_API_URL,
			"https://cdn.discordapp.com",
		]
			.filter(Boolean)
			.join(" ");
		return [
			{
				source: "/(.*)",
				headers: [
					{
						key: "X-Content-Type-Options",
						value: "nosniff",
					},
					{
						key: "X-Frame-Options",
						value: "DENY",
					},
					{
						key: "Referrer-Policy",
						value: "strict-origin-when-cross-origin",
					},
					{
						key: "Content-Security-Policy",
						value:
							process.env.NODE_ENV === "production"
								? `default-src 'self'; img-src ${imgSrc}; base-uri 'self'; frame-ancestors 'none'; object-src 'none'; script-src 'self' 'unsafe-inline' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline'; connect-src ${connectSrc}`
								: `default-src 'self'; img-src ${imgSrc}; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline'; connect-src ${connectSrc}`,
					},
				],
			},
		];
	},
};

export default nextConfig;
