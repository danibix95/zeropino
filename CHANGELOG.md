# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.1] 2021-04-11

### Changes

- fixed logging - path property now displays the whole request URI, not just the path component
- added CODE_OF_CONDUCT.md
- added CONTRIBUTING.md
- added issues templates

## [v0.2.0] 2021-04-11

### Changes

- fixed fiber middleware
- improved tests
- updated README
- added benchmarks

## [v0.1.0] 2021-04-10

### Initial release provides
- README
- changelog
- pre-commit configuration
- custom default logger inspired by Mia Platform logging guidelines
- tests to verify custom logger behaviour
- LICENSE
- gorilla mux middleware adapted from [glogger](https://github.com/mia-platform/glogger)
- fiber middleware
