const { test, expect } = require("@playwright/test");

// the visit history static/app.js keeps in localStorage, most visited first.
// "Opp Cycle & Carriage" is one of the 24 real stop names containing an &
const HISTORY = {
  "03219": { count: 5, name: "Opp Cycle & Carriage" },
  "12345": { count: 2 },
};

const seed = (page, value) =>
  page.addInitScript((v) => window.localStorage.setItem("history", v), value);

test("lists visited stops, most visited first", async ({ page }) => {
  await seed(page, JSON.stringify(HISTORY));
  await page.goto("/?id=99999");

  // 99999 is this visit, so it joins the list with a count of 1
  await expect(page.locator("#stations li")).toHaveText([
    "03219 Opp Cycle & Carriage (5)",
    "12345",
    "99999 Test Bus Stop (1)",
  ]);

  // no &name= : the server never read it, and encodeURI left the & in stop
  // names unescaped, producing a malformed query string
  const href = await page.locator("#stations a").first().getAttribute("href");
  expect(href).toBe("/?id=03219");
});

test("records the visit in localStorage", async ({ page }) => {
  const history = () =>
    page.evaluate(() => JSON.parse(window.localStorage.getItem("history")));

  await page.goto("/?id=99999");
  expect(await history()).toEqual({
    "99999": { count: 1, name: "Test Bus Stop" },
  });

  await page.reload();
  expect((await history())["99999"].count).toBe(2);
});

test("survives a corrupt history", async ({ page }) => {
  await seed(page, "{not json");
  await page.goto("/?id=99999");

  // a bad value must not take out the rest of the load handler
  await expect(page.locator("#stations li")).toHaveText([
    "99999 Test Bus Stop (1)",
  ]);
});
