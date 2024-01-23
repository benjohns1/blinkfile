Feature: Password Protection
  Uploaded files can be protected by a password.

Background:
  Given I am logged in

Scenario: Upload a file with a password
  Given I am on the file upload page
  And I have a file "files/password-protect.txt"
  When I browse for the file to upload
  And I enter the password "12345"
  And I upload the file
  Then I should see a file upload success message
  And I should see the file at the top of the list
  And it should look like it is password protected

Scenario: Owner can download a password-protected file without needing a password
  Given I have uploaded a file "files/password-protect.txt" with the password "12345"
  And I am on the file list page
  When I download the top file from the list
  Then I should download the file without needing a password

Scenario: A user cannot download a password-protected file without the correct password
  Given I have uploaded a file "files/password-protect.txt" with the password "12345"
  When I log out
  And I download the file with the password "invalid-password"
  Then I should see an invalid password message

Scenario: A user can download a password-protected file with the correct password
  Given I have uploaded a file "files/password-protect.txt" with the password "12345"
  When I log out
  And I download the file with the password "12345"
  Then I should download the file