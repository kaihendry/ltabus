const fs = require("fs");

// A full-page screenshot and an ARIA snapshot of what the user sees now.
async function proof(page, testInfo, name) {
  await page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true });
  const aria = await page.locator("body").ariaSnapshot();
  fs.writeFileSync(testInfo.outputPath(`${name}.aria.yml`), `# ${page.url()}\n${aria}\n`);
}

// The server's log, for checking side effects. Slice from a length taken
// before the action to see only what that action caused.
const serverLog = () => fs.readFileSync(`${process.env.VERIFY_RUN}/server.log`, "utf8");

module.exports = { proof, serverLog };
