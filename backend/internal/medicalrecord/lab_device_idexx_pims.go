package medicalrecord

import (
	"fmt"
	"time"
)

const (
	idexxPIMSSTX     = 0x02
	idexxPIMSETX     = 0x03
	idexxPIMSFlag    = 0x80
	idexxPIMSACK     = 0x06
	idexxPIMSNACK    = 0x15
	idexxPIMSHeaderN = 5
)

// IDEXXPIMSFrame is one STX…ETX session frame on the VetLab PIMS serial
// (I / s inquiries from the station, A / IM / SM replies from the host).
type IDEXXPIMSFrame struct {
	Addr0   byte
	Addr1   byte
	Payload []byte
	Kind    byte
}

// IDEXXPIMSHost is the Source/Port pair Tugi59 reads from the Drワン form.
// CByte of these integers is the wire byte (1 → 0x01, not ASCII '1').
type IDEXXPIMSHost struct {
	Source byte
	Port   byte
}

// IDEXXPIMSChecksum is (sum(payload) | 0x80) & 0xFF. Matches Tugi59 and
// field I/s captures from 城東 2026-08-19/20.
func IDEXXPIMSChecksum(payload []byte) byte {
	sum := 0
	for _, b := range payload {
		sum += int(b)
	}
	return byte(sum | 0x80)
}

// ParseIDEXXPIMSFrame returns false when the bytes are not a checksum-valid
// PIMS control frame. Measurement frames (hematology / SNAP) return false.
func ParseIDEXXPIMSFrame(raw []byte) (IDEXXPIMSFrame, bool) {
	n := len(raw)
	if n < 7 || raw[0] != idexxPIMSSTX || raw[n-1] != idexxPIMSETX {
		return IDEXXPIMSFrame{}, false
	}
	if raw[3] != idexxPIMSFlag {
		return IDEXXPIMSFrame{}, false
	}
	payloadLen := int(raw[4] & 0x7f)
	if n != idexxPIMSHeaderN+payloadLen+2 {
		return IDEXXPIMSFrame{}, false
	}
	payload := raw[idexxPIMSHeaderN : idexxPIMSHeaderN+payloadLen]
	if IDEXXPIMSChecksum(payload) != raw[n-2] {
		return IDEXXPIMSFrame{}, false
	}
	kind := byte(0)
	if payloadLen > 0 {
		kind = payload[0]
	}
	out := IDEXXPIMSFrame{
		Addr0:   raw[1],
		Addr1:   raw[2],
		Payload: append([]byte(nil), payload...),
		Kind:    kind,
	}
	return out, true
}

func looksLikeIDEXXPIMSControlFrame(raw []byte) bool {
	n := len(raw)
	if n < 7 || raw[0] != idexxPIMSSTX || raw[n-1] != idexxPIMSETX || raw[3] != idexxPIMSFlag {
		return false
	}
	payloadLen := int(raw[4] & 0x7f)
	if n != idexxPIMSHeaderN+payloadLen+2 || payloadLen == 0 {
		return false
	}
	switch raw[idexxPIMSHeaderN] {
	case 'I', 's', 'A', 'S':
		return true
	default:
		return false
	}
}

func encodeIDEXXPIMS(addr0, addr1 byte, payload []byte) []byte {
	if len(payload) > 0x7f {
		payload = payload[:0x7f]
	}
	out := make([]byte, 0, idexxPIMSHeaderN+len(payload)+2)
	out = append(out, idexxPIMSSTX, addr0, addr1, idexxPIMSFlag, idexxPIMSFlag|byte(len(payload)))
	out = append(out, payload...)
	out = append(out, IDEXXPIMSChecksum(payload), idexxPIMSETX)
	return out
}

// BuildIDEXXPIMSAFrame is Tugi59's type-A reply (16-byte payload, 14 trailing zeros).
func BuildIDEXXPIMSAFrame(source byte) []byte {
	payload := make([]byte, 16)
	payload[0] = 'A'
	payload[1] = idexxPIMSFlag
	return encodeIDEXXPIMS('0', source, payload)
}

// BuildIDEXXPIMSIMFrame uses 24-hour hhmm to match field I timestamps.
// Tugi59 Format(Now,"hhmm") may be 12-hour; TX capture must confirm.
func BuildIDEXXPIMSIMFrame(source, port byte, clock time.Time) []byte {
	stamp := clock.Format("0201061504")
	payload := []byte{'I', 'M', ' ', port, ' '}
	payload = append(payload, stamp...)
	return encodeIDEXXPIMS('0', source, payload)
}

// BuildIDEXXPIMSSMFrame is Tugi59's status response.
func BuildIDEXXPIMSSMFrame(source, port byte) []byte {
	return encodeIDEXXPIMS('0', source, []byte{'S', 'M', ' ', port, ' ', 's'})
}

// ReplyIDEXXPIMSInquiry returns ACK + A + IM (for I) or ACK + A + SM (for s).
// It does not open a serial port. Do not send the result to a live VetLab
// until Source/Port are confirmed from Drワン TX.
func ReplyIDEXXPIMSInquiry(inquiry []byte, host IDEXXPIMSHost, clock time.Time) ([][]byte, error) {
	got, ok := ParseIDEXXPIMSFrame(inquiry)
	if !ok {
		return nil, fmt.Errorf("idexx pims: not a control inquiry")
	}
	ack := []byte{idexxPIMSACK}
	a := BuildIDEXXPIMSAFrame(host.Source)
	switch got.Kind {
	case 'I':
		return [][]byte{ack, a, BuildIDEXXPIMSIMFrame(host.Source, host.Port, clock)}, nil
	case 's':
		return [][]byte{ack, a, BuildIDEXXPIMSSMFrame(host.Source, host.Port)}, nil
	default:
		return nil, fmt.Errorf("idexx pims: kind %q is not I or s", got.Kind)
	}
}
