const { test } = require("@playwright/test");

// Slow 4G with a cold cache: a first visit on a phone, which is where
// render-blocking actually costs something. Localhost hides all of it.
const NETWORK = {
  offline: false,
  latency: 150,
  downloadThroughput: (1.6 * 1024 * 1024) / 8,
  uploadThroughput: (750 * 1024) / 8,
};
const RUNS = 7;

// the paint entry is usually not recorded yet on load, so wait for it
const firstPaint = (page) =>
  page.evaluate(() =>
    Promise.race([
      new Promise((resolve) => {
        new PerformanceObserver((list, obs) => {
          obs.disconnect();
          resolve(list.getEntriesByName("first-contentful-paint")[0]?.startTime);
        }).observe({ type: "paint", buffered: true });
      }),
      new Promise((resolve) => setTimeout(resolve, 5000, null)),
    ])
  );

test("throttled first contentful paint", async ({ page }) => {
  test.setTimeout(120000);
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Network.enable");
  await cdp.send("Network.emulateNetworkConditions", NETWORK);
  // CACHE=warm measures a meta-refresh reload, where the assets are cached
  const warm = process.env.CACHE === "warm";
  await cdp.send("Network.setCacheDisabled", { cacheDisabled: !warm });
  if (warm) await page.goto("/?id=99999", { waitUntil: "load" }); // prime it
  await cdp.send("Emulation.setCPUThrottlingRate", { rate: 4 });

  const samples = [];
  for (let i = 0; i < RUNS; i++) {
    if (!warm) await page.goto("about:blank");
    await page.goto("/?id=99999", { waitUntil: "load" });
    samples.push(await firstPaint(page));
  }
  samples.sort((a, b) => a - b);
  console.log(
    "\n  FCP median %sms   (all: %s)\n",
    Math.round(samples[Math.floor(RUNS / 2)]),
    samples.map((s) => Math.round(s)).join(", ")
  );
});
