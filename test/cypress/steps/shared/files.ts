export const getUploadButton = () => {
    return cy.get("[data-test=upload]");
};

export const getFileBrowser = () => {
    return cy.get("input[type=file][data-test=file]");
};

export const getFileLinks = () => {
    return cy.get("[data-test=file_table] tbody tr [data-test=file_link]");
}

export const getFileAccess = () => {
    return cy.get("[data-test=file_table] tbody tr [data-test=access]");
}

export const getFileExpirations = () => {
    return cy.get("[data-test=file_table] tbody tr [data-test=expires]");
}

export const getPasswordField = () => {
    return cy.get("[data-test=password]");
}

export const getExpirationDateField = () => {
    return cy.get("[data-test=expiration_date]");
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

export const verifyDownloadedFile = (file: string) => {
    const filename = filepathBase(file);
    cy.readFile(file).then((contents) => {
        cy.readFile(`${downloadsFolder}/${filename}`).should('eq', contents);
    })
};

export const shouldSeeUploadSuccessMessage = (filepath: string) => {
    const filename = filepathBase(filepath);
    getMessage().should("contain", `Successfully uploaded ${filename}`);
}