import {Given,When,Then} from "@badeball/cypress-cucumber-preprocessor";
import {login} from "./shared/login";
import {
    getUsername,
    getPassword,
    getCreateUserButton,
    getUsernames, getMessage
} from "./shared/users";

const state: {
    user?: string,
} = {};

Given("I am logged in as the admin", () => {
    login("{admin}", "{admin}");
});

Given("there are no other users registered", () => {
    cy.request({
        method: "POST",
        url: "/test-automation",
        form: true,
        body: {
            delete_all_users: true,
        },
    }).then(response => {
        cy.log(JSON.stringify(response.headers));
    });
});

Given("I am on the user list page", () => {
   cy.visit("/users");
});

Given("a user with the name {string} already exists", (user: string) => {
    const validPassword = "password12345678";
    createUser(user, validPassword)
});

const createUser = (user: string, pass: string) => {
    getUsername().type(user);
    getPassword().type(pass);
    getCreateUserButton().click();
}

When("I create a new user with the username {string} and password {string}", (user: string, pass: string) => {
    createUser(user, pass);
    state.user = user;
});

Then("I should see a user created success message", () => {
    getMessage().should("contain", `Created new user "${state.user}"`);
});

Then("I should see the user in the list", () => {
    getUsernames().first().should('contain.text', state.user);
});

Then("I should see a duplicate username failure message", () => {
    getMessage().should("contain", `Username "${state.user}" already exists.`);
});