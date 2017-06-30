# Deviceio Hub

The Deviceio Hub provides real-time access to connected devices.

# Quick Start with Docker

Start a rethinkdb instance

```bash
docker run -d --name deviceio-db rethinkdb
```

Wait for the rethinkdb instance to startup by following the container logs until you see `Server ready`

```bash
docker logs -f deviceio-db
```

Initialize the Deviceio Hub database and initial credential

```bash
docker run -ti --rm --link deviceio-db:db deviceio/hub init --db-host db
```

Start a Deviceio Hub instance and link to our rethinkdb instance

```bash
docker run -d --name deviceio-hub -p 4431:4431 -p 8975:8975 --link deviceio-db:db deviceio/hub start --db-host db
```

Test connectivity to the hub api port

```bash
curl -k -v https://127.0.0.1:4431/v1/status
```

Test connectivity to the hub gateway port

```bash
telnet 127.0.0.1 8975
```

Next:

* Install and join a device to your hub instance https://github.com/deviceio/agent
* Install and interact with your devices via the CLI integration https://github.com/deviceio/cli