Feature: File Upload
  An authenticated user can upload a file and download it again.

Background:
  Given I am logged in

Scenario: Cannot upload a file without choosing one
  When I go to the file upload page
  Then I should not be able to upload a file

Scenario: Upload a small file
  Given I am on the file upload page
  And I have a file "files/small.txt"
  When I browse for a file to upload
  And I select it from the browse file dialog
  And I upload the file
  Then I should see a file upload success message

@pending
Scenario: Download a small file
  Given I am on the file upload page
  And I have a file "files/small.txt"
  When I upload the file
  And I select the top file from the list
  Then I should download the file