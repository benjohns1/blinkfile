import { Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";

Given("I am on the file upload page", () => {
    cy.visit("/");
});

Given("I have a file {string}", (name: string, contents: string) => {
    cy.log("LOGME", name, contents);
});

When("I go to the file upload page", () => {
    cy.visit("/");
});

const state: {
    fileBrowser?: any
} = {};

When("I browse for a file to upload", () => {
    state.fileBrowser = cy.get("input[type=file][data-test=file]");
});

Then("I should not be able to upload a file", () => {
    cy.get("[data-test=upload]").should('be.disabled');
});