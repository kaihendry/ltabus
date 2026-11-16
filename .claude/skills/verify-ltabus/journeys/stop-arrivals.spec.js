const { test, expect } = require("@playwright/test");
const { proof } = require("./proof");

// Stop 99999 makes up buses relative to the server clock: 42 in 1m30s, then
// 666 in 5m30s, 17m30s and 32m30s. The text holds for 30s after the page loads.
const TIMES = ["1m", "5m", "17m", "32m"];

test("look up a stop by number", async ({ page }, testInfo) => {
  await page.goto("/");
  await page.getByLabel("Bus stop #").fill("99999");
  await proof(page, testInfo, "1-entered");
  await page.getByRole("button", { name: "Bus due times" }).click();

  await expect(page).toHaveURL(/\/\?id=99999$/);
  await expect(page).toHaveTitle("Singapore bus arrival times for stop 99999");
  await expect(page.locator("#namedBusStop")).toHaveText("Test Bus Stop");
  await expect(page.locator(".buses li strong")).toHaveText(["42", "666"]);
  await expect(page.locator(".buses time")).toHaveText(TIMES);
  await expect(page.locator(".buses time")).toHaveClass(["load-seats", "load-standing", "load-full", "load-seats"]);
  await expect(page.getByRole("link", { name: "42", exact: true })).toHaveAttribute("href", "https://busrouter.sg/#/services/42");
  await expect(page.getByLabel("Bus stop #")).toHaveValue("99999");
  await proof(page, testInfo, "2-arrivals");
});

test("a stop number that is not five digits is not submitted", async ({ page }, testInfo) => {
  await page.goto("/");
  await page.getByLabel("Bus stop #").fill("123");
  await page.getByRole("button", { name: "Bus due times" }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.locator(".buses li")).toHaveCount(0);
  await proof(page, testInfo, "rejected");
});

test("the stop has its own icon", async ({ page }, testInfo) => {
  await page.goto("/?id=99999");
  await expect(page.locator("link[rel=icon]")).toHaveAttribute("href", "/icon?stop=99999");
  const icon = await page.request.get("/icon?stop=99999");
  expect(icon.status()).toBe(200);
  expect(icon.headers()["content-type"]).toBe("image/png");
  const png = await icon.body();
  require("fs").writeFileSync(testInfo.outputPath("icon-99999.png"), png);
  expect([png.readUInt32BE(16), png.readUInt32BE(20)]).toEqual([200, 200]);
  expect((await page.request.get("/icon?stop=abc")).status()).toBe(400);
});

test.describe("without JavaScript", () => {
  test.use({ javaScriptEnabled: false });

  test("the server-rendered countdown is already right", async ({ page }, testInfo) => {
    await page.goto("/");
    await page.getByLabel("Bus stop #").fill("99999");
    await page.getByRole("button", { name: "Bus due times" }).click();
    await expect(page).toHaveURL(/\?id=99999$/);
    await expect(page.locator(".buses time")).toHaveText(TIMES);
    await proof(page, testInfo, "nojs");
  });
});

test("[needs ACCOUNTKEY] a real stop shows live datamall arrivals", async ({ page }, testInfo) => {
  const res = await page.goto("/?id=01012");
  await proof(page, testInfo, "real-stop");
  expect(res.status(), "real stops need ACCOUNTKEY at launch").toBe(200);
  await expect(page.locator("#namedBusStop")).toHaveText("Hotel Grand Pacific");
});
