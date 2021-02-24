<div align="center">
  <h1>Zerolog Mia</h1>

  [![Test and Build Go Project][github-actions-svg]][github-actions]

</div>

Provides a custom log format for [zerolog](https://github.com/rs/zerolog) logger. This log format is inspired by [Mia-Platform logging guidelines][logging-guidelines].

The logger output is provided in JSON format, and it can be parsed by prettifiers to simplify its inspection. Considering the inspiration of this logger from Pino node js logger, it is possible to exploit [pino-pretty][pino-pretty] library to achieve this goal.

[github-actions]: https://github.com/danibix95/zerolog-mia/actions/workflows/go.yml
[github-actions-svg]: https://github.com/danibix95/zerolog-mia/actions/workflows/go.yml/badge.svg?branch=main
[logging-guidelines]: https://docs.mia-platform.eu/docs/getting_started/monitoring-dashboard/dev_ops_guide/log#json-logging-format
[pino-pretty]: https://github.com/pinojs/pino-pretty
