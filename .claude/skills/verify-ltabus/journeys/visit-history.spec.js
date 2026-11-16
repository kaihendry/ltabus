const { test, expect } = require("@playwright/test");
const { proof } = require("./proof");

test("stops you have opened are listed with a visit count", async ({ page }, testInfo) => {
  await page.goto("/?id=99999");
  await expect(page.locator("#stations li")).toHaveText(["99999 Test Bus Stop (1)"]);
  await page.goto("/?id=99999");

  await page.goto("/");
  await expect(page.locator("#stations li")).toHaveText(["99999 Test Bus Stop (2)"]);
  await proof(page, testInfo, "1-home");

  await page.getByRole("link", { name: "99999 Test Bus Stop (2)" }).click();
  await expect(page).toHaveURL(/\/\?id=99999$/);
  await expect(page.locator("#stations li")).toHaveText(["99999 Test Bus Stop (3)"]);
  await proof(page, testInfo, "2-revisited");
});

test("[needs ACCOUNTKEY] the most visited stop is listed first", async ({ page }, testInfo) => {
  await page.goto("/?id=01012");
  await page.goto("/?id=99999");
  await page.goto("/?id=99999");
  await page.goto("/");
  await proof(page, testInfo, "order");
  await expect(page.locator("#stations li")).toHaveText([
    "99999 Test Bus Stop (2)",
    "01012 Hotel Grand Pacific (1)",
  ]);
});
