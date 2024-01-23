import {Given, When, Then} from "@badeball/cypress-cucumber-preprocessor";
import {
    getFileBrowser,
    getFileLinks,
    getUploadButton,
    getMessage,
    visitFileUploadPage,
    getExpirationDateField,
    verifyDownloadedFile,
    deleteDownloadsFolder,
    filepathBase,
    getFileAccess, shouldSeeUploadSuccessMessage, getFileExpirations
} from "./shared/files";
import {login, logout} from "./shared/login";
import dayjs from "dayjs";

const state: {
    fileToUpload?: any,
    fileLink?: any,
} = {};

Given("I am logged in", () => {
    login("{admin}", "{admin}");
});

Given("I have selected the file {string} to upload", (name: string) => {
    visitFileUploadPage();
    state.fileToUpload = `features/${name}`;
    deleteDownloadsFolder();
    getFileBrowser().selectFile(state.fileToUpload);
});

When("I enter {string} for the expiration date", (date: string) => {
    let expirationDate: dayjs.Dayjs;
    if (date === "today") {
        expirationDate = dayjs();
    } else {
        throw `date string ${date} not implemented`;
    }
    getExpirationDateField().type(expirationDate.format("MM/DD/YYYY"));
});

When("I upload the file", () => {
    getUploadButton().click();
    getFileLinks().first().invoke("attr", "href").then(href => {
        state.fileLink = href;
    });
});

Then("it should upload successfully", () => {
    shouldSeeUploadSuccessMessage(state.fileToUpload);
    const filename = filepathBase(state.fileToUpload);
    getFileLinks().first().should('contain.text', filename);
});

Then("it should look like it is set to expire {string}", (expiration: string) => {
    let format: string;
    if (expiration === "tomorrow at midnight") {
        format = dayjs().add(1, "day").format("MM/DD/YYYY") + " 12:00 AM";
    } else {
        throw `expiration string ${expiration} not implemented`;
    }
    getFileExpirations().first().should('contain.text', format);
});
