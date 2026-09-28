package services

import (
	"testing"

	"codecollab/models"
	"github.com/pion/webrtc/v4"
)

func TestReconnectedSourceReplacesTrack(t *testing.T) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	s := &SFUService{}
	source := &models.VoicePeer{UserID: "source"}
	target := &models.VoicePeer{UserID: "listener", PeerConnection: pc}
	newTrack := func() *webrtc.TrackLocalStaticRTP {
		track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}, "audio-source", "stream-source")
		if err != nil {
			t.Fatal(err)
		}
		return track
	}
	first := newTrack()
	sender, added := s.addTrackToPeer(source, target, first, nil, false)
	if !added {
		t.Fatal("initial track not added")
	}
	s.addTrackToPeer(source, target, first, nil, false)
	second := newTrack()
	s.addTrackToPeer(source, target, second, nil, false)
	if sender.Track() != second {
		t.Fatal("listener still bound to old track after reconnect")
	}
	if len(pc.GetSenders()) != 1 {
		t.Fatal("duplicate sender created")
	}
	if _, added := s.addTrackToPeer(target, target, second, nil, false); added {
		t.Fatal("self track was added")
	}
}
