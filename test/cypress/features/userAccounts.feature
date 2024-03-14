Feature: User Accounts
  The admin can manually create and delete user accounts

Background:
  Given I am logged in as the admin
  And there are no other users registered

Scenario: Create a new user with a valid username and password
  Given I am on the user list page

  When I create a new user with the username "testuser" and password "password12345678"

  Then I should see a user created success message
  And I should see the user in the list

Scenario: Admin cannot create a user with a duplicate username
  Given I am on the user list page
  And a user with the name "testuser" exists

  When I create a new user with the username "testuser" and password "password12345678"

  Then I should see a duplicate username failure message

Scenario: Admin cannot create a user with the same username as the admin
  Given I am on the user list page

  When I create a new user with the username "admin" and password "password12345678"

  Then I should see a reserved username failure message

Scenario: Admin can delete users
  Given I am on the user list page
  And a user with the name "testuser1" exists
  And a user with the name "testuser2" exists

  When I delete users "testuser1" and "testuser2"

  Then I should see a 2 users deleted success message
  And I should see an empty user list

@implementing
Scenario: Newly created user can log in
  Given I have created a new user "testuser" with the password "password12345678"

  When I log out
  And log in with the username "testuser" and password "password12345678"

  Then I should successfully log in

@pending
Scenario Outline: Admin cannot create a user with an invalid password
  Given I am on the user list page

  When I create a new user with the username "testuser" and password <password>

  Then I should see an invalid password failure message

  Examples:
    |          password |
    |                "" |
    | "password1234567" |
