<aside class="notice">
This is alpha grade software and is not suitable for production use. Breaking changes are frequent as we march towards 1.0 and our <a href="">1.0 Compatability Promise</a>
</aside>

# Deviceio Hub

The Deviceio Hub forms the heart of the Deviceio Real-Time orchestration platform. The Hub listens for incoming Deviceio Agent connections and exposes the Agent APIs via HTTP endpoints available from the Hub. The Deviceio Hub provides a centralized location to conduct real-time system orchestration of multiple connected devices.

<!-- TOC -->

- [Deviceio Hub](#deviceio-hub)
- [Deployment](#deployment)
    - [Docker - Quick Start for Testing and Evaulation](#docker---quick-start-for-testing-and-evaulation)
    - [Docker - Swarm Cluster](#docker---swarm-cluster)

<!-- /TOC -->

# Deployment

## Docker - Quick Start for Testing and Evaulation

In this section we are going to deploy the hub for use with testing and evaluation. This method does not address scalability or security concerns.

Ensure Docker is installed (and in linux mode if using docker for windows 10)

From your cli, run the following docker command to start a basic rethinkdb database instance

```bash
docker run -d --name deviceio-test-db rethinkdb
```

Run the following docker command to start a basic Deviceio Hub instance

```bash
docker run -d --name deviceio-test-hub -p 4431:443 -p 8975:8975 --link deviceio-test-db:db deviceio/hub hub start --dbhost db
```

## Docker - Swarm Cluster

TBD