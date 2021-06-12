# The Open Dependencies project (OpenDeps) CLI

Top level command:

```
Usage:
  opendeps [command]

Available Commands:
  help        Help about any command
  mock        Start live mocks of API dependencies
  validate    Validate a file against the OpenDeps schema
```

Validate OpenDeps file:

```
Validates a YAML file against the OpenDeps schema.

Usage:
  opendeps validate FILE
```

Create and start mocks:

```
Starts a live mock of your API dependencies, based
on their OpenAPI specifications defined in the OpenDeps file.

This assumes that the specification URL is reachable
by this tool.

Usage:
  opendeps mock FILE
```

Help:

```
Provides help for any command in the application.
Simply type opendeps help [path to command] for full details.

Usage:
  opendeps help [command] [flags]
```

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
