import { test, expect } from '@playwright/test';

/**
 * Smoke tests for the Finances app home page.
 * These tests verify the app loads correctly and core UI is present.
 */

test.describe('Home page', () => {
  test('loads with the correct title', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/finances/i);
  });

  test('displays the main heading', async ({ page }) => {
    await page.goto('/');
    const heading = page.locator('.card h1');
    await expect(heading).toBeVisible();
  });

  test('page returns HTTP 200', async ({ request }) => {
    const response = await request.get('/');
    expect(response.status()).toBe(200);
  });
});
