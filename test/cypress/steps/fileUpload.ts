import { Before, Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";
import { login } from "./shared/login";
import { deleteDownloadsFolder, getUploadButton, getFileBrowser, filepathBase, verifyDownloadedFile } from "./shared/files";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

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

When("I upload the file", () => {
    deleteDownloadsFolder();
    getFileBrowser().selectFile(state.fileToUpload);
    getUploadButton().click();
});

When("I should see a file upload success message", () => {
    const filename = filepathBase(state.fileToUpload);
    cy.get("[data-test=message]").should("contain", `Successfully uploaded ${filename}`);
});

When("I select the top file from the list", () => {
    state.fileLink = cy.get("[data-test=file_table] tbody tr [data-test=file_link]").first();
});

Then("I should download the file", () => {
    state.fileLink.click();
    verifyDownloadedFile(state.fileToUpload);
});