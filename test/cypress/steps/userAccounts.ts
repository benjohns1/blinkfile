import {Given,When,Then} from "@badeball/cypress-cucumber-preprocessor";
import {login} from "./shared/login";
import {
    getUsername,
    getPassword,
    getCreateUserButton,
    getUsernames, getMessage, getDeleteCheckboxForUsername, getDeleteUsersButton
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

Given("a user with the name {string} exists", (user: string) => {
    const validPassword = "password12345678";
    createUser(user, validPassword)
});

Given("I have created a new user {string} with the password {string}", (user: string, pass: string) => {
    cy.visit("/users");
    createUser(user, pass);
});

const createUser = (user: string, pass: string) => {
    if (user !== "") {
        getUsername().type(user);
    }
    if (pass !== "") {
        getPassword().type(pass);
    }
    getCreateUserButton().click();
};

When("I create a new user with the username {string} and password {string}", (user: string, pass: string) => {
    createUser(user, pass);
    state.user = user;
});

When("I delete users {string} and {string}", (user1: string, user2: string) => {
    getDeleteCheckboxForUsername(user1).check();
    getDeleteCheckboxForUsername(user2).check();
    getDeleteUsersButton().click();
});

When("I log out", () => {
    cy.visit("/logout");
});

When("log in with the username {string} and password {string}", (user: string, pass: string) => {
    login(user, pass);
});

Then("I should see a user created success message", () => {
    getMessage().should("contain", `Created new user "${state.user}"`);
});

Then("I should see a {int} users deleted success message", (count: number) => {
    getMessage().should("contain", `Deleted ${count} users.`);
});

Then("I should see a duplicate username failure message", () => {
    getMessage().should("contain", `Username "${state.user}" already exists.`);
});

Then("I should see a reserved username failure message", () => {
    getMessage().should("contain", `Username "${state.user}" is reserved and cannot be used.`);
});

Then("I should see the user in the list", () => {
    getUsernames().first().should('contain.text', state.user);
});

Then("I should see an empty user list", () => {
    getUsernames().should('not.exist');
});

Then("I should successfully log in", () => {
    cy.location("pathname").should("equal", "/");
});

Then("I should see failure message text {string}", (msg: string) => {
    getMessage().should("contain", msg);
});

Then("I should not see the users link", () => {
    cy.get("[data-test=nav] [data-test=users]").should("not.exist");
});

Then("I should not be able to access the user list page", () => {
    cy.request({
        method: "GET",
        url: "/users",
        failOnStatusCode: false,
    }).as("response");
    cy.get("@response").should("have.property", "status", 404);
});

When("I edit user {string}", (user: string) => {
    cy.get(`[data-test=user_edit_link]:contains(${user})`).click();
});

When("I update their username {string}", (user: string) => {
    cy.get("[data-test=username]").clear().type(user);
    cy.get("[data-test=change_username]").click();
    state.user = user;
});

Then("I should see a username changed success message", () => {
    cy.get("[data-test=username]").should("have.value", state.user)
    getMessage().should("contain", `Username changed to "${state.user}"`);
});
