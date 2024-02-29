import {Given} from "@badeball/cypress-cucumber-preprocessor";
import {login} from "./shared/login";

Given("I am logged in as the admin", () => {
    login("{admin}", "{admin}");
});