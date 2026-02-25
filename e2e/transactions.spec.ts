import { expect, test } from "@playwright/test";

/**
 * E2E tests for Feature 4: Manual Transaction Tracking
 *
 * All API-level tests — HTMX partial responses are inspected directly
 * so tests run without a browser and remain fast and deterministic.
 */

const TS = Date.now();
const TEST_OTP = "00000000";

type Req = import("@playwright/test").APIRequestContext;

async function registerAndLogin(
	request: Req,
	email: string,
	fullName = "Test User",
): Promise<void> {
	await request.post("/auth/request", { form: { email } });
	await request.post("/register", {
		form: { email, full_name: fullName, currency: "EUR" },
	});
	await request.post("/auth/verify", { form: { email, code: TEST_OTP } });
}

async function createTransaction(
	request: Req,
	fields: {
		type?: string;
		amount?: string;
		currency?: string;
		title?: string;
		date?: string;
		description?: string;
	} = {},
) {
	return request.post("/transactions", {
		form: {
			type: "income",
			amount: "100.00",
			currency: "EUR",
			title: "Test Income",
			date: "2026-02-25T10:00",
			description: "",
			...fields,
		},
	});
}

// ── Auth guard ────────────────────────────────────────────────────────────────

test("unauthenticated GET /transactions redirects to /login", async ({
	playwright,
}) => {
	const ctx = await playwright.request.newContext({
		baseURL: "http://localhost:3000",
	});
	const res = await ctx.get("/transactions");
	expect(res.url()).toMatch(/\/login/);
	await ctx.dispose();
});

// ── Transactions page rendering ───────────────────────────────────────────────

test.describe("Transactions page", () => {
	test("renders Transactions heading and New Transaction button", async ({
		request,
	}) => {
		const email = `e2e-tx-page-${TS}@example.com`;
		await registerAndLogin(request, email);

		const body = await (await request.get("/transactions")).text();
		expect(body).toContain("Transactions");
		expect(body).toContain("New Transaction");
	});

	test("shows empty state when user has no transactions", async ({
		request,
	}) => {
		const email = `e2e-tx-empty-${TS}@example.com`;
		await registerAndLogin(request, email);

		const body = await (await request.get("/transactions")).text();
		expect(body).toContain("No transactions yet");
	});

	test("GET /transactions returns HTTP 200", async ({ request }) => {
		const email = `e2e-tx-200-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await request.get("/transactions");
		expect(res.status()).toBe(200);
	});
});

// ── Create transaction ────────────────────────────────────────────────────────

test.describe("Create transaction", () => {
	test("POST /transactions with valid income returns row HTML", async ({
		request,
	}) => {
		const email = `e2e-create-income-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, {
			type: "income",
			amount: "99.99",
			title: "Freelance payment",
		});
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Freelance payment");
		expect(body).toContain("income");
	});

	test("POST /transactions with valid expense returns row HTML", async ({
		request,
	}) => {
		const email = `e2e-create-expense-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, {
			type: "expense",
			amount: "42.50",
			title: "Grocery shopping",
			description: "REWE, Berlin",
		});
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Grocery shopping");
		expect(body).toContain("expense");
	});

	test("successful POST sets HX-Trigger closeTransactionForm and retargets list", async ({
		request,
	}) => {
		const email = `e2e-create-headers-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { title: "Header check" });
		expect(res.headers()["hx-trigger"]).toContain("closeTransactionForm");
		expect(res.headers()["hx-retarget"]).toBe("#transaction-list");
		expect(res.headers()["hx-reswap"]).toBe("afterbegin");
	});

	test("POST /transactions with missing title returns form with error", async ({
		request,
	}) => {
		const email = `e2e-create-notitle-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { title: "" });
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Title is required");
	});

	test("POST /transactions with zero amount returns form with error", async ({
		request,
	}) => {
		const email = `e2e-create-zero-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { amount: "0" });
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Amount must be a positive number");
	});

	test("POST /transactions with negative amount returns form with error", async ({
		request,
	}) => {
		const email = `e2e-create-neg-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { amount: "-10" });
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Amount must be a positive number");
	});

	test("POST /transactions with invalid currency returns form with error", async ({
		request,
	}) => {
		const email = `e2e-create-cur-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { currency: "XYZ" });
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Invalid currency");
	});

	test("POST /transactions with missing date returns form with error", async ({
		request,
	}) => {
		const email = `e2e-create-nodate-${TS}@example.com`;
		await registerAndLogin(request, email);

		const res = await createTransaction(request, { date: "" });
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Date is required");
	});
});

