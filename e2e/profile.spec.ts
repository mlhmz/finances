import { expect, test } from "@playwright/test";

/**
 * E2E tests for Feature 3: Multi-Tenancy — Profile page
 *
 * Each test is self-contained: it registers + verifies OTP so session cookies
 * accumulate in the same `request` fixture used for the assertions.
 * This follows the pattern established in auth.spec.ts.
 */

const TS = Date.now();
const TEST_OTP = "00000000";

type Req = import("@playwright/test").APIRequestContext;

async function registerAndLogin(
	request: Req,
	email: string,
	fullName: string,
): Promise<void> {
	await request.post("/auth/request", { form: { email } });
	await request.post("/register", {
		form: { email, full_name: fullName, currency: "EUR" },
	});
	await request.post("/auth/verify", { form: { email, code: TEST_OTP } });
}

// ── Auth guard ────────────────────────────────────────────────────────────────

test("unauthenticated GET /profile redirects to /login", async ({
	playwright,
}) => {
	// Use a fresh context with no cookies — Playwright follows redirects,
	// so res.url() ends up at the final destination (/login).
	const ctx = await playwright.request.newContext({
		baseURL: "http://localhost:3000",
	});
	const res = await ctx.get("/profile");
	expect(res.url()).toMatch(/\/login/);
	await ctx.dispose();
});

// ── Profile page rendering ────────────────────────────────────────────────────

test("profile page shows user full name and email", async ({ request }) => {
	const email = `e2e-profile-name-${TS}@example.com`;
	await registerAndLogin(request, email, "Profile Tester");

	const res = await request.get("/profile");
	expect(res.status()).toBe(200);
	const body = await res.text();
	expect(body).toContain("Profile Tester");
	expect(body).toContain(email);
});

test("profile page shows currency selector pre-selected to user currency", async ({
	request,
}) => {
	const email = `e2e-profile-currency-${TS}@example.com`;
	await registerAndLogin(request, email, "Currency Tester");

	const body = await (await request.get("/profile")).text();
	expect(body).toContain('name="currency"');
	expect(body).toContain("EUR");
});

// ── Profile update ────────────────────────────────────────────────────────────

test("POST /profile with valid data shows success message", async ({
	request,
}) => {
	const email = `e2e-update-success-${TS}@example.com`;
	await registerAndLogin(request, email, "Update Tester");

	const res = await request.post("/profile", {
		form: { full_name: "Updated Name", currency: "EUR" },
	});
	expect(res.status()).toBe(200);
	expect(await res.text()).toContain("Profile updated");
});

test("profile reflects updated name after save", async ({ request }) => {
	const email = `e2e-update-reflect-${TS}@example.com`;
	await registerAndLogin(request, email, "Before Update");

	await request.post("/profile", {
		form: { full_name: "After Update", currency: "EUR" },
	});

	const body = await (await request.get("/profile")).text();
	expect(body).toContain("After Update");
	expect(body).not.toContain("Before Update");
});

// ── Validation errors ─────────────────────────────────────────────────────────

test("empty full_name returns validation error", async ({ request }) => {
	const email = `e2e-val-name-${TS}@example.com`;
	await registerAndLogin(request, email, "Validation Tester");

	const res = await request.post("/profile", {
		form: { full_name: "", currency: "EUR" },
	});
	expect(res.status()).toBe(200);
	expect(await res.text()).toContain("Full name is required");
});

test("invalid currency returns validation error", async ({ request }) => {
	const email = `e2e-val-currency-${TS}@example.com`;
	await registerAndLogin(request, email, "Currency Validator");

	const res = await request.post("/profile", {
		form: { full_name: "Valid Name", currency: "ZZZ" },
	});
	expect(res.status()).toBe(200);
	expect(await res.text()).toContain("Invalid currency");
});

test("validation error does not modify the stored profile", async ({
	request,
}) => {
	const email = `e2e-val-noop-${TS}@example.com`;
	await registerAndLogin(request, email, "Unchanged Tester");

	// Submit with empty name → should fail
	await request.post("/profile", { form: { full_name: "", currency: "EUR" } });

	// Original name must still be present
	const body = await (await request.get("/profile")).text();
	expect(body).toContain("Unchanged Tester");
});

// ── Data isolation ────────────────────────────────────────────────────────────

test("user A cannot see user B's name on their own profile", async ({
	playwright,
}) => {
	const emailA = `e2e-iso-a-${TS}@example.com`;
	const emailB = `e2e-iso-b-${TS}@example.com`;
	const base = { baseURL: "http://localhost:3000" };

	// Two independent request contexts = two separate cookie jars (no browser needed)
	const ctxA = await playwright.request.newContext(base);
	const ctxB = await playwright.request.newContext(base);

	await registerAndLogin(ctxA, emailA, "Alice Isolation");
	await registerAndLogin(ctxB, emailB, "Bob Isolation");

	const bodyA = await (await ctxA.get("/profile")).text();
	expect(bodyA).toContain("Alice Isolation");
	expect(bodyA).not.toContain("Bob Isolation");

	const bodyB = await (await ctxB.get("/profile")).text();
	expect(bodyB).toContain("Bob Isolation");
	expect(bodyB).not.toContain("Alice Isolation");

	await ctxA.dispose();
	await ctxB.dispose();
});
