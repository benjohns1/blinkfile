export const getUsername = () => {
    return cy.get("[data-test=username]");
};

export const getPassword = () => {
    return cy.get("[data-test=password]");
};

export const shouldSeeCreatedSuccessMessage = (user: string) => {
    getMessage().should("contain", `Created new user "${user}"`);
};

export const getMessage = () => {
    return cy.get("[data-test=message]");
};

export const getCreateUserButton = () => {
    return cy.get("[data-test=create_user]");
};

export const userRowsSelector = "[data-test=user_table] tbody tr";

export const getUsernames = () => {
    return cy.get(`${userRowsSelector} [data-test=username]`);
};