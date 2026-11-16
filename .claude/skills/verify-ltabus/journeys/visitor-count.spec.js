const fs = require("fs");
const { test, expect, request } = require("@playwright/test");
const { serverLog } = require("./proof");

test("a new browser is counted once it sends its cookie back", async ({ page, context, baseURL }, testInfo) => {
  const from = serverLog().length;
  const doc = await page.goto("/?id=99999");
  expect(await doc.headerValue("set-cookie")).toMatch(/^visitor=visitor-[A-Z0-9]+;/);
  await page.waitForLoadState("load");
  const id = (await context.cookies()).find((c) => c.name === "visitor").value;

  // The page's own assets carry the cookie back, so those response lines count it.
  const counted = () => serverLog().slice(from).split("\n").filter((l) => l.includes(`unique=${id}`));
  await expect.poll(() => counted().length).toBeGreaterThan(0);
  const docLine = serverLog().slice(from).split("\n").find((l) => l.includes('req_path="/?id=99999"'));
  expect(docLine, "the first document request has no cookie yet").not.toContain("unique=");

  // A return visit keeps the same ID and does not set a new one.
  const again = await page.reload();
  expect(await again.headerValue("set-cookie")).toBeNull();
  expect((await context.cookies()).find((c) => c.name === "visitor").value).toBe(id);

  // Static assets never hand out an ID.
  const anon = await request.newContext({ baseURL });
  expect((await anon.get("/static/app.css")).headers()["set-cookie"]).toBeUndefined();
  await anon.dispose();

  fs.writeFileSync(testInfo.outputPath("server-log.txt"), serverLog().slice(from));
});
