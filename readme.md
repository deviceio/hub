# Deviceio Hub

The Deviceio Hub provides a centralized location to conduct real-time system orchestration of multiple connected devices. The Hub listens for incoming Deviceio Agent connections and exposes Agent APIs via HTTP endpoints available from the Hub. 

# Deployment

## Docker - Quick Start

Start a rethinkdb instance

```bash
docker run -d --name deviceio-test-db rethinkdb
```

Wait for the rethinkdb instance to startup by querying the container logs until you see `Server ready`

```bash
docker logs deviceio-test-db
```

Start a Deviceio Hub instance and link to our rethinkdb instance

```bash
docker run -d --name deviceio-test-hub -p 4431:4431 -p 8975:8975 --link deviceio-test-db:db deviceio/hub --db-host db
```

Test connectivity to the hub api port

```bash
curl -k -v https://127.0.0.1:4431/v1/status
```

Test connectivity to the hub gateway port

```bash
curl -k -v https://127.0.0.1:8975/v1/status
```

Next:

* Install and join a Deviceio Agent to your hub instance https://github.com/deviceio/agent
* Install and interact with the hub with the CLI tools https://github.com/deviceio/cli