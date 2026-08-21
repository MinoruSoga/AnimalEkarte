package medicalrecord

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

// Field I/s from 城東 2026-08-19/20 idle captures. No patient values.
var (
	fieldIDEXXStatusInquiry = mustHex("023130808173f303")
	fieldIDEXXPortInquiry   = mustHex("023130808e4920312032303038323631393534bf03")
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestParseIDEXXPIMSFrame_FieldStatusInquiry(t *testing.T) {
	t.Parallel()
	got, ok := ParseIDEXXPIMSFrame(fieldIDEXXStatusInquiry)
	if !ok {
		t.Fatal("parse s")
	}
	if got.Kind != 's' || string(got.Payload) != "s" {
		t.Fatalf("kind/payload: %+v", got)
	}
	if got.Addr0 != '1' || got.Addr1 != '0' {
		t.Fatalf("addrs: %q %q", got.Addr0, got.Addr1)
	}
}

func TestParseIDEXXPIMSFrame_FieldPortInquiry(t *testing.T) {
	t.Parallel()
	got, ok := ParseIDEXXPIMSFrame(fieldIDEXXPortInquiry)
	if !ok {
		t.Fatal("parse I")
	}
	if got.Kind != 'I' {
		t.Fatalf("kind %q", got.Kind)
	}
	if string(got.Payload) != "I 1 2008261954" {
		t.Fatalf("payload %q", got.Payload)
	}
}

func TestParseIDEXXPIMSFrame_RejectsBadChecksum(t *testing.T) {
	t.Parallel()
	bad := append([]byte{}, fieldIDEXXStatusInquiry...)
	bad[len(bad)-2] ^= 0x01
	if _, ok := ParseIDEXXPIMSFrame(bad); ok {
		t.Fatal("expected reject")
	}
}

func TestIDEXXPIMSChecksum_MatchesFieldVectors(t *testing.T) {
	t.Parallel()
	cases := [][]byte{fieldIDEXXStatusInquiry, fieldIDEXXPortInquiry}
	for _, raw := range cases {
		got, ok := ParseIDEXXPIMSFrame(raw)
		if !ok {
			t.Fatalf("parse %x", raw)
		}
		if cs := IDEXXPIMSChecksum(got.Payload); cs != raw[len(raw)-2] {
			t.Fatalf("cs got %02x want %02x", cs, raw[len(raw)-2])
		}
	}
}

func TestBuildIDEXXPIMS_AFrameZerosAndChecksum(t *testing.T) {
	t.Parallel()
	frame := BuildIDEXXPIMSAFrame(1)
	got, ok := ParseIDEXXPIMSFrame(frame)
	if !ok {
		t.Fatalf("parse A %x", frame)
	}
	if got.Kind != 'A' {
		t.Fatalf("kind %q", got.Kind)
	}
	if got.Addr0 != '0' || got.Addr1 != 1 {
		t.Fatalf("addrs %q %d", got.Addr0, got.Addr1)
	}
	if len(got.Payload) != 16 || got.Payload[0] != 'A' || got.Payload[1] != 0x80 {
		t.Fatalf("payload %x", got.Payload)
	}
	for i := 2; i < 16; i++ {
		if got.Payload[i] != 0 {
			t.Fatalf("pad %d = %02x", i, got.Payload[i])
		}
	}
}

func TestBuildIDEXXPIMS_IMAndSMRoundTrip(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 8, 20, 19, 54, 0, 0, time.UTC)
	im := BuildIDEXXPIMSIMFrame(1, 1, clock)
	sm := BuildIDEXXPIMSSMFrame(1, 1)
	imGot, ok := ParseIDEXXPIMSFrame(im)
	if !ok {
		t.Fatalf("parse IM %x", im)
	}
	smGot, ok := ParseIDEXXPIMSFrame(sm)
	if !ok {
		t.Fatalf("parse SM %x", sm)
	}
	if imGot.Kind != 'I' || string(imGot.Payload) != "IM \x01 2008261954" {
		t.Fatalf("IM %+q", imGot.Payload)
	}
	if smGot.Kind != 'S' || string(smGot.Payload) != "SM \x01 s" {
		t.Fatalf("SM %+q", smGot.Payload)
	}
	// Port 0x01 SM payload sums to 0x154; CS must be 0xD4 (bit 7 forced), not 0x54.
	if sm[len(sm)-2] != 0xd4 {
		t.Fatalf("SM cs %02x want d4", sm[len(sm)-2])
	}
}

