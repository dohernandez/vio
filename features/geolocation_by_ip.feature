Feature: Expose geolocation information by IP
  As a user, I want to know the geolocation information of an IP address.

  Background:
    Given there is a clean "postgres" database
    And these rows are stored in table "geolocation" of database "postgres":
      | ip_address     | country_code | country      | city         | latitude           | longitude           | mystery_value |
      | 200.106.141.15 | SI           | Nepal        | DuBuquemouth | -84.87503094689836 | 7.206435933364332   | 7823011346    |
      | 160.103.7.140  | CZ           | Nicaragua    | New Neva     | -68.31023296602508 | -37.62435199624531  | 7301823115    |
      | 70.95.73.73    | TL           | Saudi Arabia | Gradymouth   | -49.16675918861615 | -86.05920084416894  | 2559997162    |
      | 125.159.20.54  | LI           | Guyana       | Port Karson  | -78.2274228596799  | -163.26218895343357 | 1337885276    |

  Scenario: Expose geolocation information by IP successfully
    When I request HTTP endpoint with method "GET" and URI "/v1/geolocations/200.106.141.15"

    Then I should have response with status "OK"
    And I should have response with header "Content-Type: application/json"
    And I should have response with body
    """
    {
        "ip_address": "200.106.141.15",
        "country_code": "SI",
        "country": "Nepal",
        "city": "DuBuquemouth",
        "latitude": -84.87503094689836,
        "longitude": 7.206435933364332,
        "mystery_value": 7823011346
    }
    """

  Scenario: Expose geolocation information by IP not found
    When I request HTTP endpoint with method "GET" and URI "/v1/geolocations/160.168.85.54"

    Then I should have response with status "Not Found"

  Scenario: Expose geolocation information by IP with invalid IP
    When I request HTTP endpoint with method "GET" and URI "/v1/geolocations/invalid-ip"

    Then I should have response with status "Bad Request"