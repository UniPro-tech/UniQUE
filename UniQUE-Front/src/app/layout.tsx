import type { Metadata, Viewport } from "next";

import "./globals.css";
import { BIZ_UDPGothic, Noto_Sans } from "next/font/google";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { cn } from "@/lib/utils";

const notoSans = Noto_Sans({
  weight: "400",
  subsets: ["cyrillic", "greek-ext", "latin", "latin-ext"],
  variable: "--font-sans",
  fallback: ["BIZ UDPGothic"],
});

const udpbizGothic = BIZ_UDPGothic({
  weight: "400",
  subsets: ["cyrillic", "greek-ext", "latin", "latin-ext"],
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
    <html
      lang="ja"
      className={`${cn("font-sans", notoSans.variable)} ${cn(udpbizGothic.variable)}`}
      suppressHydrationWarning
    >
      <body className={`antialiased`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          {children}
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
