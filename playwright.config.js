// Browser tests, not part of CI. Run `make browsertest` when refactoring the
// HTML. The server is started with canned arrivals so it needs no ACCOUNTKEY
// and the countdown is the same every run.
module.exports = {
  testDir: "./e2e",
  reporter: [["list"]],
  use: { baseURL: "http://localhost:8081" },
  webServer: {
    command: "FIXTURE=testdata/buses.json PORT=8081 go run .",
    url: "http://localhost:8081/?id=01019",
    reuseExistingServer: true,
  },
};
