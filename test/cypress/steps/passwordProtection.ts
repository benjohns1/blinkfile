import {Given, When} from "@badeball/cypress-cucumber-preprocessor";
import {deleteDownloadsFolder, filepathBase, getFileBrowser, getUploadButton} from "./shared/files";
import {login} from "./shared/login";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

Given("I am on the file upload page", () => {
    cy.visit("/");
});

Given("I have a file {string}", (name: string) => {
    const filename = `features/${name}`;
    cy.readFile(filename);
    state.fileToUpload = filename;
});

When("I browse for the file to upload", () => {
    getFileBrowser().selectFile(state.fileToUpload);
});

When("I enter the password {string}", (password: string) => {
    cy.get("[data-test=password]").type(password);
});

When("I upload the file", () => {
    getUploadButton().click();
});

When("I should see a file upload success message", () => {
    const filename = filepathBase(state.fileToUpload);
    cy.get("[data-test=message]").should("contain", `Successfully uploaded ${filename}`);
});