Feature: Admin Login
  The initial admin user can login with the configured username and password credentials.

Background:
  Given I am on the login page

Scenario: Login attempt with an invalid username
  When I login with username "an invalid username" and password "{admin}"
  Then I should see a generic authentication failed message

Scenario: Login attempt with an invalid password
  When I login with username "{admin}" and password "an invalid password"
  Then I should see a generic authentication failed message

Scenario: Login success
  When I login with username "{admin}" and password "{admin}"
  Then I should be logged in successfully
