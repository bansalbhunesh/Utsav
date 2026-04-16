"use client";

import { useEffect } from "react";

/** Registers `public/sw.js` for PWA baseline (extend with Workbox for offline schedule). */
export function ServiceWorkerRegister() {
  useEffect(() => {
    if (typeof window === "undefined" || !("serviceWorker" in navigator)) return;
    void navigator.serviceWorker.register("/sw.js").catch(() => {
      /* ignore registration errors in dev */
    });
  }, []);
  return null;
}
