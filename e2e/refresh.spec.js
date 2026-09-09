const { test, expect } = require("@playwright/test");

const STOP = "/?id=99999";

// Headless Chromium keeps tabs visible. Model the browser's visibility API,
// leaving the real refresh code, timers and document requests under test.
const setHidden = (page, hidden) => page.evaluate((hidden) => {
  Object.defineProperty(document, "hidden", { configurable: true, value: hidden });
  document.dispatchEvent(new Event("visibilitychange"));
}, hidden);

test("a partially entered stop survives the old meta-refresh deadline", async ({ page }) => {
  let requests = 0;
  page.on("request", (request) => {
    if (request.isNavigationRequest()) requests++;
  });
  await page.goto(STOP);
  await page.getByLabel("Bus stop #").fill("123");

  // HTML meta refresh ignores Playwright's JS clock: use real time here so
  // reintroducing it cannot silently pass the timer-controlled tests below.
  await page.waitForTimeout(21000);
  await expect(page.getByLabel("Bus stop #")).toHaveValue("123");
  expect(requests).toBe(1);
});

test("an untouched focused stop refreshes without adding browser history", async ({ page }) => {
  await page.clock.install({ time: new Date() });
  await page.goto("/");
  await page.getByLabel("Bus stop #").fill("99999");
  await page.getByRole("button", { name: "Bus due times" }).click();
  await expect(page.getByLabel("Bus stop #")).toBeFocused();
  const arrival = await page.locator("time").first().getAttribute("datetime");
  const historyLength = await page.evaluate(() => history.length);

  // Test-stop timestamps have second precision; let the real server clock
  // advance so the next response can be distinguished from the first one.
  await page.waitForTimeout(1100);
  await Promise.all([page.waitForEvent("load"), page.clock.runFor("00:21")]);
  await expect(page.locator("time").first()).not.toHaveAttribute("datetime", arrival);
  expect(await page.evaluate(() => history.length)).toBe(historyLength);
  await page.goBack();
  await expect(page).toHaveURL("/");
});

test("hidden stops wait, then refresh stale arrivals on return", async ({ page }) => {
  await page.clock.install({ time: new Date() });
  await page.goto(STOP);
  let requests = 0;
  page.on("request", (request) => {
    if (request.isNavigationRequest()) requests++;
  });

  await setHidden(page, true);
  await page.clock.runFor("00:05");
  await setHidden(page, false);
  expect(requests).toBe(0); // A quick tab switch does not reload fresh data.
  await setHidden(page, true);
  await page.clock.runFor("01:00");
  expect(requests).toBe(0);

  await Promise.all([page.waitForEvent("load"), setHidden(page, false)]);
  expect(requests).toBe(1);
});

test("returning to an edited stop preserves it until manual submission", async ({ page }) => {
  await page.clock.install({ time: new Date() });
  await page.goto(STOP);
  await page.getByLabel("Bus stop #").fill("123");
  await setHidden(page, true);
  await page.clock.runFor("01:00");
  await setHidden(page, false);
  await page.clock.runFor("00:21");
  await expect(page.getByLabel("Bus stop #")).toHaveValue("123");

  await page.getByLabel("Bus stop #").fill("99999");
  await Promise.all([
    page.waitForEvent("load"),
    page.getByRole("button", { name: "Bus due times" }).click(),
  ]);
  await expect(page.locator("#stations li")).toHaveText(["99999 Test Bus Stop (2)"]);
});
