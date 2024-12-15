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

Scenario: Admin can delete users
  Given I am on the user list page
  And a user with the name "testuser1" exists
  And a user with the name "testuser2" exists
  When I delete users "testuser1" and "testuser2"
  Then I should see a 2 users deleted success message
  And I should see an empty user list

Scenario: Newly created user can log in
  Given I have created a new user "testuser" with the password "password12345678"
  When I log out
  And log in with the username "testuser" and password "password12345678"
  Then I should successfully log in

Scenario Outline: Admin cannot create a user with invalid credentials
  Given I am on the user list page
  When I create a new user with the username <username> and password <password>
  Then I should see failure message text <message>
  Examples:
    | username   |           password |                                              message |
    | "admin"    | "password12345678" | "Username \"admin\" is reserved and cannot be used." |
    | "testuser" |        "too-short" |       "password must be at least 16 characters long" |

Scenario: Only the admin can manage users
  Given I have created a new user "testuser" with the password "password12345678"
  When I log out
  And log in with the username "testuser" and password "password12345678"
  Then I should successfully log in
  And I should not see the users link
  And I should not be able to access the user list page

Scenario: Admin can change a user's username
  Given I have created a new user "testuser1" with the password "password12345678"
  When I edit user "testuser1"
  And I update their username to "testuser2"
  Then I should see a username changed success message

Scenario: Admin can change a user's password
  Given I have created a new user "testuser1" with the password "old_password12345678"
  When I edit user "testuser1"
  And I update their password to "new_pass12345678"
  Then I should see a password changed success message

Scenario: User can log in with newly updated credentials
  Given I have created a new user "testuser1" with the password "old_pass12345678"
  And I have updated the user "testuser1" to have username "testuser2" and password "new_pass12345678"
  When I log out
  And log in with the username "testuser2" and password "new_pass12345678"
  Then I should successfully log in