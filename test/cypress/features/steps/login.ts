import { Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";

const adminUsername = "admin";
const adminPassword = "1234123412341234"

Given("I am on the login page", () => {
    cy.visit("/login")
});

When("I login with username {string} and password {string}", (username: string, password: string) => {
    if (username === "{admin}") {
        username = adminUsername;
    }
    if (password === "{admin}") {
        password = adminPassword;
    }
    cy.get("[data-test=username]").type(username);
    cy.get("[data-test=password]").type(`${password}{enter}`);
});

Then("I should see a generic authentication failed message", () => {
    cy.get("[data-test=message]").should("contain", "Authentication failed");
    cy.get("[data-test=message]").should("contain", "Your credentials were not correct.");
});

Then("I should be logged in successfully", () => {
    cy.location("pathname").should("equal", "/");
});
