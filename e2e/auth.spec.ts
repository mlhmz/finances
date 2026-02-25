import { expect, test } from "@playwright/test";

/**
 * E2E tests for Feature 1: Authentication
 *
 * Notes on Web Awesome (wa-*) components:
 *  - `wa-input` renders the actual <input> inside an open shadow root.
 *    To fill or read values, chain `.locator('input')` to pierce shadow DOM.
 *  - `wa-button` renders a <button> inside shadow DOM.
 *    Use `locator('wa-button:has-text("…")')` for text-based lookup.
 *  - Use `toBeAttached()` (DOM present) instead of `toBeVisible()` for
 *    outer custom elements whose CSS visibility depends on component init.
 */

const TS = Date.now();
const RETURN_USER_EMAIL = `e2e-return-${TS}@example.com`;
const LOCKOUT_EMAIL = `e2e-lockout-${TS}@example.com`;

// In TEST_MODE the server always issues the fixed OTP "00000000".
const TEST_OTP = "00000000";

type Req = import("@playwright/test").APIRequestContext;

/** Register a brand-new user (registration OTP is the fixed TEST_OTP). */
async function registerUser(request: Req, email: string, name: string) {
	await request.post("/auth/request", { form: { email } });
	await request.post("/register", {
		form: { email, full_name: name, currency: "EUR" },
	});
}

// ── Auth guards ───────────────────────────────────────────────────────────────

test.describe("Auth guards", () => {
	test("unauthenticated GET / redirects to /login", async ({ page }) => {
		await page.goto("/");
		await expect(page).toHaveURL(/\/login/);
	});

	test("GET /register without ?email= redirects to /login", async ({
		page,
	}) => {
		await page.goto("/register");
		await expect(page).toHaveURL(/\/login/);
	});
});

// ── Login page ────────────────────────────────────────────────────────────────

test.describe("Login page", () => {
	test.beforeEach(async ({ page }) => {
		await page.goto("/login");
	});

	test("renders the sign-in heading", async ({ page }) => {
		await expect(page.locator("h1")).toContainText("Sign in");
	});

	test("has an email wa-input element", async ({ page }) => {
		await expect(page.locator('wa-input[name="email"]')).toBeAttached();
	});

	test("has a Continue submit wa-button", async ({ page }) => {
		await expect(page.locator('wa-button[type="submit"]')).toBeAttached();
	});
});

// ── Register page ─────────────────────────────────────────────────────────────

test.describe("Register page", () => {
	test("pre-fills the email from query param (attribute)", async ({ page }) => {
		await page.goto("/register?email=test%40example.com");
		await expect(page.locator('wa-input[name="email"]')).toHaveAttribute(
			"value",
			"test@example.com",
		);
	});

	test("shows full-name wa-input and currency select", async ({ page }) => {
		await page.goto("/register?email=test%40example.com");
		await expect(page.locator('wa-input[name="full_name"]')).toBeAttached();
		await expect(page.locator('select[name="currency"]')).toBeAttached();
	});

	test("server returns error page for blank full name (API)", async ({
		request,
	}) => {
		const res = await request.post("/register", {
			form: { email: "blank-name@example.com", full_name: "", currency: "EUR" },
		});
		const body = await res.text();
		expect(body).toContain("Full name is required");
	});
});

// ── New user flow (API) ───────────────────────────────────────────────────────

