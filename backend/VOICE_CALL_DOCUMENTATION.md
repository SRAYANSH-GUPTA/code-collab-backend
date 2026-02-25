# Voice Call System Documentation

## Overview

This document describes the **Discord-like voice call system** implemented using **SFU (Selective Forwarding Unit)** architecture with Pion WebRTC v4.

## Architecture

### Why SFU?

**SFU (Selective Forwarding Unit)** was chosen over Mesh and MCU for these reasons:

| Architecture | Pros | Cons | Best For |
|-------------|------|------|----------|
| **Mesh (P2P)** | Simple, no server load | O(n²) bandwidth, breaks >4 users | 2-4 users max |
| **SFU** ✅ | O(n) bandwidth, scalable, low CPU | Requires media server | 5-100+ users |
| **MCU** | Best quality, O(1) bandwidth | Expensive CPU, high latency | Enterprise conferencing |

**Our Choice: SFU** - Same as Discord, Zoom, Google Meet
- Scales to 50+ users per room
- Low server CPU (just forwards packets)
- Client-side quality control
- Cost-effective

### System Components

```
┌─────────────┐                    ┌──────────────────┐                    ┌─────────────┐
│  Client A   │◄──────────────────►│  Signaling       │◄──────────────────►│  Client B   │
│  (Browser)  │   WebSocket        │  Server (Go)     │   WebSocket        │  (Browser)  │
└─────────────┘                    └──────────────────┘                    └─────────────┘
       │                                    │                                      │
       │                                    │                                      │
       │         WebRTC Media (P2P)        │         WebRTC Media (P2P)          │
       │    ┌──────────────────────────────┼──────────────────────────────┐      │
       └────┤         SFU Server            │                              ├──────┘
            │  (Forwards RTP Packets)       │                              │
            └───────────────────────────────┴──────────────────────────────┘
```

## Implementation Details

### 1. Tech Stack

**Backend:**
- **Pion WebRTC v4**: WebRTC implementation in Go
- **Gorilla WebSocket**: Signaling channel
- **Custom SFU**: Built on Pion for full control

**Frontend (to be implemented):**
- Native WebRTC API or `simple-peer`
- React/Next.js for UI

**Infrastructure:**
- **STUN**: Google's free STUN servers (default)
- **TURN**: Optional (for restrictive NATs)

### 2. File Structure

```
backend/
├── models/
│   └── voice.go              # Voice data models (VoiceRoom, VoicePeer, etc.)
├── services/
│   ├── room_manager.go       # Thread-safe room management
│   └── sfu.go                # SFU core logic (WebRTC handling)
├── handlers/
│   ├── voice.go              # Voice WebSocket message handlers
│   └── websocket.go          # Main WebSocket handler (updated)
├── metrics/
│   └── voice_metrics.go      # Prometheus metrics for voice
├── config/
│   └── config.go             # Configuration (STUN/TURN servers)
└── main.go                   # Initialization
```

### 3. Key Data Models

#### VoiceRoom
```go
type VoiceRoom struct {
    ID              string
    Name            string
    Peers           map[string]*VoicePeer
    MaxParticipants int (50 by default)
    // Thread-safe with sync.RWMutex
}
```

#### VoicePeer
```go
type VoicePeer struct {
    UserID         string
    PeerConnection *webrtc.PeerConnection
    WSConnection   *websocket.Conn
    IsMuted        bool
    LocalTracks    []*webrtc.TrackLocalStaticRTP
    RemoteTracks   []*webrtc.TrackRemote
    // Quality metrics: PacketsLost, Jitter, RTT
}
```

### 4. WebSocket Message Protocol

All voice messages are sent over the same WebSocket connection as code analysis.

#### Join Room
```json
{
  "action": "join",
  "roomId": "room-123"
}
```

**Response:**
```json
{
  "type": "joined",
  "roomId": "room-123",
  "userId": "user-456",
  "users": [
    {"userId": "user-123", "isMuted": false, "isSpeaking": false},
    {"userId": "user-456", "isMuted": false, "isSpeaking": false}
  ]
}
```

#### WebRTC Offer (after join)
```json
{
  "action": "offer",
  "roomId": "room-123",
  "sdp": {
    "type": "offer",
    "sdp": "v=0\r\no=- 123456..."
  }
}
```

