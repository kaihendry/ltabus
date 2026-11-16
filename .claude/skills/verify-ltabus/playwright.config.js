// Drives the instance bin/launch started. Unlike the repo's playwright.config.js
// it never starts a server or reuses whatever answers on :8081.
const baseURL = process.env.LTABUS_URL;
const out = process.env.VERIFY_OUT;
if (!baseURL || !out) throw new Error("run journeys through bin/drive <run dir>");

module.exports = {
  testDir: "./journeys",
  outputDir: out + "/artifacts",
  reporter: [["list"]],
  workers: 1,
  use: { baseURL, trace: "on", screenshot: "on" },
};
