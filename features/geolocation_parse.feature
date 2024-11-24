Feature: Parse geolocation
  As a user, I want to parse the geolocation information.

  Background:
    Given there is a clean "postgres" database

  Scenario: Parse geolocation successfully from file source
    When I run the command "parse" with the arguments "filesystem -f ./resources/sample_data/test_data.csv"

    Then the command "parse" finishes successfully
    And Then these rows are available in table "geolocation" of database "postgres"
      | ip_address     | country_code | country      | city         | latitude           | longitude           | mystery_value |
      | 200.106.141.15 | SI           | Nepal        | DuBuquemouth | -84.87503094689836 | 7.206435933364332   | 7823011346    |
      | 160.103.7.140  | CZ           | Nicaragua    | New Neva     | -68.31023296602508 | -37.62435199624531  | 7301823115    |
      | 70.95.73.73    | TL           | Saudi Arabia | Gradymouth   | -49.16675918861615 | -86.05920084416894  | 2559997162    |
      | 125.159.20.54  | LI           | Guyana       | Port Karson  | -78.2274228596799  | -163.26218895343357 | 1337885276    |