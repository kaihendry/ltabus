const { test, expect } = require("@playwright/test");

test.use({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });

test("a manually entered stop shows countdowns without JavaScript", async ({ page }) => {
  await page.goto("/");
  await page.getByLabel("Bus stop #").fill("99999");
  await page.getByRole("button", { name: "Bus due times" }).click();

  await expect(page).toHaveURL(/\?id=99999$/);
  // Disabling JS makes this fail even if app.js would repair the text on load.
  await expect(page.locator(".buses time")).toHaveText(["1m", "5m", "17m", "32m"]);

  // The same ordinary form also remains a manual refresh without JavaScript.
  await Promise.all([
    page.waitForEvent("load"),
    page.getByRole("button", { name: "Bus due times" }).click(),
  ]);
  await expect(page).toHaveURL(/\?id=99999$/);
  await expect(page.locator(".buses time")).toHaveText(["1m", "5m", "17m", "32m"]);
});
