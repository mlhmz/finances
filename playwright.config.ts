import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E test configuration for the Finances app.
 * The app runs on http://localhost:3000 (go run ./cmd/finances).
 *
 * Docs: https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e',
  /* Run tests in parallel */
  fullyParallel: true,
  /* Fail the build on CI if test.only is accidentally left in source */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  /* Single worker on CI to avoid port conflicts */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter */
  reporter: [
    ['list'],
    ['html', { open: 'never' }],
  ],
  /* Shared settings for all projects */
  use: {
    /* Base URL for the Finances app */
    baseURL: 'http://localhost:3000',
    /* Collect trace on first retry */
    trace: 'on-first-retry',
    /* Screenshots on failure */
    screenshot: 'only-on-failure',
  },

  /* Test against Chromium by default; add more browsers as needed */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    // Uncomment to test in additional browsers:
    // { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    // { name: 'webkit',  use: { ...devices['Desktop Safari'] } },
  ],

  /*
   * Automatically start the Go app before running tests and shut it down after.
   * Remove `webServer` if you prefer to start the app manually first.
   */
  webServer: {
    command: 'go run ./cmd/finances',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
