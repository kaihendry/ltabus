const { test, expect } = require("@playwright/test");

for (const url of ["/", "/?id=99999"]) {
  test(`page landmarks and stop link on ${url}`, async ({ page }) => {
    await page.goto(url);
    await expect(page.getByRole("main")).toHaveCount(1);
    await expect(page.getByRole("main").getByRole("heading", { level: 1 })).toHaveCount(1);
    await expect(page.getByRole("heading").getByRole("link", { name: "Find closest bus stop" })).toHaveAttribute("href", "/");
    await expect(page.getByRole("contentinfo")).toHaveCount(1);
    await expect(page.getByRole("main").locator("footer")).toHaveCount(0);
    await expect(page.getByRole("heading", { name: /Last updated/ })).toHaveCount(0);
  });
}
