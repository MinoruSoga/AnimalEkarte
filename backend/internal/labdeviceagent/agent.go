package labdeviceagent

import (
	"context"
	"io"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	portPattern    = "/dev/cu.usbserial-*"
	frameIdle      = 2 * time.Second
	portScanPeriod = 2 * time.Second
	maxFrameBytes  = 8 * 1024
)

type GlobFunc func(pattern string) ([]string, error)
type OpenFunc func(ctx context.Context, path string) (io.ReadCloser, error)
// PIMSReplyFunc consumes a carry+chunk buffer and returns session replies plus
// the number of prefix bytes that were parsed (complete frames). The tail stays
// in the caller for the next USB read.
type PIMSReplyFunc func(buf []byte) (replies [][]byte, consumed int)

type FrameBuffer struct {
	chunks   [][]byte
	size     int
	overflow bool
}

func (b *FrameBuffer) Push(chunk []byte) bool {
	if len(chunk) == 0 {
		return true
	}
	if b.overflow || b.size+len(chunk) > maxFrameBytes {
		b.chunks = nil
		b.size = 0
		b.overflow = true
		return false
	}
	b.chunks = append(b.chunks, append([]byte(nil), chunk...))
	b.size += len(chunk)
	return true
}

func (b *FrameBuffer) Take() []byte {
	if b.overflow {
		b.overflow = false
		return nil
	}
	if len(b.chunks) == 0 {
		return nil
	}
	total := 0
	for _, chunk := range b.chunks {
		total += len(chunk)
	}
	frame := make([]byte, 0, total)
	for _, chunk := range b.chunks {
		frame = append(frame, chunk...)
	}
	b.chunks = nil
	b.size = 0
	return frame
}

func DiscoverPorts(glob GlobFunc) ([]string, error) {
	ports, err := glob(portPattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(ports)
	return ports, nil
}

type Agent struct {
	queue        *Queue
	status       *Status
	glob         GlobFunc
	open         OpenFunc
	allowedPorts map[string]struct{}
	pimsReply    PIMSReplyFunc
}

func NewAgent(queue *Queue, status *Status, allowedPorts []string) *Agent {
	allowed := make(map[string]struct{}, len(allowedPorts))
	for _, path := range allowedPorts {
		allowed[path] = struct{}{}
	}
	status.SetConfiguredPorts(len(allowed))
	return &Agent{
		queue:        queue,
		status:       status,
		glob:         filepath.Glob,
		open:         openSerial,
		allowedPorts: allowed,
	}
}

// EnablePIMSReply writes ACK+A+IM/SM on the same usbserial when the callback
// returns session replies. Default is off. Do not enable on a hospital VetLab cable.
func (a *Agent) EnablePIMSReply(fn PIMSReplyFunc) {
	a.pimsReply = fn
}

// UseReadWriteSerial opens allowlisted cu.usbserial ports O_RDWR so PIMS replies
// can be written. Call only together with EnablePIMSReply.
func (a *Agent) UseReadWriteSerial() {
	a.open = openSerialRDWR
}

func (a *Agent) Run(ctx context.Context) {
	ticker := time.NewTicker(portScanPeriod)
	defer ticker.Stop()
	active := make(map[string]context.CancelFunc)
	ended := make(chan string, 16)
	var monitors sync.WaitGroup
	scan := func() {
		ports, err := DiscoverPorts(a.glob)
		if err != nil {
			a.status.AddDiscoveryError()
			return
		}
		for _, path := range ports {
			if _, allowed := a.allowedPorts[path]; !allowed {
				continue
			}
			if _, exists := active[path]; exists {
				continue
			}
			portCtx, cancel := context.WithCancel(ctx)
			active[path] = cancel
			monitors.Add(1)
			go func() {
				defer monitors.Done()
				a.monitorPort(portCtx, path)
				select {
				case ended <- path:
				case <-ctx.Done():
				}
			}()
		}
	}
	scan()
	for {
		select {
		case <-ctx.Done():
			for _, cancel := range active {
				cancel()
			}
			monitors.Wait()
			return
		case path := <-ended:
			if cancel, exists := active[path]; exists {
				cancel()
				delete(active, path)
			}
		case <-ticker.C:
			scan()
		}
	}
}

func (a *Agent) monitorPort(ctx context.Context, path string) {
	portCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, err := a.open(portCtx, path)
	if err != nil {
		if ctx.Err() == nil {
			a.status.AddOpenError()
		}
		return
	}
	var closeOnce sync.Once
	closeReader := func() {
		closeOnce.Do(func() {
			if closeErr := reader.Close(); closeErr != nil {
				a.status.AddCloseError()
			}
		})
	}
	a.status.AddOpenPorts(1)
	defer a.status.AddOpenPorts(-1)
	stopCloser := make(chan struct{})
	closerDone := make(chan struct{})
	go func() {
		defer close(closerDone)
		select {
		case <-portCtx.Done():
			closeReader()
		case <-stopCloser:
		}
	}()
	reads, readDone := pumpPortReads(portCtx, reader)
	var writer io.Writer
	if a.pimsReply != nil {
		if w, ok := reader.(io.Writer); ok {
			writer = w
		}
	}
	a.collectPortFrames(portCtx, cancel, reads, writer)
	close(stopCloser)
	closeReader()
	<-closerDone
	<-readDone
}

type portReadResult struct {
	bytes []byte
	err   error
}

func pumpPortReads(ctx context.Context, reader io.Reader) (<-chan portReadResult, <-chan struct{}) {
	reads := make(chan portReadResult)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 4096)
		for {
			count, readErr := reader.Read(buffer)
			result := portReadResult{bytes: append([]byte(nil), buffer[:count]...), err: readErr}
			select {
			case reads <- result:
			case <-ctx.Done():
				return
			}
			if readErr != nil {
				return
			}
		}
	}()
	return reads, done
}

