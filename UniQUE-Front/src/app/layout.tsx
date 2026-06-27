import type { Metadata, Viewport } from "next";

import "./globals.css";
import { JetBrains_Mono } from "next/font/google";
import { cn } from "@/lib/utils";

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

export const metadata: Metadata = {
  title: {
    default: "UniQUE - デジタル創作サークルUniProject メンバーズポータル",
    template: "%s | UniQUE",
  },
  description:
    "デジタル創作サークルUniProjectのメンバーズポータルサイトです。ここでメンバー登録を行ったり、各種サービスにアクセスできます。",
  applicationName: "UniQUE",
  robots: {
    index: false,
    follow: false,
  },
};

export const viewport: Viewport = {
  themeColor: "#1f8ae1",
  width: "device-width",
  initialScale: 1,
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja" className={cn("font-mono", jetbrainsMono.variable)}>
      <body className={`antialiased`}>{children}</body>
    </html>
  );
}