func TestBuildIDEXXPIMS_JoutoMdcon4SourcePortAreTwo(t *testing.T) {
	t.Parallel()
	// clinics/jouto mdcon4.cmd lines 18 and 19 are "2" / "2" (not shown on FrmIDEXX).
	im := BuildIDEXXPIMSIMFrame(2, 2, time.Date(2026, 8, 20, 19, 54, 0, 0, time.UTC))
	got, ok := ParseIDEXXPIMSFrame(im)
	if !ok {
		t.Fatalf("parse %x", im)
	}
	if got.Addr1 != 2 || got.Payload[3] != 2 {
		t.Fatalf("want CByte(2)=0x02, addr1=%d payload[3]=%d", got.Addr1, got.Payload[3])
	}
}

func TestDecodeIDEXXFrames_BadChecksumInquiryDoesNotAbortHematology(t *testing.T) {
	t.Parallel()
	bad := append([]byte{}, fieldIDEXXPortInquiry...)
	bad[len(bad)-2] ^= 0x01
	payload := append(append([]byte{}, bad...), synthIDEXXLongFrame()...)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatalf("bad I + long must still decode, got %v", err)
	}
	if len(frames) != 1 || frames[0].Items[0].Code != "WBC" {
		t.Fatalf("frames %+v", frames)
	}
}

func TestReplyIDEXXPIMSInquiry_Sequences(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 8, 20, 19, 54, 0, 0, time.UTC)
	iReplies, err := ReplyIDEXXPIMSInquiry(fieldIDEXXPortInquiry, IDEXXPIMSHost{Source: 1, Port: 1}, clock)
	if err != nil {
		t.Fatal(err)
	}
	if len(iReplies) != 3 || !bytes.Equal(iReplies[0], []byte{0x06}) {
		t.Fatalf("I replies %d first %x", len(iReplies), iReplies)
	}
	if g, ok := ParseIDEXXPIMSFrame(iReplies[1]); !ok || g.Kind != 'A' {
		t.Fatalf("I A %x", iReplies[1])
	}
	if g, ok := ParseIDEXXPIMSFrame(iReplies[2]); !ok || g.Kind != 'I' {
		t.Fatalf("I IM %x", iReplies[2])
	}

	sReplies, err := ReplyIDEXXPIMSInquiry(fieldIDEXXStatusInquiry, IDEXXPIMSHost{Source: 1, Port: 1}, clock)
	if err != nil {
		t.Fatal(err)
	}
	if len(sReplies) != 3 || sReplies[0][0] != 0x06 {
		t.Fatalf("s replies %d", len(sReplies))
	}
	if g, ok := ParseIDEXXPIMSFrame(sReplies[2]); !ok || g.Kind != 'S' {
		t.Fatalf("s SM %x", sReplies[2])
	}
}

func TestDecodeIDEXXFrames_PortInquiryMustNotAbortHematology(t *testing.T) {
	t.Parallel()
	payload := append(append([]byte{}, fieldIDEXXPortInquiry...), synthIDEXXLongFrame()...)
	frames, err := DecodeLabDeviceFrames(payload, "idexx_vetlab")
	if err != nil {
		t.Fatalf("I+long must decode, got %v", err)
	}
	if len(frames) != 1 || frames[0].Items[0].Code != "WBC" {
		t.Fatalf("frames %+v", frames)
	}
}

func TestDecodeIDEXXFrames_ControlOnlyStillInvalid(t *testing.T) {
	t.Parallel()
	_, err := DecodeLabDeviceFrames(fieldIDEXXPortInquiry, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) {
		t.Fatalf("I-only want invalid, got %v", err)
	}
	_, err = DecodeLabDeviceFrames(fieldIDEXXStatusInquiry, "idexx_vetlab")
	if !errors.Is(err, ErrInvalidLabDevicePayload) {
		t.Fatalf("s-only want invalid, got %v", err)
	}
}
