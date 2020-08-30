# prom-snapshot

Prometheus snapshot(s) viewer.

## Usage

### Host multiple prometheus data snapshot with v1 api

```
$ server -s snapshot1=/path/to/snapshot1 --snapshot snapshot2=/path/to/snapshot2
```

```
$ curl http://127.0.0.1:8080/snapshot1/api/v1/metadata
{"status":"success","data":{}}
$ curl http://127.0.0.1:8080/snapshot2/api/v1/metadata
{"status":"success","data":{}}
```

## TODO

- [ ] build
- [ ] docker image + kubernetes
- [ ] bundle grafana?
- [ ] (long term) download snapshot from remote

## LICENSE

MIT

---

a [@b4fun][@b4fun] project

[@b4fun]: https://www.build4.fun
