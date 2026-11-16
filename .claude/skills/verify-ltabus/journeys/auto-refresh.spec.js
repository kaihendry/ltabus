const { test, expect } = require("@playwright/test");
const { proof } = require("./proof");

// Real time throughout, no page.clock: this proves what a phone left on the
// page actually does. Each test waits out the 20s refresh interval.
test.describe.configure({ timeout: 60_000 });

test("an untouched stop page reloads itself with fresh times", async ({ page }, testInfo) => {
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

test("a stop number being edited is not wiped by a refresh", async ({ page }, testInfo) => {
  let navigations = 0;
  page.on("framenavigated", (frame) => frame === page.mainFrame() && navigations++);
  await page.goto("/?id=99999");
  await page.getByLabel("Bus stop #").fill("123");

  await page.waitForTimeout(25_000);
  await expect(page.getByLabel("Bus stop #")).toHaveValue("123");
  expect(navigations).toBe(1);
  await proof(page, testInfo, "edited");
});
