Feature:

  Scenario: Running command
    When I run the command "greet" with the arguments "--name US"

    Then the command "greet" should output "Hello US"