# User Microservice

## Running on _localhost_ With _docker-compose_

### Prerequisites:

1. `docker-compose` installed.
2. Open `8080` and `8081` ports.

#### Steps:

1. Run `docker-compose up`, grpc-server is accessible at `localhost:8080` and mongoDB dashboard at `localhost:8081`.  
   Wait for `Listening at [::]:8080` log from `server` container  
   Replica can try to setup even a few minutes, alternatively consider using, MongoDB Altas free tier cluster.  
   In order to run with remote cluster provide connection URI as `MONGODB_URI` env. variable.

## Testing

Endpoints can be tested with [evans-cli](https://github.com/ktr0731/evans) or [bloomrpc](https://github.com/uw-labs/bloomrpc).  
For endpoints documentation see [protobuf definition file](/usersvc/v1/proto.proto).
