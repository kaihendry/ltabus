const { test, expect } = require("@playwright/test");
const { proof } = require("./proof");

// Real time, no page.clock: this proves what a phone left on the page does.
// e2e/refresh.spec.js covers the edit guard and hidden tabs.
test("an untouched stop page reloads itself with fresh times", async ({ page }, testInfo) => {
  test.setTimeout(60_000);
  await page.goto("/?id=99999");
  const first = await page.locator(".buses time").first().getAttribute("datetime");
  const historyLength = await page.evaluate(() => history.length);
  const updated = await page.locator("#lastupdated").textContent();
  await expect(page.locator("#lastupdated")).not.toHaveText(updated); // it ticks
  await proof(page, testInfo, "1-before");

  await page.waitForEvent("load", { timeout: 30_000 });
  await expect(page).toHaveURL(/\?id=99999$/);
  await expect(page.locator(".buses time").first()).not.toHaveAttribute("datetime", first);
  expect(await page.evaluate(() => history.length)).toBe(historyLength);
  await proof(page, testInfo, "2-reloaded");
});
