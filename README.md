Augurken is a tool to format [Gherkin](https://cucumber.io/docs/gherkin/reference) features.

# Install<a id="install"></a>

Download the latest binary for your architecture [here](https://github.com/judimator/augurken/releases/latest).

# Usage<a id="usage"></a>

```
Usage:
  augurken [command]

Available Commands:
  check       Check formatting of gherkin file(s)
  completion  Generate the autocompletion script for the specified shell
  format      Format gherkin file(s)
  help        Help about any command

Flags:
  -h, --help         help for augurken
  -i, --indent int   set the indentation for Gherkin features (default 2)
  -v, --version      version for augurken
 
Use "augurken [command] --help" for more information about a command.
```

Format a feature file

```shell
$ augurken format /path/to/filename.feature
```

Format a folder with features

```shell
$ augurken format /path/to/features
```

Format a feature file with indent. Augurken uses **space** as indent

```shell
$ augurken format -i 2 /path/to/filename.feature
```

⚠️ Augurken works only on `UTF-8` encoded files, it will detect and convert automatically files that are not encoded in this charset.

# Features
- Format Gherkin features
- Format JSON in step doc string
- Scenario Outline. Recognize and compact JSON inside table 

 ## Supported JSON format in step doc string 

```json
{
  ...
  "key": <value>,
  ...
}
```

```json
{
  ...
  "key1": <value1>,
  <placeholder1>,
  <placeholder2>,
  ...
}
```

```json
[
  ...
  <value1>,
  <value2>,
  "value3",
  ...
]
```

## Compact JSON inside table

Examples like

```gherkin
Feature: The feature

  Scenario Outline: Compact json
    Given I load data:
    """
    <data>
    """
    Examples:
      | data                                       |
      | {"key1":   "value2",   "key2":   "value2"} |
      | [1,   2,   3]                              |
```

become

```gherkin
Feature: The feature

  Scenario Outline: Compact json
    Given I load data:
    """
    <data>
    """
    Examples:
      | data                              |
      | {"key1":"value2","key2":"value2"} |
      | [1,2,3]                           |
```

# Contribute<a id="contribute"></a>

If you want to add a new feature, open an issue with proposal

# Acknowledgements

Thanks [antham](https://github.com/antham) for the idea and some functions. Also, some functions got from [here](https://github.com/antham/ghokin) and official Go encoding [library](https://github.com/golang/go/tree/master/src/encoding/json)
