Feature: File Upload
  An authenticated user can upload a file and download it again.

Background:
  Given I am logged in
  And I am on the file upload page

@pending @selected
Scenario: Cannot upload a file without choosing one
  When I browse for a file to upload
  And I cancel the browse file dialog
  Then I should not be able to upload a file

@pending
Scenario: Upload a small file
  When I browse for a file to upload
  And I select "small-file.txt" from the browse file dialog
  And I upload the file
  Then I should see a file upload success message

@pending
Scenario: Download a small file
  When I upload the file "small-file.txt"
  And I select the top file from the list
  Then I should download the file "small-file.txt"