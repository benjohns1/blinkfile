import { Before, Given, When, Then } from "@badeball/cypress-cucumber-preprocessor";
import { login } from "./shared/login";
import path from "path";

Before(() => {
    deleteDownloadsFolder();
});

const state: {
    fileBrowser?: any,
    fileToUpload?: any,
    fileLink?: any,
} = {};

const getUploadButton = () => {
    return cy.get("[data-test=upload]");
};

const getFileBrowser = () => {
    return cy.get("input[type=file][data-test=file]");
};

const downloadsFolder = Cypress.config('downloadsFolder');

const deleteDownloadsFolder = () => {
    cy.task('deleteFolder', downloadsFolder)
};

const verifyFileDownloaded = (file: string) => {
    const filename = filepathBase(file);
    const wantContents = cy.readFile(file);
    cy.get(path.join(downloadsFolder, filename)).should('eq', wantContents);
};

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
    state.fileBrowser = getFileBrowser();
});

When("I select it from the browse file dialog", () => {
    state.fileBrowser.selectFile(state.fileToUpload);
});

When("I upload the file", () => {
    getFileBrowser().selectFile(state.fileToUpload);
    getUploadButton().click();
});

const filepathBase = (filename: string) => {
    return filename.replaceAll("\\", "/").split("/").pop();
}

When("I should see a file upload success message", () => {
    const filename = filepathBase(state.fileToUpload);
    cy.get("[data-test=message]").should("contain", `Successfully uploaded ${filename}`);
});

When("I select the top file from the list", () => {
    state.fileLink = cy.get("[data-test=file_table] tbody tr [data-test=file_link]").first();
});

Then("I should download the file", () => {
    state.fileLink.click();
    verifyFileDownloaded(state.fileToUpload);
});