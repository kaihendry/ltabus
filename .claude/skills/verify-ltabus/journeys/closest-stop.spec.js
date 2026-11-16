const { test, expect } = require("@playwright/test");
const { proof } = require("./proof");

// Standing exactly on stop 01012, so it is the closest one.
test.use({
  geolocation: { latitude: 1.29684825487647, longitude: 103.85253591654006 },
  permissions: ["geolocation"],
});

test("opening the home page goes to the closest stop", async ({ page }, testInfo) => {
  const arrived = page.waitForResponse((r) => new URL(r.url()).search === "?id=01012");
  await page.goto("/");
  const res = await arrived;
  const via = res.request().redirectedFrom();
  expect(via, "reached through /closest").not.toBeNull();
  expect(new URL(via.url()).pathname).toBe("/closest");
  expect((await via.response()).status()).toBe(302);
  await expect(page).toHaveURL(/\/\?id=01012$/);
  await proof(page, testInfo, "arrived");
});

test("tapping the bus stop sign from another stop relocates", async ({ page }, testInfo) => {
  await page.goto("/?id=99999");
  const arrived = page.waitForResponse((r) => new URL(r.url()).search === "?id=01012");
  await page.getByRole("heading").getByRole("link", { name: "Find closest bus stop" }).click();
  await arrived;
  await expect(page).toHaveURL(/\/\?id=01012$/);
  await proof(page, testInfo, "relocated");
});

test("[needs ACCOUNTKEY] the closest stop shows its arrivals", async ({ page }, testInfo) => {
  const arrived = page.waitForResponse((r) => new URL(r.url()).search === "?id=01012");
  await page.goto("/");
  const res = await arrived;
  await page.waitForLoadState();
  await proof(page, testInfo, "closest-arrivals");
  expect(res.status(), "real stops need ACCOUNTKEY at launch").toBe(200);
  await expect(page.locator("#namedBusStop")).toHaveText("Hotel Grand Pacific");
});
