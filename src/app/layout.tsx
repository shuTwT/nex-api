import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/auth-provider";
import { InitCheck } from "@/components/init-check";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";
import { Toaster } from "sonner";
import { ThemeProvider } from "@/components/theme-provider";
import { getConfigValue } from "@/lib/config";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const getMetadata = async():Promise<Metadata>=>{
  try{
    const siteName = await getConfigValue("general","siteName")
    const siteDescription = await getConfigValue("general","siteDescription")
    return {
      title: siteName||"One API - API 聚合管理系统",
      description: siteDescription||"一站式 API 聚合平台，提供高质量 API 接口服务",
    }
  }catch(e){
    console.error("Get config value error:", e);
    return {
      title: "One API - API 聚合管理系统",
      description: "一站式 API 聚合平台，提供高质量 API 接口服务",
    }
  }
}

export const metadata: Metadata = await getMetadata();






export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const session = await getServerSession(authOptions)
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="light"
          enableSystem
          disableTransitionOnChange
        >
          <AuthProvider session={session}>
            <InitCheck>{children}</InitCheck>
            <Toaster position="top-center" richColors={true} expand={true}/>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
