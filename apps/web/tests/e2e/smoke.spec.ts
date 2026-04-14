import { expect, test } from "@playwright/test";

const phone = process.env.E2E_PHONE || "+919876543210";
const otp = process.env.E2E_OTP || "123456";
const slug = process.env.E2E_EVENT_SLUG || `e2e-${Date.now()}`;
const title = process.env.E2E_EVENT_TITLE || "E2E Wedding";

test("login -> create event -> guest RSVP", async ({ page }) => {
  await page.goto("/login");
  await page.getByPlaceholder("Phone").fill(phone);
  await page.getByRole("button", { name: "Request OTP" }).click();
  await page.getByRole("button", { name: "Verify" }).click();
  await page.waitForURL("**/dashboard");

  await page.getByRole("link", { name: "New event" }).click();
  await page.waitForURL("**/events/new");

  const inputs = page.locator("input");
  await inputs.nth(0).fill(slug);
  await inputs.nth(1).fill(title);
  await inputs.nth(2).fill("host@upi");
  await page.getByRole("button", { name: "Create" }).click();
  await page.waitForURL("**/events/*");

  await page.goto(`/e/${slug}/rsvp`);
  await page.getByRole("button", { name: "Request OTP" }).click();
  await page.locator("input").nth(1).fill(otp);
  await page.getByRole("button", { name: "Verify OTP" }).click();
  await page.getByRole("button", { name: "Submit RSVP" }).click();
  await expect(page.getByText("RSVP saved.")).toBeVisible();
});
