import {Given, When, Then} from "@badeball/cypress-cucumber-preprocessor";
import {
    filepathBase,
    getFileAccess,
    getFileBrowser,
    getFileLinks,
    getPasswordField,
    getUploadButton,
    getMessage,
    visitFileUploadPage,
    visitFileListPage,
    verifyDownloadedFile,
    deleteDownloadsFolder,
    shouldSeeUploadSuccessMessage
} from "./shared/files";
import {login, logout} from "./shared/login";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

Given("I am on the file upload page", () => {
    visitFileUploadPage();
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
    getPasswordField().type(password);
});

When("I upload the file", () => {
    getUploadButton().click();
});

Then("I should see a file upload success message", () => {
    shouldSeeUploadSuccessMessage(state.fileToUpload);
});

Then("I should see the file at the top of the list", () => {
    const filename = filepathBase(state.fileToUpload);
    getFileLinks().first().should('contain.text', filename);
});

Then("it should look like it is password protected", () => {
    getFileAccess().first().should('contain.text', "Password");
});

Given("I have uploaded a file {string} with the password {string}", (name: string, password: string) => {
    visitFileUploadPage();
    state.fileToUpload = `features/${name}`;
    deleteDownloadsFolder();
    getFileBrowser().selectFile(state.fileToUpload);
    getPasswordField().type(password);
    getUploadButton().click();
    getFileLinks().first().invoke("attr", "href").then(href => {
        state.fileLink = href;
    });
});

Given("I am on the file list page", () => {
    visitFileListPage();
});

When("I download the top file from the list", () => {
    getFileLinks().first().click();
});

Then("I should download the file without needing a password", () => {
    verifyDownloadedFile(state.fileToUpload);
});

When("I log out", () => {
    logout();
});

When("I download the file with the password {string}", (password: string) => {
    cy.visit(state.fileLink);
    cy.get("[data-test=password]").type(password);
    cy.get("[data-test=download]").click();
});

Then("I should see an invalid password message", () => {
    getMessage()
        .should("contain", "Authorization failed")
        .should("contain", "Invalid password");
});

Then("I should download the file", () => {
    verifyDownloadedFile(state.fileToUpload);
});
