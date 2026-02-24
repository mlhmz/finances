describe("Smoke", () => {
  it("loads the home page", () => {
    cy.visit("/");
    cy.contains("Finances");
  });

  it("greet endpoint returns text", () => {
    cy.request("/greet").its("body").should("include", "Hello");
  });
});
