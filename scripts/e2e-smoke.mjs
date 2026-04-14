const baseUrl = process.env.E2E_BASE_URL?.trim() || "http://127.0.0.1:3000";
const paths = ["/", "/login"];

async function check(path) {
  const url = new URL(path, baseUrl).toString();
  const res = await fetch(url, { redirect: "follow" });
  if (!res.ok) {
    throw new Error(`Smoke check failed for ${url}: status ${res.status}`);
  }
  console.log(`ok ${res.status} ${url}`);
}

async function main() {
  console.log(`Running e2e smoke checks against ${baseUrl}`);
  for (const path of paths) {
    await check(path);
  }
  console.log("E2E smoke checks passed.");
}

main().catch((err) => {
  console.error(err instanceof Error ? err.message : String(err));
  process.exit(1);
});
