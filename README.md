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
For endpoints documentation see [protobuf definition file](/faceit/usersvc/v1/proto.proto).

## Assumptions

### Database

1. Hashed passwords are in the separate collection from user's data (this is the reason for having transactions, see next point).
2. We need to have transaction, so standalone mongoDB deployment doesn't work.

### Code organization

1. Storing generated code in source control, perhaps can be revised too.
2. Keeping things simple for start, starting with two-tier architecture (not having "services").

### Other details

1. For configuration I've used [package](https://github.com/gopher-lib/config) which I developed and open-sourced a few months ago.
2. Reusage of protobuf message in the Spirit of [Google Cloud API Design Guide](https://cloud.google.com/apis/design).
3. Having [protobuf definition file](/faceit/usersvc/v1/proto.proto) as a source of documentation.

## Possible extensions and improvements

0. add wait for script to docker-compose (or migrate to more serious deployment)
1. **Improve error handling**: instead of sending error details to API consumer log them and send only error description.
2. **Revise validation**: add validation for passwords and send better error message (perhaps with [errdetails package](https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/errdetails)).
3. **Extract code from controllers into services**: for cleaner architecture.
4. **Add sorting and searching functionality for Service.ListUsers RPC**.
5. **Setup error reporting**: can be with [Cloud Error Reporting](https://cloud.google.com/error-reporting).