**Response:**
```json
{
  "type": "answer",
  "roomId": "room-123",
  "userId": "user-456",
  "sdp": {
    "type": "answer",
    "sdp": "v=0\r\no=- 789012..."
  }
}
```

#### ICE Candidate (trickle ICE)
```json
{
  "action": "ice_candidate",
  "roomId": "room-123",
  "candidate": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}
```

#### Mute/Unmute
```json
{
  "action": "mute",
  "roomId": "room-123"
}
```

#### Leave Room
```json
{
  "action": "leave",
  "roomId": "room-123"
}
```

### 5. Connection Flow

```
1. User clicks "Join Voice Channel"
   ↓
2. Frontend sends WebSocket: {"action": "join", "roomId": "..."}
   ↓
3. Backend creates PeerConnection, adds user to room
   ↓
4. Frontend receives: {"type": "joined", "users": [...]}
   ↓
5. Frontend creates offer → sends WebSocket: {"action": "offer", "sdp": {...}}
   ↓
6. Backend processes offer, creates answer
   ↓
7. Frontend receives: {"type": "answer", "sdp": {...}}
   ↓
8. Both sides exchange ICE candidates
   ↓
9. WebRTC connection established (P2P media flow through SFU)
   ↓
10. User starts sending audio → SFU forwards to all other peers
```

### 6. SFU Forwarding Logic

When a user sends audio:

```go
1. Client A sends RTP packets → Server
2. Server receives on track: OnTrack(track)
3. Server creates LocalTrack
4. Server forwards packets to all other peers
5. Clients B, C, D receive audio from A
```

**Key Code:**
```go
pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    // Create local track to forward
    localTrack, _ := webrtc.NewTrackLocalStaticRTP(...)

    // Read from remote track, write to local track
    go func() {
        for {
            n, _ := track.Read(rtpBuf)
            localTrack.Write(rtpBuf[:n])  // Forwards to all peers
        }
    }()

    // Add track to all other peer connections
    for _, otherPeer := range room.GetAllPeers() {
        otherPeer.PeerConnection.AddTrack(localTrack)
    }
})
```

## Configuration

### Environment Variables

```bash
# Optional: Custom STUN servers (comma-separated)
STUN_SERVERS=stun:stun1.example.com:19302,stun:stun2.example.com:19302

# Optional: TURN servers for users behind restrictive NATs
TURN_SERVERS=turn:turn.example.com:3478
TURN_USERNAME=youruser
TURN_PASSWORD=yourpass
```

### Default Configuration

If not specified, uses Google's free STUN servers:
- `stun:stun.l.google.com:19302`
- `stun:stun1.l.google.com:19302`
- `stun:stun2.l.google.com:19302`

## Best Practices Implemented

### 1. Thread Safety
- All room operations use `sync.RWMutex`
- Concurrent access to peers is protected
- No data races

### 2. Resource Cleanup
- Automatic cleanup on disconnect
- Empty room removal (prevents memory leaks)
- Periodic cleanup job (every 5 minutes)
- Proper PeerConnection closure

### 3. Scalability
- SFU architecture (O(n) bandwidth)
- Room-based isolation
- Max 50 participants per room (configurable)

### 4. Monitoring
- Prometheus metrics for:
  - Active rooms count
  - Total peers
  - Peers per room distribution
  - Connection quality (packet loss, jitter, RTT)
  - Operation success/failure rates
  - ICE state changes

### 5. Error Handling
- Graceful degradation
- Proper error responses
- Connection state monitoring
- Auto-reconnection support

### 6. WebRTC Best Practices
- **Trickle ICE**: Candidates sent immediately
- **Unified Plan**: Modern SDP semantics
- **Multiple STUN servers**: Redundancy
- **TURN fallback**: For restrictive networks

## Monitoring & Metrics

### Prometheus Metrics

Access at: `http://localhost:8080/metrics`

**Voice Metrics:**
- `voice_rooms_active` - Number of active voice rooms
- `voice_peers_total` - Total peers in calls
- `voice_peers_per_room` - Distribution histogram
- `voice_operations_total` - Operations (join, leave, mute)
- `voice_ice_state_changes_total` - ICE state transitions
- `voice_errors_total` - Error counts by type
- `voice_packet_loss_percent` - Packet loss distribution
- `voice_jitter_milliseconds` - Jitter distribution
- `voice_rtt_milliseconds` - RTT distribution

### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-14T12:00:00Z",
  "active_connections": 5,
  "voice": {
    "initialized": true,
    "roomCount": 2,
    "totalPeers": 8
  }
}
```

## Testing

### 1. Start the Server

```bash
cd backend
go run main.go
```

### 2. Test WebSocket Connection

```bash
# Install wscat
npm install -g wscat

# Connect
wscat -c "ws://localhost:8080/ws?token=test-token"

# Join room
> {"action":"join","roomId":"test-room"}

# You should receive
< {"type":"joined","roomId":"test-room","userId":"test-user-...","users":[]}
```

### 3. Frontend Integration Example

```javascript
// Connect WebSocket
const ws = new WebSocket('ws://localhost:8080/ws?token=your-token');

// Create peer connection
const pc = new RTCPeerConnection({
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' }
  ]
});

// Get user media
const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
stream.getTracks().forEach(track => pc.addTrack(track, stream));

// Create and send offer
const offer = await pc.createOffer();
await pc.setLocalDescription(offer);
ws.send(JSON.stringify({
  action: 'offer',
  roomId: 'room-123',
  sdp: offer
}));

// Handle answer
ws.onmessage = async (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'answer') {
    await pc.setRemoteDescription(msg.sdp);
  }
  if (msg.type === 'ice_candidate') {
    await pc.addIceCandidate(msg.candidate);
  }
};

// Send ICE candidates
pc.onicecandidate = (event) => {
  if (event.candidate) {
    ws.send(JSON.stringify({
      action: 'ice_candidate',
      roomId: 'room-123',
      candidate: event.candidate
    }));
  }
};

// Receive audio
pc.ontrack = (event) => {
  const audio = new Audio();
  audio.srcObject = event.streams[0];
  audio.play();
};
```

## Scaling Considerations

### Current Limits
- 50 users per room (configurable)
- Single server deployment

### To Scale Further:

1. **Horizontal Scaling**
   - Add load balancer
   - Use Redis for room state
   - Sticky sessions for WebSocket

2. **Database Integration**
   - Persist room metadata
   - User presence tracking
   - Call history/analytics

3. **Advanced Features**
   - Simulcast (multiple quality streams)
   - SVC (Scalable Video Coding)
   - Bandwidth estimation
   - Auto quality adjustment

4. **Production TURN**
   - Deploy coturn server
   - Or use Twilio/Agora TURN

## Security Considerations

### Current Implementation
- ✅ Token-based authentication
- ✅ Room isolation
- ✅ Max participant limits
- ✅ Rate limiting (existing)

### Production Recommendations
- [ ] Encrypt TURN credentials
- [ ] Room access control (permissions)
- [ ] DTLS for WebRTC (auto-enabled by Pion)
- [ ] SRTP for media encryption (auto-enabled)
- [ ] Ban/kick functionality
- [ ] Audit logging

## Troubleshooting

### Connection Issues

**Symptom**: ICE connection fails

**Solutions:**
1. Check firewall allows UDP traffic
2. Add TURN server for NAT traversal
3. Check browser console for WebRTC errors

**Symptom**: No audio

**Solutions:**
1. Check microphone permissions
2. Verify `getUserMedia()` succeeds
3. Check `pc.ontrack` receives streams
4. Verify audio element is not muted

**Symptom**: High latency

**Solutions:**
1. Check RTT metrics in Prometheus
2. Reduce number of participants
3. Use TURN server closer to users
4. Enable Opus codec DTX (discontinuous transmission)

## Next Steps

### Frontend Implementation
1. Create React voice channel UI
2. Implement WebRTC client logic
3. Add mute/unmute buttons
4. Voice activity detection (speaking indicator)
5. Volume controls

### Backend Enhancements
1. Recording support
2. Push-to-talk mode
3. Voice filters/effects
4. Screen sharing support
5. Video support

### DevOps
1. Docker containerization
2. Kubernetes deployment
3. CI/CD pipeline
4. Load testing

## References

- [Pion WebRTC Documentation](https://github.com/pion/webrtc)
- [WebRTC Specification](https://webrtc.org/)
- [SFU Architecture Guide](https://webrtcglossary.com/sfu/)
- [Discord Engineering Blog](https://discord.com/category/engineering)

## Support

For issues or questions:
- Check logs: Backend shows detailed WebRTC events
- Check metrics: `/metrics` endpoint
- Review browser console: WebRTC errors logged
- Test with `wscat` to isolate frontend vs backend issues