// ── Amount storage ────────────────────────────────────────────────────────────

test.describe("Amount storage", () => {
	test("income is stored as positive amount and appears in list", async ({
		request,
	}) => {
		const email = `e2e-amt-income-${TS}@example.com`;
		await registerAndLogin(request, email);
		await createTransaction(request, {
			type: "income",
			amount: "800.00",
			title: "Invoice paid",
		});

		const body = await (await request.get("/transactions")).text();
		expect(body).toContain("Invoice paid");
		expect(body).toContain("income");
	});

	test("expense is stored as negative amount and appears in list", async ({
		request,
	}) => {
		const email = `e2e-amt-expense-${TS}@example.com`;
		await registerAndLogin(request, email);
		await createTransaction(request, {
			type: "expense",
			amount: "25.00",
			title: "Bus ticket",
		});

		const body = await (await request.get("/transactions")).text();
		expect(body).toContain("Bus ticket");
		expect(body).toContain("expense");
	});
});

// ── Edit transaction ──────────────────────────────────────────────────────────

test.describe("Edit transaction", () => {
	test("GET /transactions/:id/edit returns pre-filled form", async ({
		request,
	}) => {
		const email = `e2e-edit-form-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, {
			title: "Original title",
			amount: "55.00",
		});

		// Get the transaction ID from the list page
		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		expect(idMatch).not.toBeNull();
		const txId = idMatch![1];

		const res = await request.get(`/transactions/${txId}/edit`);
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Original title");
		expect(body).toContain("55.00");
		expect(body).toContain("Edit Transaction");
	});

	test("PUT /transactions/:id with valid data returns updated row", async ({
		request,
	}) => {
		const email = `e2e-edit-update-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, { title: "Before edit" });

		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		const res = await request.put(`/transactions/${txId}`, {
			form: {
				type: "expense",
				amount: "15.00",
				currency: "EUR",
				title: "After edit",
				date: "2026-02-25T12:00",
			},
		});
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("After edit");
		expect(body).toContain("expense");
		expect(res.headers()["hx-trigger"]).toContain("closeTransactionForm");
	});

	test("PUT /transactions/:id with invalid data returns form with error", async ({
		request,
	}) => {
		const email = `e2e-edit-invalid-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, { title: "To edit" });

		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		const res = await request.put(`/transactions/${txId}`, {
			form: {
				type: "income",
				amount: "",
				currency: "EUR",
				title: "",
				date: "2026-02-25T12:00",
			},
		});
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("required");
	});

	test("PUT /transactions/:id for another user's transaction returns 404", async ({
		playwright,
	}) => {
		const base = { baseURL: "http://localhost:3000" };
		const ctxA = await playwright.request.newContext(base);
		const ctxB = await playwright.request.newContext(base);

		const emailA = `e2e-edit-iso-a-${TS}@example.com`;
		const emailB = `e2e-edit-iso-b-${TS}@example.com`;
		await registerAndLogin(ctxA, emailA, "User A");
		await registerAndLogin(ctxB, emailB, "User B");

		// A creates a transaction
		await createTransaction(ctxA, { title: "A's private tx" });
		const listBody = await (await ctxA.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		// B tries to edit A's transaction
		const res = await ctxB.put(`/transactions/${txId}`, {
			form: {
				type: "income",
				amount: "1.00",
				currency: "EUR",
				title: "Hijacked",
				date: "2026-02-25T12:00",
			},
		});
		expect(res.status()).toBe(404);

		await ctxA.dispose();
		await ctxB.dispose();
	});
});

// ── Delete transaction ────────────────────────────────────────────────────────

test.describe("Delete transaction", () => {
	test("GET /transactions/:id/confirm-delete returns confirmation strip", async ({
		request,
	}) => {
		const email = `e2e-del-confirm-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, { title: "To confirm delete" });

		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		const res = await request.get(`/transactions/${txId}/confirm-delete`);
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("To confirm delete");
		expect(body).toContain("Delete");
		expect(body).toContain("Cancel");
	});

	test("DELETE /transactions/:id removes the transaction", async ({
		request,
	}) => {
		const email = `e2e-del-remove-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, { title: "Will be deleted" });

		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		const delRes = await request.delete(`/transactions/${txId}`);
		expect(delRes.status()).toBe(200);

		// Transaction no longer appears in the list
		const afterBody = await (await request.get("/transactions")).text();
		expect(afterBody).not.toContain(txId);
	});

	test("DELETE /transactions/:id for another user's transaction is a no-op", async ({
		playwright,
	}) => {
		const base = { baseURL: "http://localhost:3000" };
		const ctxA = await playwright.request.newContext(base);
		const ctxB = await playwright.request.newContext(base);

		const emailA = `e2e-del-iso-a-${TS}@example.com`;
		const emailB = `e2e-del-iso-b-${TS}@example.com`;
		await registerAndLogin(ctxA, emailA, "Del User A");
		await registerAndLogin(ctxB, emailB, "Del User B");

		// A creates a transaction
		await createTransaction(ctxA, { title: "Protected income tx" });
		const listBody = await (await ctxA.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		// B tries to delete A's transaction
		await ctxB.delete(`/transactions/${txId}`);

		// A's transaction is still there
		const afterBody = await (await ctxA.get("/transactions")).text();
		expect(afterBody).toContain("Protected income tx");

		await ctxA.dispose();
		await ctxB.dispose();
	});

	test("GET /transactions/:id/row restores the row after cancel", async ({
		request,
	}) => {
		const email = `e2e-row-restore-${TS}@example.com`;
		await registerAndLogin(request, email);

		await createTransaction(request, { title: "Restored tx" });

		const listBody = await (await request.get("/transactions")).text();
		const idMatch = listBody.match(/id="tx-([^"]+)"/);
		const txId = idMatch![1];

		const res = await request.get(`/transactions/${txId}/row`);
		expect(res.status()).toBe(200);
		const body = await res.text();
		expect(body).toContain("Restored tx");
		expect(body).toContain(`id="tx-${txId}"`);
	});
});

// ── Data isolation ────────────────────────────────────────────────────────────

test.describe("Data isolation", () => {
	test("user A's transactions are not visible to user B", async ({
		playwright,
	}) => {
		const base = { baseURL: "http://localhost:3000" };
		const ctxA = await playwright.request.newContext(base);
		const ctxB = await playwright.request.newContext(base);

		const emailA = `e2e-iso-tx-a-${TS}@example.com`;
		const emailB = `e2e-iso-tx-b-${TS}@example.com`;
		await registerAndLogin(ctxA, emailA, "Alice Isolation");
		await registerAndLogin(ctxB, emailB, "Bob Isolation");

		await createTransaction(ctxA, { title: "Alice private income" });

		const bodyB = await (await ctxB.get("/transactions")).text();
		expect(bodyB).not.toContain("Alice private income");

		const bodyA = await (await ctxA.get("/transactions")).text();
		expect(bodyA).toContain("Alice private income");

		await ctxA.dispose();
		await ctxB.dispose();
	});
});

// ── Pagination ────────────────────────────────────────────────────────────────

test.describe("Pagination", () => {
	test("no pagination shown when transactions <= 20", async ({ request }) => {
		const email = `e2e-pag-none-${TS}@example.com`;
		await registerAndLogin(request, email);

		// Create 3 transactions
		for (let i = 0; i < 3; i++) {
			await createTransaction(request, { title: `Tx ${i}` });
		}

		const body = await (await request.get("/transactions")).text();
		// Pagination only rendered when TotalPages > 1 (need >20 transactions)
		expect(body).not.toContain("Page 1 of");
	});

	test("pagination appears with more than 20 transactions", async ({
		request,
	}) => {
		const email = `e2e-pag-show-${TS}@example.com`;
		await registerAndLogin(request, email);

		// Create 22 transactions
		for (let i = 0; i < 22; i++) {
			await createTransaction(request, { title: `Paged tx ${i}` });
		}

		const body = await (await request.get("/transactions")).text();
		expect(body).toContain("Page 1 of");
	});

	test("page 2 shows the correct transactions", async ({ request }) => {
		const email = `e2e-pag-p2-${TS}@example.com`;
		await registerAndLogin(request, email);

		// Create 21 transactions; oldest title will be on page 2 (sorted newest first)
		for (let i = 0; i < 21; i++) {
			await createTransaction(request, { title: `Sequential tx ${i}` });
		}

		const page2Body = await (await request.get("/transactions?page=2")).text();
		expect(page2Body).toContain("Sequential tx");
	});
});
