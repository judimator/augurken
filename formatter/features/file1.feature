@tag1 @tag2
Feature: A Feature
  Description
  Description

  # A comment
  # A second comment
  Background:
    # A comment
    # A second comment
    Given some background
    Given some datas
      """
      1
      2
      3
      """
    Given a doc string
      # A comment
      # A second comment
      """
      Hello world
      """
    Given a table
      # A comment
      # A second comment
      | Hello | world    | ! |
      | Hello | universe | ! |

  # A comment
  # A second comment
  @tag3 @tag4
  Scenario: A scenario to realize
    When doing something

    Then something happens

    Then we found it equals :
      | 1 | 2 | 3 |
      | 4 | 5 | 6 |

  # A comment
  # A second comment
  Scenario Outline: An outline scenario
    When something happens
    And something else happens
    But something different happens

    Examples:
      | first row | second row |
      | 1         | 2          |

  # A comment
  # A second comment
  Scenario: A scenario with json object with placeholders in doc string
    When something happens
    And something else happens
    But something different happens:
      """
      {
        <this>,
        "some": <value>,
        <that>,
        <that>
        "this": "this",
        "that": <this>,
        "another": {
          <this>,
          <that>,
          "this": "this",
          "that": <this>
        }
      }
      """

  # A comment
  # A second comment
  Scenario: A scenario json array with placeholders in doc string
    When something happens
    And something else happens
    But something different happens:
      """
      [
        <this>,
        <that>,
        "this",
        <some>
        <any>
      ]
      """
