Feature: Expiration
  Uploaded files can be set to automatically expire after a period of time.

Background:
  Given I am logged in
  And there are no uploaded files

Scenario: Upload a file with a future expiration date
  Given I have selected the file "files/expiration.txt" to upload
  When I enter "today" for the expiration date
  And I upload the file
  Then it should upload successfully
  And it should look like it is set to expire "tomorrow at midnight"

Scenario: Upload a file that expires in a timeframe
  Given I have selected the file "files/expiration.txt" to upload
  When I set it to expire in "3 days"
  And I upload the file
  Then it should upload successfully
  And it should look like it is set to expire "3 days from now"

Scenario: A file is removed after it expires
  Given I have uploaded a file "files/expiration.txt" set to expire in "3 days"
  And I can successfully download the file
  When "3 days" has passed
  Then I can no longer download the file
  And it no longer shows up in the file list

Scenario: Cannot upload a file that expires in the past
  Given I have selected the file "files/expiration.txt" to upload
  When I set it to expire in "-1 minutes"
  And I try to upload the file
  Then I should see an error message that contains "Cannot upload a file that expires in the past"