func (a *Agent) collectPortFrames(ctx context.Context, cancel context.CancelFunc, reads <-chan portReadResult, writer io.Writer) {
	frames := &FrameBuffer{}
	var pimsBuf []byte
	var idle *time.Timer
	var idleChannel <-chan time.Time
	stopIdle := func() {
		if idle != nil && !idle.Stop() {
			select {
			case <-idle.C:
			default:
			}
		}
	}
	defer stopIdle()
	enqueueBuffered := func() {
		frame := frames.Take()
		if len(frame) > 0 {
			if _, enqueueErr := a.queue.Enqueue(frame, time.Now()); enqueueErr != nil {
				a.status.AddQueueError()
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			enqueueBuffered()
			return
		case result := <-reads:
			if len(result.bytes) > 0 {
				wasOverflow := frames.overflow
				if !frames.Push(result.bytes) {
					if !wasOverflow {
						a.status.AddInputOverflow()
					}
				}
				if a.pimsReply != nil && writer != nil {
					if len(pimsBuf)+len(result.bytes) > maxFrameBytes {
						pimsBuf = nil
						a.status.AddInputOverflow()
					} else {
						pimsBuf = append(pimsBuf, result.bytes...)
					}
					replies, consumed := a.pimsReply(pimsBuf)
					if consumed > 0 {
						pimsBuf = append([]byte(nil), pimsBuf[consumed:]...)
					}
					for _, reply := range replies {
						if len(reply) == 0 {
							continue
						}
						if _, writeErr := writer.Write(reply); writeErr != nil {
							a.status.AddOpenError()
							cancel()
							enqueueBuffered()
							return
						}
					}
				}
				stopIdle()
				idle = time.NewTimer(frameIdle)
				idleChannel = idle.C
			}
			if result.err != nil {
				enqueueBuffered()
				return
			}
		case <-idleChannel:
			idleChannel = nil
			enqueueBuffered()
		}
	}
}
