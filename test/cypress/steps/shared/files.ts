export const getUploadButton = () => {
    return cy.get("[data-test=upload]");
};

export const getFileBrowser = () => {
    return cy.get("input[type=file][data-test=file]");
};

export const fileRowsSelector = "[data-test=file_table] tbody tr";

export const getFileLinks = () => {
    return cy.get(`${fileRowsSelector} [data-test=file_link]`);
}

export const getFileAccess = () => {
    return cy.get(`${fileRowsSelector} [data-test=access]`);
}

export const getFileDownloads = () => {
    return cy.get(`${fileRowsSelector} [data-test=downloads]`);
}

export const getFileExpirations = () => {
    return cy.get(`${fileRowsSelector} [data-test=expires]`);
}

export const getPasswordField = () => {
    return cy.get("[data-test=password]");
}

export const getDownloadLimitField = () => {
    return cy.get("[data-test=download-limit]");
}

export const getExpirationDateField = () => {
    return cy.get("[data-test=expiration_date]");
}

export const getExpiresInField = () => {
    return cy.get("[data-test=expire_in]");
}

export const getExpiresInUnitField = () => {
    return cy.get("[data-test=expire_in_unit]");
}

export const getMessage = () => {
    return cy.get("[data-test=message]");
}

export const visitFileUploadPage = () => {
    cy.visit("/");
}

export const visitFileListPage = () => {
    cy.visit("/");
}

export const filepathBase = (filename: string) => {
    return filename.replaceAll("\\", "/").split("/").pop();
}

export const downloadsFolder = "cypress/downloads";

export const deleteDownloadsFolder = () => {
    cy.task('deleteFolder', downloadsFolder);
};

export const verifyFileResponse = (file: string, response: any) => {
    cy.readFile(file).then((contents) => {
        expect(response.body).to.equal(contents);
    });
}

export const verifyDownloadedFile = (file: string) => {
    const filename = filepathBase(file);
    cy.readFile(file).then((contents) => {
        cy.readFile(`${downloadsFolder}/${filename}`).should('eq', contents);
    });
};

export const shouldSeeUploadSuccessMessage = (filepath: string) => {
    const filename = filepathBase(filepath);
    getMessage().should("contain", `Successfully uploaded ${filename}`);
}