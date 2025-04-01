# Logging Package

This package implements a common logging package for Infrastructure as a Service applications, based
on [zerolog](https://github.com/rs/zerolog).

## Controlling the Log Level

This logging package defines a CLI `flag` named `-globalLogLevel`. As the name
suggests, it sets the global log level exposed by zerolog. Apps that want to
expose this flag, must call `flag.Parse()` in their `main` function. This is the
preferred approach to ensure a consistent UX. Should an app need to deviate, it
can call `zerolog.SetGlobalLevel(...)` as required.

## Output Formatting

By default, logging output is in machine-readable JSON format. For use cases
where a more human-readable format is desired, the `HUMAN` environment variable
should be set.

## Security Logging

Logging package exposes a tag called `McSec` which can be used to identify
security events happening across MC components

```go
// zlog is MCLogger, printing a security event
zlog.McSec().Info().Msgf("Client %s authorized", client.UUID)

// zlog is MCCtxLogger, printing a security event
zlog := zlog.TraceCtx(ctx)
zlog.McSec().Info().Msgf("Client %s authorized", client.UUID)
```

## Error Logging

Logging package exposes utilities to append `error` into the logs which can be easily
scraped by external tools

```go
// zlog is MCLogger, printing a security event and error
err := errors.Errorfc(codes.PermissionDenied, "Permission denied for client: %s", "1")
zlog.McSec().McErr(err).Msg("")

// zlog is MCCtxLogger, printing a security event and error
zlog := zlog.TraceCtx(ctx)
zlog.McSec().McErr(err).Msg("")

// zlog is MCLogger, printing a security event and error string
zlog.McSec().McError("Permission denied for client: %s", "1").Msg("CreateResource")

// zlog is MCCtxLogger, printing a security event and error string
zlog := zlog.TraceCtx(ctx)
zlog.McSec().McError("Permission denied for client: %s", "1").Msg("CreateResource")
```
