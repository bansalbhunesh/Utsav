import type { Metadata } from "next";
<<<<<<< HEAD
import { Inter, Montserrat } from "next/font/google";
import "./globals.css";
import QueryProvider from "@/providers/QueryProvider";
import AuthProvider from "@/providers/AuthProvider";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const montserrat = Montserrat({
  variable: "--font-montserrat",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "UTSAV | The Operating System for Indian Events",
  description: "Replace WhatsApp chaos, notebook shagun tracking, and scattered guest lists with one unified platform for weddings, birthdays, and parties.",
  keywords: ["event management", "shagun tracking", "indian weddings", "event planner app", "utsav"],
  authors: [{ name: "UTSAV Technologies" }],
  viewport: "width=device-width, initial-scale=1, maximum-scale=1",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${montserrat.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col font-inter">
        <QueryProvider>
          <AuthProvider>
            {children}
          </AuthProvider>
        </QueryProvider>
=======
import { ServiceWorkerRegister } from "@/components/ServiceWorkerRegister";
import "./globals.css";

export const metadata: Metadata = {
  title: "UTSAV",
  description: "India's Event Operating System",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-zinc-950 text-zinc-100 antialiased">
        <ServiceWorkerRegister />
        {children}
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
      </body>
    </html>
  );
}
