import { Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";
import { login } from "./shared/login";

const state: {
    fileBrowser?: any,
    fileToUpload?: any,
} = {};

const getUploadButton = () => {
    return cy.get("[data-test=upload]");
}

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

When("I go to the file upload page", () => {
    cy.visit("/");
});

Then("I should not be able to upload a file", () => {
    getUploadButton().should('be.disabled');
});

Given("I am on the file upload page", () => {
    cy.visit("/");
});

Given("I have a file {string}", (name: string) => {
    const filename = `features/${name}`;
    cy.readFile(filename);
    state.fileToUpload = filename;
});

When("I browse for a file to upload", () => {
    state.fileBrowser = cy.get("input[type=file][data-test=file]");
});

When("I select it from the browse file dialog", () => {
    state.fileBrowser.selectFile(state.fileToUpload);
});

When("I upload the file", () => {
    getUploadButton().click();
});

When("I should see a file upload success message", () => {
    const filename = state.fileToUpload.replaceAll("\\", "/").split("/").pop();
    cy.get("[data-test=message]").should("contain", `Successfully uploaded ${filename}`);
});
