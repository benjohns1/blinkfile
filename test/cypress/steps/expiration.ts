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
    getFileAccess, shouldSeeUploadSuccessMessage, getFileExpirations, getExpiresInField, getExpiresInUnitField
} from "./shared/files";
import {login, logout} from "./shared/login";
import dayjs from "dayjs";

const state: {
    fileToUpload?: any,
    fileLink?: any,
    startUpload?: Date,
    endUpload?: Date,
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

When("I set it to expire in {string}", (expiresIn: string) => {
    let expiresInAmount: string;
    let expiresInUnit: string;
    if (expiresIn === "3 days") {
        expiresInAmount = "3";
        expiresInUnit = "d";
    } else {
        throw `expiresIn string ${expiresIn} not implemented`;
    }
    getExpiresInField().type(expiresInAmount);
    getExpiresInUnitField().select(expiresInUnit);
});

When("I upload the file", () => {
    state.startUpload = new Date();
    getUploadButton().click().then(() => {
        state.endUpload = new Date();
    });
    getFileLinks().first().invoke("attr", "href").then(href => {
        state.fileLink = href;
    });
});

Then("it should upload successfully", () => {
    shouldSeeUploadSuccessMessage(state.fileToUpload);
    const filename = filepathBase(state.fileToUpload);
    getFileLinks().first().should('contain.text', filename);
});

const allowedClockDriftMinutes = 1;

Then("it should look like it is set to expire {string}", (expiration: string) => {
    getFileExpirations().first().invoke('text').as('expires');
    if (expiration === "tomorrow at midnight") {
        const format = dayjs().add(1, "day").format("MM/DD/YYYY") + " 12:00 AM";
        cy.get('@expires').should('contain', format);
        return;
    }
    if (expiration === "3 days from now") {
        const earliest = dayjs(state.startUpload).add(3, "day").subtract(allowedClockDriftMinutes, "minute").toDate();
        const latest = dayjs(state.endUpload).add(3, "day").add(allowedClockDriftMinutes, "minute").toDate();
        cy.get('@expires').then($expires => {
            const actual = dayjs($expires.toString(), "MM/DD/YYYY HH:mm A").toDate();
            const cyDate = cy.wrap(actual);
            cyDate.should("be.at.least", earliest);
            cyDate.should("be.at.most", latest);
        });
        return;
    }
    throw `expiration string ${expiration} not implemented`;
});
