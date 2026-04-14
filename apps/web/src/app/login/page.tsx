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
  );
}
