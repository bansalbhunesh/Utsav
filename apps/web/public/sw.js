/* Minimal service worker: extend with Workbox for offline schedule (spec Phase 4). */
self.addEventListener("install", (e) => {
  e.waitUntil(self.skipWaiting());
});
self.addEventListener("activate", (e) => {
  e.waitUntil(self.clients.claim());
});
