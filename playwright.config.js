// Browser tests, not part of CI. Run `make browsertest` when refactoring the
// HTML. Everything drives the test stop 99999, which makes up its own buses,
// so no ACCOUNTKEY is needed and the countdown is the same every run.
module.exports = {
  testDir: "./e2e",
  reporter: [["list"]],
  use: { baseURL: "http://localhost:8081" },
  webServer: {
    command: "PORT=8081 go run .",
    url: "http://localhost:8081/?id=99999",
    reuseExistingServer: true,
  },
};
