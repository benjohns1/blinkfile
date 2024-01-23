export const getUploadButton = () => {
    return cy.get("[data-test=upload]");
};

export const getFileBrowser = () => {
    return cy.get("input[type=file][data-test=file]");
};

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