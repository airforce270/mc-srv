# mc-srv

A simple Minecraft server to play around with the protocol.

## Implementation status

### Server list ping

- [x] Handle handshake packet
- [x] Handle status request packet
- [x] Send status response packet
- [x] Handle ping request packet
- [x] Send ping response packet

### Login

- [x] Handle handshake packet (state=2)
- [x] Handle login start packet
- [x] Send encryption request packet
- [x] Handle encryption response packet
- [x] Send login success packet
- [x] Handle login acknowledged packet

### Configuration

- [x] Send plugin message configuration packets (not needed)
- [ ] Send disconnect packets when needed
- [ ] Send finish configuration packet
- [x] Send keep alive packets
- [x] Send ping packets (not needed)
- [ ] Send registry data packet
- [ ] Send remove resource pack packet
- [ ] Send add resource pack packet
- [ ] Send feature flags packet
- [ ] Send update tags packet
- [x] Handle client information packet
- [ ] Store data from client information packet(?)
- [x] Handle serverbound plugin message packet
- [ ] Handle acknowledge finish configuration packet
- [x] Handle serverbound keep alive packets
- [x] Disconnect clients if they don't respond to keepalive pings in a reasonable time
- [x] Handle pong packets (not needed)
- [x] Handle resource pack response packets
- [ ] Store data from resource pack response packets(?)

### Play

- [ ] A lot :)
