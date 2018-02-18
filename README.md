# Bird Brain

Bird brain is a service that listens for messages and reconfigured a local instance of bird to match those messages.

## Supported actions
* AddPeer - Adds a new BGP peer
* DeletePeer - Delete a BGP peer
* AddRoute - Adds a static route
* DeleteRoute - Deletes a static route

The service is written in Go. The services uses gRPC, so you can write a client in whatever language you want that supports gRPC
