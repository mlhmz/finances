import { expect, test } from "@playwright/test";

/**
 * E2E tests for Feature 2: Multi-Currency with Money Pattern
 *
 * Verifies that the registration form's currency picker is wired to the
 * Go currency registry (currency.Supported()) and behaves correctly.
 */

const REGISTER_URL = "/register?email=currency-test%40example.com";

test.describe("Currency picker on registration form", () => {
	test("shows exactly one option: Euro (EUR)", async ({ page }) => {
		await page.goto(REGISTER_URL);

		const options = page.locator('select[name="currency"] option');
		await expect(options).toHaveCount(1);
		await expect(options.first()).toHaveText("Euro (EUR)");
		await expect(options.first()).toHaveAttribute("value", "EUR");
	});

	test("currency select element is present and required", async ({ page }) => {
		await page.goto(REGISTER_URL);
		const select = page.locator('select[name="currency"]');
		await expect(select).toBeAttached();
		await expect(select).toHaveAttribute("required");
	});
});

test.describe("Registration with currency (API)", () => {
	const TS = Date.now();

	test("accepts EUR as a valid currency code", async ({ request }) => {
		const email = `e2e-currency-${TS}@example.com`;
		await request.post("/auth/request", { form: { email } });
		const res = await request.post("/register", {
			form: { email, full_name: "Currency Tester", currency: "EUR" },
		});
		expect(res.status()).toBe(200);
		const body = await res.text();
		// Successful registration renders the OTP verification form
		expect(body).toContain('name="code"');
	});

	test("rejects an unsupported currency code", async ({ request }) => {
		const email = `e2e-bad-currency-${TS}@example.com`;
		await request.post("/auth/request", { form: { email } });
		const res = await request.post("/register", {
			form: { email, full_name: "Bad Currency", currency: "XYZ" },
		});
		const body = await res.text();
		expect(body).toContain("Invalid currency");
	});
});
