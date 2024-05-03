Feature: Test Formatter
  Scenario: Comment between rows
    Given some state
      | col a | col b |
      | 1     | 2     |
      # A comment about the row below
      | 3     | 4     |
      # Another comment about the row below
      | 5     | 6     |
      | 7     | 8     |
      | 9     | 10    |
    Given some another state
      | col a | col b |
      | 1     | 2     |
      # A comment about the row below
      | 3     | 4     |
      # Another comment about the row below
      | 5     | 6     |
