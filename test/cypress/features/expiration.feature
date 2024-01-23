Feature: Expiration
  Uploaded files can be set to automatically expire after a period of time.

Background:
  Given I am logged in

Scenario: Upload a file with a future expiration date
  Given I have selected the file "files/expiration.txt" to upload
  When I enter "today" for the expiration date
  And I upload the file
  Then it should upload successfully
  And it should look like it is set to expire "tomorrow at midnight"