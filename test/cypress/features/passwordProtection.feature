Feature: Password Protection
  Uploaded files can be protected by a password.

Background:
  Given I am logged in

Scenario: Upload a small file with a password
  Given I am on the file upload page
  And I have a file "files/small.txt"
  When I browse for the file to upload
  And I enter the password "12345"
  And I upload the file
  Then I should see a file upload success message