# The Open Dependencies project (OpenDeps) CLI

Features:

* Start live mocks of dependencies
* Test the availability of dependencies
* Validate OpenDeps files against [the specification](https://github.com/opendeps/specification)

## Getting started & documentation

### Installation

See the [Installation](./docs/install.md) instructions for your system.

#### Homebrew

If you have Homebrew installed:

    brew tap opendeps/cli
    brew install opendeps

#### Shell script

Or, use this one liner (macOS and Linux only):

```shell
curl -L https://raw.githubusercontent.com/opendeps/cli/main/install/install_opendeps.sh | bash -
```

### Usage

Top level command:

```
Usage:
  opendeps [command]

Available Commands:
  mock        Start live mocks of API dependencies
  test        Tests the availability of dependencies
  scaffold    Create an OpenDeps manifest from OpenAPI files
  validate    Validate a file against the OpenDeps schema
  help        Help about any command
```

#### Create and start mocks

Example:

    opendeps mock

Usage:

```
Starts a live mock of your API dependencies, based
on their OpenAPI specifications defined in the OpenDeps file.

This assumes that the specification URL is reachable
by this tool.

Usage:
  opendeps mock OPENDEPS_FILE

Flags:
  -p, --port   Port on which to listen (default 8080)
```

#### Test dependencies are available

Example:

    opendeps test

Usage:

```
Invokes the availability endpoints of each dependency,
optionally ignoring failures if the dependency is not
marked as required.

Usage:
  opendeps test OPENDEPS_FILE [flags]

Flags:
  -c, --continue                Continue to check further dependencies if one or more is down (default true)
  -h, --help                    help for test
  -z, --non-zero-exit           Exit with non-zero status if dependencies are down
  -o, --require-optional        Require optional dependencies to be available
  -s, --server stringToString   Override server base URL for a dependency (e.g. foo_service=https://example.com) (default [])
```

#### Create an OpenDeps manifest from OpenAPI files

Example:

    opendeps scaffold

Usage:

```
Creates an OpenDeps manifest based on the OpenAPI specification files in a directory.

If DIR is not specified, the current working directory is used.

Usage:
  opendeps scaffold DIR [flags]

Flags:
  -f, --force-overwrite   Force overwrite of destination file(s) if already exist
  -h, --help              help for scaffold
```

#### Validate OpenDeps file

Example:

    opendeps validate

Usage:

```
Validates a YAML manifest file against the OpenDeps schema.

Usage:
  opendeps validate OPENDEPS_FILE
```

#### Help

```
Provides help for any command in the application.
Simply type opendeps help [path to command] for full details.

Usage:
  opendeps help [command] [flags]
```

### Logging

The default log level is `debug`. You can override this by setting the `LOG_LEVEL` environment variable:

    export LOG_LEVEL=info

---

## About OpenDeps

OpenDeps allows you to express your application's _external runtime dependencies_. Using OpenDeps, you use a YAML file to clearly communicate what APIs your software component needs to run correctly.

Formalising your application's runtime dependencies helps improve software quality (e.g. earlier integration testing) and operations (e.g. deployment automation) to help you move faster, safely.

Benefits:
- Verify dependencies are in place before deploying your app
- Rapidly spin up live mocks of all your app's dependencies so you can build/test your app quickly
- Integrates with the [OpenAPI specification](https://github.com/OAI/OpenAPI-Specification)
- Language agnostic and cross-platform
- Human readable and machine readable; comprehensible by developers and non-developers

> Learn more about the [OpenDeps specification](https://github.com/opendeps/specification).

---

## Contributing

Suggestions and improvements to the CLI or documentation are welcome. Please raise pull requests targeting the `main` branch.
