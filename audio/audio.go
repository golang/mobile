// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux,!android

// Package audio provides a basic audio player.
package audio

import (
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/mobile/audio/al"
	"golang.org/x/mobile/audio/alc"
)

// Format represents an PCM data format.
type Format int

const (
	Mono8 Format = iota
	Mono16
	Stereo8
	Stereo16
)

func (f Format) String() string { return formatStrings[f] }

// formatBytes is the product of bytes per sample and number of channels.
var formatBytes = [...]int64{
	Mono8:    1,
	Mono16:   2,
	Stereo8:  2,
	Stereo16: 4,
}

var formatCodes = [...]uint32{
	Mono8:    al.FormatMono8,
	Mono16:   al.FormatMono16,
	Stereo8:  al.FormatStereo8,
	Stereo16: al.FormatStereo16,
}

var formatStrings = [...]string{
	Mono8:    "mono8",
	Mono16:   "mono16",
	Stereo8:  "stereo8",
	Stereo16: "stereo16",
}

// State indicates the current playing state of the player.
type State int

const (
	Unknown State = iota
	Initial
	Playing
	Paused
	Stopped
)

func (s State) String() string { return stateStrings[s] }

var stateStrings = [...]string{
	Unknown: "unknown",
	Initial: "initial",
	Playing: "playing",
	Paused:  "paused",
	Stopped: "stopped",
}

var codeToState = map[int32]State{
	0:          Unknown,
	al.Initial: Initial,
	al.Playing: Playing,
	al.Paused:  Paused,
	al.Stopped: Stopped,
}

var device struct {
	sync.Mutex
	d *alc.Device
}

type track struct {
	format           Format
	samplesPerSecond int64
	src              io.ReadSeeker
}

// Player is a basic audio player that plays PCM data.
// Operations on a nil *Player are no-op, a nil *Player can
// be used for testing purposes.
type Player struct {
	t      *track
	source al.Source

	mu        sync.Mutex
	prep      bool
	bufs      []al.Buffer // buffers are created and queued to source during prepare.
	sizeBytes int64       // size of the audio source
}

// NewPlayer returns a new Player.
// It initializes the underlying audio devices and the related resources.
func NewPlayer(src io.ReadSeeker, format Format, samplesPerSecond int64) (*Player, error) {
	device.Lock()
	defer device.Unlock()

	if device.d == nil {
		device.d = alc.Open("")
		c := device.d.CreateContext(nil)
		if !alc.MakeContextCurrent(c) {
			return nil, fmt.Errorf("audio: cannot initiate a new player")
		}
	}
	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		return nil, fmt.Errorf("audio: cannot generate an audio source [err=%x]", code)
	}
	return &Player{
		t:      &track{format: format, src: src, samplesPerSecond: samplesPerSecond},
		source: s[0],
	}, nil
}

func (p *Player) prepare(offset int64, force bool) error {
	p.mu.Lock()
	if !force && p.prep {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	if _, err := p.t.src.Seek(offset, 0); err != nil {
		return err
	}
	var bufs []al.Buffer
	// TODO(jbd): Limit the number of buffers in use, unqueue and reuse
	// the existing buffers as buffers are processed.
	buf := make([]byte, 128*1024)
	size := offset
	for {
		n, err := p.t.src.Read(buf)
		if n > 0 {
			size += int64(n)
			b := al.GenBuffers(1)
			b[0].BufferData(formatCodes[p.t.format], buf[:n], int32(p.t.samplesPerSecond))
			bufs = append(bufs, b[0])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	p.mu.Lock()
	if len(p.bufs) > 0 {
		p.source.UnqueueBuffers(p.bufs)
		al.DeleteBuffers(p.bufs)
	}
	p.sizeBytes = size
	p.bufs = bufs
	p.prep = true
	if len(bufs) > 0 {
		p.source.QueueBuffers(bufs)
	}
	p.mu.Unlock()
	return nil
}

// Play buffers the source audio to the audio device and starts
// to play the source.
// If the player paused or stopped, it reuses the previously buffered
// resources to keep playing from the time it has paused or stopped.
func (p *Player) Play() error {
	if p == nil {
		return nil
	}
	// Prepares if the track hasn't been buffered before.
	if err := p.prepare(0, false); err != nil {
		return err
	}
	al.PlaySources(p.source)
	return lastErr()
}

// Pause pauses the player.
func (p *Player) Pause() error {
	if p == nil {
		return nil
	}
	al.PauseSources(p.source)
	return lastErr()
}

// Stop stops the player.
func (p *Player) Stop() error {
	if p == nil {
		return nil
	}
	al.StopSources(p.source)
	return lastErr()
}

// Seek moves the play head to the given offset relative to the start of the source.
func (p *Player) Seek(offset time.Duration) error {
	if p == nil {
		return nil
	}
	if err := p.Stop(); err != nil {
		return err
	}
	size := durToByteOffset(p.t, offset)
	if err := p.prepare(size, true); err != nil {
		return err
	}
	al.PlaySources(p.source)
	return lastErr()
}

// Current returns the current playback position of the audio that is being played.
func (p *Player) Current() time.Duration {
	if p == nil {
		return 0
	}
	// TODO(jbd): Current never returns the Total when the playing is finished.
	// OpenAL may be returning the last buffer's start point as an OffsetByte.
	return byteOffsetToDur(p.t, int64(p.source.OffsetByte()))
}

// Total returns the total duration of the audio source.
func (p *Player) Total() time.Duration {
	if p == nil {
		return 0
	}
	// Prepare is required to determine the length of the source.
	// We need to read the entire source to calculate the length.
	p.prepare(0, false)
	return byteOffsetToDur(p.t, p.sizeBytes)
}

// Volume returns the current player volume. The range of the volume is [0, 1].
func (p *Player) Volume() float64 {
	if p == nil {
		return 0
	}
	return float64(p.source.Gain())
}

// SetVolume sets the volume of the player. The range of the volume is [0, 1].
func (p *Player) SetVolume(vol float64) {
	if p == nil {
		return
	}
	p.source.SetGain(float32(vol))
}

// State returns the player's current state.
func (p *Player) State() State {
	if p == nil {
		return Unknown
	}
	return codeToState[p.source.State()]
}

// Destroy frees the underlying resources used by the player.
// It should be called as soon as the player is not in-use anymore.
func (p *Player) Destroy() {
	if p == nil {
		return
	}
	if p.source != 0 {
		al.DeleteSources(p.source)
	}
	p.mu.Lock()
	if len(p.bufs) > 0 {
		al.DeleteBuffers(p.bufs)
	}
	p.mu.Unlock()
}

func byteOffsetToDur(t *track, offset int64) time.Duration {
	return time.Duration(offset * formatBytes[t.format] * int64(time.Second) / t.samplesPerSecond)
}

func durToByteOffset(t *track, dur time.Duration) int64 {
	return int64(dur) * t.samplesPerSecond / (formatBytes[t.format] * int64(time.Second))
}

// lastErr returns the last error or nil if the last operation
// has been succesful.
func lastErr() error {
	if code := al.Error(); code != 0 {
		return fmt.Errorf("audio: openal failed with %x", code)
	}
	return nil
}

// TODO(jbd): Destroy context, close the device.
