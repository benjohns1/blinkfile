import { Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";
import { login } from "./shared/login";
import {
    deleteDownloadsFolder,
    getUploadButton,
    getFileBrowser,
    filepathBase,
    verifyDownloadedFile,
    getFileLinks, getMessage, shouldSeeUploadSuccessMessage
} from "./shared/files";

const state: {
    fileToUpload?: any,
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
    shouldSeeUploadSuccessMessage(state.fileToUpload);
});

When("I download the top file from the list", () => {
    getFileLinks().first().click();
});

Then("I should download the file", () => {
    verifyDownloadedFile(state.fileToUpload);
});