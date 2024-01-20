import {Given} from "@badeball/cypress-cucumber-preprocessor";

const adminUsername = "admin";
const adminPassword = "1234123412341234"

export const login = (username: string, password: string) => {
    cy.visit("/login");
    if (username === "{admin}") {
        username = adminUsername;
    }
    if (password === "{admin}") {
        password = adminPassword;
    }
    cy.get("[data-test=username]").type(username);
    cy.get("[data-test=password]").type(`${password}{enter}`);
}
