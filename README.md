ktee
====

Intercepts `stdout`/`stderr` and tees them to a Kafka topic.

### Usage

```bash
~> go get github.com/ian-kent/ktee
~> export KTEE_BROKERS="localhost:9092"
~> export KTEE_OUT_TOPIC="log"
~> export KTEE_ERR_TOPIC="log"
~> ktee echo "Grumpy wizards make toxic brew for the evil Queen and Jack"
```

Each line (`\n` separated) is treated as a new Kafka message. Data written to
`stdout` and `stderr` is buffered in-memory until a `\n` is found.

Both `stdout` and `stderr` are also written to the parent process file descriptors,
so piping and bash file descriptor redirection still work. This happens immediately
and is not affected by in-memory buffering for Kafka.

The current environment is passed to the child process without modification.

### Kafka errors

If a Kafka write fails, the parent and child process will be terminated.

Buffering of failed writes is not currently supported.

### Configuration

If `KTEE_BROKERS` is unset, no Kafka connection is attempted. `KTEE_OUT_TOPIC` and
`KTEE_ERR_TOPIC` are ignored.

If `KTEE_BROKERS` is set, a connection to Kafka is initiated. Kafka connection errors
will cause the process to exit and the command will not be executed.

If `KTEE_BROKERS` is set, but either `KTEE_OUT_TOPIC` or `KTEE_ERR_TOPIC` are unset,
the corresponding file descriptor is not written to Kafka.

### Licence

Copyright ©‎ 2015, Ian Kent (http://iankent.uk).

Released under MIT license, see [LICENSE](LICENSE.md) for details.
