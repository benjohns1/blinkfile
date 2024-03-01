import {Given,When,Then} from "@badeball/cypress-cucumber-preprocessor";
import {login} from "./shared/login";
import {
    getUsername,
    getPassword,
    shouldSeeCreatedSuccessMessage,
    getCreateUserButton,
    getUsernames
} from "./shared/users";

const state: {
    user?: string,
} = {};

Given("I am logged in as the admin", () => {
    login("{admin}", "{admin}");
});

Given("I am on the user list page", () => {
   cy.visit("/users");
});

When("I create a new user with the username {string} and password {string}", (user: string, pass: string) => {
    getUsername().type(user);
    getPassword().type(pass);
    state.user = user;
    getCreateUserButton().click();
});

Then("I should see a user created success message", () => {
    shouldSeeCreatedSuccessMessage(state.user);
});

Then("I should see the user in the list", () => {
    getUsernames().first().should('contain.text', state.user);
});