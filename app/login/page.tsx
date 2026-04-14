<<<<<<< HEAD
import { LoginForm } from "@/components/auth/LoginForm";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";

export default function LoginPage() {
  return (
    <div className="min-h-screen flex flex-col bg-zinc-50 lg:bg-white">
      {/* Header */}
      <header className="p-6">
        <Link 
          href="/" 
          className="inline-flex items-center text-sm font-medium text-zinc-500 hover:text-orange-600 transition-colors"
        >
          <ChevronLeft className="mr-1 h-4 w-4" />
          Back to Home
        </Link>
      </header>

      {/* Hero-like layout for login */}
      <main className="flex-1 flex flex-col items-center justify-center px-6 py-12">
        <div className="w-full max-w-[1000px] grid lg:grid-cols-2 gap-12 items-center">
          
          {/* Brand/Marketing Side (Hidden on mobile) */}
          <div className="hidden lg:flex flex-col space-y-8">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 bg-orange-600 rounded-xl flex items-center justify-center text-white font-bold text-2xl shadow-lg shadow-orange-200">
                U
              </div>
              <span className="text-2xl font-bold font-heading tracking-tight text-orange-600">
                UTSAV
              </span>
            </div>
            
            <div className="space-y-4">
              <h1 className="text-5xl font-bold font-heading tracking-tight text-zinc-900 leading-tight">
                Start Managing Your <br />
                <span className="text-orange-600 italic">Event Excellence</span>
              </h1>
              <p className="text-xl text-zinc-600 max-w-md leading-relaxed">
                Join thousands of hosts and organisers who use Utsav to create 
                memorable experiences without the stress.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-6 pt-4">
              <div className="space-y-1">
                <p className="text-2xl font-bold text-zinc-900 tracking-tight">10M+</p>
                <p className="text-sm text-zinc-500 font-medium">Annual Weddings</p>
              </div>
              <div className="space-y-1">
                <p className="text-2xl font-bold text-zinc-900 tracking-tight">Zero</p>
                <p className="text-sm text-zinc-500 font-medium">WhatsApp Chaos</p>
              </div>
            </div>
          </div>

          {/* Login Form Side */}
          <div className="flex justify-center">
            <LoginForm />
          </div>

        </div>
      </main>

      {/* Footer (Simplified) */}
      <footer className="p-8 text-center border-t border-zinc-100 bg-white lg:bg-zinc-50/50">
        <p className="text-zinc-500 text-xs font-medium">
          &copy; 2026 UTSAV Technologies. Handcrafted in India.
        </p>
      </footer>
    </div>
=======
"use client";

import Link from "next/link";
import { useState } from "react";
import { apiFetch, setTokens } from "@/lib/api";

export default function LoginPage() {
  const [phone, setPhone] = useState("+919876543210");
  const [code, setCode] = useState("123456");
  const [msg, setMsg] = useState<string | null>(null);

  return (
    <main className="mx-auto flex max-w-md flex-col gap-4 px-6 py-16">
      <Link href="/" className="text-sm text-zinc-400 hover:text-white">
        ← Home
      </Link>
      <h1 className="text-2xl font-semibold text-white">Phone login</h1>
      <p className="text-sm text-zinc-500">Dev OTP defaults to 123456 (`DEV_OTP_CODE` on API).</p>
      <input
        className="rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
        value={phone}
        onChange={(e) => setPhone(e.target.value)}
        placeholder="Phone"
      />
      <button
        type="button"
        className="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-black"
        onClick={() =>
          void (async () => {
            setMsg(null);
            await apiFetch("/v1/auth/otp/request", { method: "POST", json: { phone } });
            setMsg("OTP sent.");
          })()
        }
      >
        Request OTP
      </button>
      <input
        className="rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-white"
        value={code}
        onChange={(e) => setCode(e.target.value)}
      />
      <button
        type="button"
        className="rounded-lg border border-zinc-700 px-4 py-2 text-sm text-white"
        onClick={() =>
          void (async () => {
            setMsg(null);
            const d = await apiFetch<{ access_token: string; refresh_token?: string }>(
              "/v1/auth/otp/verify",
              { method: "POST", json: { phone, code } },
            );
            setTokens(d.access_token, d.refresh_token);
            window.location.href = "/dashboard";
          })()
        }
      >
        Verify
      </button>
      {msg && <p className="text-sm text-zinc-400">{msg}</p>}
    </main>
>>>>>>> f7494df (feat: Architectural Level Up - Go-Authoritative Backend, RSVP OTP Flow, and Frontend Consolidation (v1.5 Final))
  );
}
