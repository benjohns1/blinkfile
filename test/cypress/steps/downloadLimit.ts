import {Given, When, Then} from "@badeball/cypress-cucumber-preprocessor";
import {login} from "./shared/login";
import {
    deleteDownloadsFolder,
    getFileBrowser, getFileLinks,
    getDownloadLimitField,
    getUploadButton,
    visitFileUploadPage,
    shouldSeeUploadSuccessMessage,
    filepathBase,
    getFileDownloads,
    verifyFileResponse,
    visitFileListPage, fileRowsSelector, cannotDownloadFileNoPassword, fileNotInList
} from "./shared/files";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

const uploadFile = (name: string, limit: number) => {
    visitFileUploadPage();
    state.fileToUpload = `features/${name}`;
    deleteDownloadsFolder();
    getFileBrowser().selectFile(state.fileToUpload);
    if (limit !== undefined) {
        getDownloadLimitField().type(limit.toString());
    }
    getUploadButton().click();
    getFileLinks().first().invoke("attr", "href").then(href => {
        state.fileLink = href;
    });
};

Given("I have uploaded a file {string} with a download limit of {int}", uploadFile);

Given("I have uploaded a file {string} without a download limit", uploadFile);

When("I upload a file {string} with a download limit of {int}", uploadFile);

When("I download the file {int} times", (count: number) => {
    for (let i = 0; i < count; i++) {
        deleteDownloadsFolder();
        cy.request(state.fileLink).then(response => {
            verifyFileResponse(state.fileToUpload, response);
        });
    }
});

Then("I should see a file upload success message", () => {
    shouldSeeUploadSuccessMessage(state.fileToUpload);
});

Then("I should see the file at the top of the list", () => {
    visitFileListPage();
    const filename = filepathBase(state.fileToUpload);
    getFileLinks().first().should('have.text', filename);
});

Then("I should see a file download count of {int} out of {int}", (count: number, limit: number) => {
    visitFileListPage();
    getFileDownloads().first().should('have.text', `${count}/${limit}`);
});

Then("I should see a file download count of {int}", (count: number) => {
    visitFileListPage();
    getFileDownloads().first().should('have.text', `${count}`);
});

Then("I can no longer download the file", () => {
    cannotDownloadFileNoPassword(state.fileLink);
});

Then("it no longer shows up in the file list", () => {
    fileNotInList(state.fileLink);
});
