// Drives the instance bin/launch started: the user-path journeys here, and the
// repo's e2e/ regression suite. Unlike the repo's playwright.config.js it never
// starts a server or reuses whatever answers on :8081.
const baseURL = process.env.LTABUS_URL;
const out = process.env.VERIFY_OUT;
if (!baseURL || !out) throw new Error("run journeys through bin/drive <run dir>");

module.exports = {
  outputDir: out + "/artifacts",
  reporter: [["list"]],
  workers: 1,
  use: { baseURL, trace: "on", screenshot: "on" },
  projects: [
    { name: "journeys", testDir: "./journeys" },
    { name: "e2e", testDir: "../../../e2e" },
  ],
};
