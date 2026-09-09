const { test, expect } = require("@playwright/test");

// Stop 99999 makes up its buses relative to the server clock: 42 in 1m30s,
// then 666 in 5m30s, 17m and 32m. Freezing the page clock at the same moment
// keeps the rendered countdown steady, and the half minutes keep each value
// well clear of the boundary where it would tick over mid-assertion.
const STOP = "/?id=99999";

test("the initial countdown stays in place when JavaScript starts, then ticks", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.clock.install({ time: new Date() });
  let releaseScript;
  const scriptGate = new Promise((resolve) => { releaseScript = resolve; });
  await page.route("**/static/app.js?*", async (route) => {
    await scriptGate;
    await route.continue();
  });

  // Hold back the real script so a later JS update cannot hide a bad first render.
  await page.goto(STOP, { waitUntil: "commit" });
  const buses = page.locator(".buses");
  let initialBounds;
  try {
    await expect(page.locator("time")).toHaveText(["1m", "5m", "17m", "32m"]);
    await expect(buses).toHaveCSS("display", "table");
    initialBounds = await buses.boundingBox();
  } finally {
    releaseScript();
  }
  await page.waitForLoadState("load");
  await expect(page.locator("time")).toHaveText(["1m", "5m", "17m", "32m"]);
  expect(await buses.boundingBox()).toEqual(initialBounds);

  // and it ticks: runFor fires the setTimeout chain in static/app.js,
  // fastForward would jump the clock straight past it
  // An edited stop defers refresh so this tests the original arrival times.
  await page.getByLabel("Bus stop #").fill("123");
  await page.clock.runFor("01:00");
  await expect(page.locator("time").nth(1)).toHaveText("4m");
  await page.getByLabel("Bus stop #").fill("99999");

  // eyeball these after an HTML change: the clock is frozen, so they only
  // differ when the rendering does
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.screenshot({ path: "test-results/desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: "test-results/mobile.png", fullPage: true });
});

test("countdown corrects itself after the tab is frozen", async ({ page }) => {
  await page.clock.install({ time: new Date() });
  await page.goto(STOP);
  await expect(page.locator("time").nth(1)).toHaveText("5m");
  await page.getByLabel("Bus stop #").fill("123"); // Keep this response while time advances.

  // a locked phone or backgrounded tab: the clock jumps forward while the
  // timer chain does not keep up with it
  await page.clock.fastForward("02:00");

  // the first tick after waking recomputes from the clock. Decrementing a
  // captured value would still read 5m here.
  await expect(page.locator("time").nth(1)).toHaveText("3m");
});

test("page weight", async ({ page }) => {
  const weights = {};
  page.on("response", async (res) => {
    const type = res.request().resourceType();
    if (["document", "stylesheet", "script"].includes(type)) {
      weights[new URL(res.url()).pathname] = (await res.body()).length;
    }
  });

  await page.goto(STOP, { waitUntil: "load" });

  // the paint entry is usually not recorded yet on load, so wait for it
  const paint = await page.evaluate(() =>
    Promise.race([
      new Promise((resolve) => {
        new PerformanceObserver((list, obs) => {
          obs.disconnect();
          resolve(list.getEntries()[0].startTime);
        }).observe({ type: "paint", buffered: true });
      }),
      new Promise((resolve) => setTimeout(resolve, 2000, null)),
    ])
  );
  const total = Object.values(weights).reduce((a, b) => a + b, 0);

  console.log("\n  first contentful paint: %sms", paint === null ? "n/a" : Math.round(paint));
  for (const [path, size] of Object.entries(weights)) {
    console.log("  %s %s bytes", path.padEnd(20), size);
  }
  console.log("  %s %s bytes\n", "TOTAL".padEnd(20), total);

  // a ratchet, not a benchmark: lower it as the refactor lands
  expect(total).toBeLessThan(12000);
});
