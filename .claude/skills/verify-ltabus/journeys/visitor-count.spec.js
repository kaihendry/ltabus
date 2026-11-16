const fs = require("fs");
const { test, expect } = require("@playwright/test");
const { serverLog } = require("./proof");

// Only a real browser shows the page's own requests carrying the cookie back.
// http_test.go covers which routes set it and that a returned one is kept.
test("a new browser is counted once it sends its cookie back", async ({ page, context }, testInfo) => {
  const from = serverLog().length;
  const doc = await page.goto("/?id=99999");
  expect(await doc.headerValue("set-cookie")).toMatch(/^visitor=visitor-[A-Z0-9]+;/);
  await page.waitForLoadState("load");
  const id = (await context.cookies()).find((c) => c.name === "visitor").value;

  const counted = () => serverLog().slice(from).split("\n").filter((l) => l.includes(`unique=${id}`));
  await expect.poll(() => counted().length).toBeGreaterThan(0);
  fs.writeFileSync(testInfo.outputPath("server-log.txt"), serverLog().slice(from));
});