test.describe("New user registration + OTP flow (API)", () => {
	test("unknown email → redirects to /register", async ({ request }) => {
		const email = `e2e-unknown-${TS}@example.com`;
		const res = await request.post("/auth/request", { form: { email } });
		expect(res.headers()["hx-redirect"]).toMatch(/\/register/);
		// drain possible OTP (no user created yet — just a guard)
	});

	test("full register → OTP → HX-Redirect: /", async ({ request }) => {
		const email = `e2e-full-${TS}@example.com`;

		await request.post("/auth/request", { form: { email } });
		const regRes = await request.post("/register", {
			form: { email, full_name: "E2E Tester", currency: "EUR" },
		});
		expect(regRes.status()).toBe(200);

		const verifyRes = await request.post("/auth/verify", {
			form: { email, code: TEST_OTP },
		});
		expect(verifyRes.status()).toBe(200);
		expect(verifyRes.headers()["hx-redirect"]).toBe("/");
	});

	test("register page has title and correct form action", async ({ page }) => {
		const email = `e2e-page-${TS}@example.com`;
		await page.goto(`/register?email=${encodeURIComponent(email)}`);
		await expect(page.locator("h1")).toContainText("Create account");
		await expect(page.locator('form[action="/register"]')).toBeAttached();
	});
});

// ── Known-user login flow ─────────────────────────────────────────────────────

test.describe("Known-user login flow", () => {
	test.beforeAll(async ({ request }) => {
		await registerUser(request, RETURN_USER_EMAIL, "Return User");
	});

	test("POST /auth/request returns OTP form HTML fragment", async ({
		request,
	}) => {
		const res = await request.post("/auth/request", {
			form: { email: RETURN_USER_EMAIL },
		});
		expect(res.status()).toBe(200);
		expect(await res.text()).toContain('name="code"');
	});

	test("wrong OTP returns attempt-count error fragment", async ({
		request,
	}) => {
		await request.post("/auth/request", { form: { email: RETURN_USER_EMAIL } });
		const res = await request.post("/auth/verify", {
			form: { email: RETURN_USER_EMAIL, code: "WRONGCO1" },
		});
		expect(await res.text()).toMatch(/incorrect code|2 attempt/i);
		// Re-issue a fresh OTP so the attempt counter resets for subsequent tests
		await request.post("/auth/request", { form: { email: RETURN_USER_EMAIL } });
	});

	test("correct OTP issues cookies and returns HX-Redirect: /", async ({
		request,
	}) => {
		await request.post("/auth/request", { form: { email: RETURN_USER_EMAIL } });
		const res = await request.post("/auth/verify", {
			form: { email: RETURN_USER_EMAIL, code: TEST_OTP },
		});
		expect(res.status()).toBe(200);
		expect(res.headers()["hx-redirect"]).toBe("/");

		const cookies = res.headers()["set-cookie"] ?? "";
		expect(cookies).toContain("access_token");
		expect(cookies).toContain("refresh_token");
	});
});

// ── Logout ────────────────────────────────────────────────────────────────────

test.describe("Logout", () => {
	test("authenticated POST /auth/logout redirects to /login", async ({
		request,
	}) => {
		const email = `e2e-logout-${TS}@example.com`;

		// Register + verify → session cookies stored in this request context
		await request.post("/auth/request", { form: { email } });
		await request.post("/register", {
			form: { email, full_name: "Logout User", currency: "EUR" },
		});
		await request.post("/auth/verify", { form: { email, code: TEST_OTP } });

		// Now authenticated; logout should clear cookies and land on /login
		const res = await request.post("/auth/logout");
		expect(res.url()).toMatch(/\/login/);
	});
});

// ── OTP lockout ───────────────────────────────────────────────────────────────

test.describe("OTP lockout", () => {
	test.beforeAll(async ({ request }) => {
		await registerUser(request, LOCKOUT_EMAIL, "Lockout User");
	});

	test("after 3 wrong codes the OTP is invalidated", async ({ request }) => {
		await request.post("/auth/request", { form: { email: LOCKOUT_EMAIL } });

		for (let i = 0; i < 3; i++) {
			await request.post("/auth/verify", {
				form: { email: LOCKOUT_EMAIL, code: `WRONG00${i}` },
			});
		}

		// OTP deleted after 3 failures; even the correct code now fails
		const res = await request.post("/auth/verify", {
			form: { email: LOCKOUT_EMAIL, code: TEST_OTP },
		});
		expect(await res.text()).toMatch(
			/no active code|too many attempts|request a new/i,
		);
	});
});
