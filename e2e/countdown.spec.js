const { test, expect } = require("@playwright/test");

// 3m12s before the first arrival in testdata/buses.json, so the countdown
// reads the same on every run
const NOW = new Date("2026-08-24T21:00:00+08:00");

test("arrival times count down", async ({ page }) => {
  await page.clock.install({ time: NOW });
  await page.goto("/?id=01019");

  // static/app.js has replaced the timestamps with a countdown
  await expect(page.locator("time")).toHaveText(["3m", "12m", "24m", "39m"]);

  // and it ticks: runFor fires the setTimeout chain in static/app.js,
  // fastForward would jump the clock straight past it
  await page.clock.runFor("00:30");
  await expect(page.locator("time").first()).toHaveText("2m");

  // eyeball these after an HTML change: the clock is frozen, so they only
  // differ when the rendering does
  await page.screenshot({ path: "test-results/desktop.png", fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: "test-results/mobile.png", fullPage: true });
});

test("countdown corrects itself after the tab is frozen", async ({ page }) => {
  await page.clock.install({ time: NOW });
  await page.goto("/?id=01019");
  await expect(page.locator("time").first()).toHaveText("3m");

  // a locked phone or backgrounded tab: the clock jumps forward while the
  // timer chain does not keep up with it
  await page.clock.fastForward("02:00");

  // the first tick after waking recomputes from the clock. Decrementing a
  // captured value would still read 3m here.
  await expect(page.locator("time").first()).toHaveText("1m");
});

test("page weight", async ({ page }) => {
  const weights = {};
  page.on("response", async (res) => {
    const type = res.request().resourceType();
    if (["document", "stylesheet", "script"].includes(type)) {
      weights[new URL(res.url()).pathname] = (await res.body()).length;
    }
  });

  await page.goto("/?id=01019", { waitUntil: "load" });

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
