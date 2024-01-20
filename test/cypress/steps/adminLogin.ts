import { Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";
import { login } from "./shared/login";

Given("I am on the login page", () => {
    cy.visit("/login")
});

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

When("I login with username {string} and password {string}", (username: string, password: string) => {
    login(username, password);
});

Then("I should see a generic authentication failed message", () => {
    cy.get("[data-test=message]")
        .should("contain", "Authentication failed")
        .should("contain", "Your credentials were not correct.");
});

Then("I should be logged in successfully", () => {
    cy.location("pathname").should("equal", "/");
});
