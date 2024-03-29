Feature: Download Limit
  User can choose the number of downloads allowed before the file is deleted.

Background:
  Given I am logged in

Scenario: Upload a file with a download limit and show zero downloads
  When I upload a file "files/download-limit.txt" with a download limit of 10
  Then I should see a file upload success message
  And I should see the file at the top of the list
  And I should see a file download count of 0 out of 10

Scenario: Upload a file with a download limit counts the number of downloads
  Given I have uploaded a file "files/download-limit.txt" with a download limit of 10
  When I download the file 5 times
  Then I should see a file download count of 5 out of 10

Scenario: Upload a file without a download limit counts the number of downloads
  Given I have uploaded a file "files/download-limit.txt" without a download limit
  When I download the file 5 times
  Then I should see a file download count of 5

Scenario: The file is removed after downloading it the maximum number of times
  Given I have uploaded a file "files/download-limit.txt" with a download limit of 3
  When I download the file 3 times
  Then I can no longer download the file
  And it no longer shows up in the file list

Scenario: Cannot upload a file that has a negative download limit
  Given I have selected the file "files/download-limit.txt" to upload
  When I set its download limit to -1
  Then I should not be able to upload the file