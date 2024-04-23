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

- [ ] Send configuration packet
- [ ] Send disconnect packets if needed
- [ ] Send finish configuration packet
- [ ] Send keep alive packets
- [ ] Send ping packets
- [ ] Send registry data packet
- [ ] Send remove resource pack packet
- [ ] Send add resource pack packet
- [ ] Send feature flags packet
- [ ] Send update tags packet
- [ ] Handle client information packet
- [ ] Handle serverbound plugin message packet
- [ ] Handle acknowledge finish configuration packet
- [ ] Handle serverbound keep alive packets
- [ ] Handle pong packets
- [ ] Handle resource pack response packets

### Play

- [ ] A lot :)
