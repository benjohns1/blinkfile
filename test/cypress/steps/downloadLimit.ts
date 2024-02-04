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
    getFileDownloads
} from "./shared/files";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

When("I upload a file {string} with a download limit of {int}", (name: string, limit: number) => {
    visitFileUploadPage();
    state.fileToUpload = `features/${name}`;
    deleteDownloadsFolder();
    getFileBrowser().selectFile(state.fileToUpload);
    getDownloadLimitField().type(limit.toString());
    getUploadButton().click();
    getFileLinks().first().invoke("attr", "href").then(href => {
        state.fileLink = href;
    });
});

Then("I should see a file upload success message", () => {
    shouldSeeUploadSuccessMessage(state.fileToUpload);
});

Then("I should see the file at the top of the list", () => {
    const filename = filepathBase(state.fileToUpload);
    getFileLinks().first().should('contain.text', filename);
});

Then("I should see a file download count of {int} out of {int}", (count: number, limit: number) => {
    getFileDownloads().first().should('contain.text', `${count}/${limit}`);
});