import {fileRowsSelector} from "./files";

export const getUsername = () => {
    return cy.get("[data-test=create_user_form] [data-test=username]");
};

export const getPassword = () => {
    return cy.get("[data-test=create_user_form] [data-test=password]");
};

export const getMessage = () => {
    return cy.get("[data-test=message]");
};

export const getCreateUserButton = () => {
    return cy.get("[data-test=create_user_form] [data-test=create_user]");
};

export const userRowsSelector = "[data-test=user_table] tbody tr";

export const getUsernames = () => {
    return cy.get(`${userRowsSelector} [data-test=username]`);
};

export const getDeleteCheckboxForUsername = (username: string) => {
    return cy.get(`${userRowsSelector} [data-test=username] ~ [data-test=delete_select]`);
};

export const getDeleteUsersButton = () => {
    return cy.get("[data-test=delete_users_form] [data-test=delete_user]");
};
