// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

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

// Format represents an audio file format.
type Format string

const (
	Mono8    = Format("mono8")
	Mono16   = Format("mono16")
	Stereo8  = Format("stereo8")
	Stereo16 = Format("stereo16")
)

var formatToCode = map[Format]int32{
	Mono8:    0x1100,
	Mono16:   0x1101,
	Stereo8:  0x1102,
	Stereo16: 0x1103,
}

// State indicates the current playing state of the player.
type State string

const (
	Unknown = State("unknown")
	Initial = State("initial")
	Playing = State("playing")
	Paused  = State("paused")
	Stopped = State("stopped")
)

var codeToState = map[int32]State{
	0:      Unknown,
	0x1011: Initial,
	0x1012: Playing,
	0x1013: Paused,
	0x1014: Stopped,
}

var device struct {
	sync.Mutex
	d *alc.Device
}

type track struct {
	format Format
	rate   int64 // sample rate
	src    io.ReadSeeker
}

// Player is a basic audio player that plays PCM data.
// Operations on a nil *Player are no-op, a nil *Player can
// be used for testing purposes.
type Player struct {
	t      *track
	source al.Source

	muPrep sync.Mutex // guards prep and bufs
	prep   bool
	bufs   []al.Buffer
	size   int64 // size of the audio source
}

// NewPlayer returns a new Player.
// It initializes the underlying audio devices and the related resources.
func NewPlayer(src io.ReadSeeker, format Format, sampleRate int64) (*Player, error) {
	device.Lock()
	defer device.Unlock()

	if device.d == nil {
		device.d = alc.Open("")
		c := device.d.CreateContext(nil)
		if !alc.MakeContextCurrent(c) {
			return nil, fmt.Errorf("player: cannot initiate a new player")
		}
	}
	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		return nil, fmt.Errorf("player: cannot generate an audio source [err=%x]", code)
	}
	bufs := al.GenBuffers(2)
	if err := lastErr(); err != nil {
		return nil, err
	}
	return &Player{
		t:      &track{format: format, src: src, rate: sampleRate},
		source: s[0],
		bufs:   bufs,
	}, nil
}

func (p *Player) prepare(offset int64, force bool) error {
	p.muPrep.Lock()
	defer p.muPrep.Unlock()
	if !force && p.prep {
		return nil
	}
	if len(p.bufs) > 0 {
		p.source.UnqueueBuffers(p.bufs)
		al.DeleteBuffers(p.bufs)
	}
	if _, err := p.t.src.Seek(offset, 0); err != nil {
		return err
	}
	p.bufs = []al.Buffer{}
	// TODO(jbd): Limit the number of buffers in use, unqueue and reuse
	// the existing buffers as buffers are processed.
	buf := make([]byte, 128*1024)
	size := offset
	for {
		n, err := p.t.src.Read(buf)
		if n > 0 {
			size += int64(n)
			b := al.GenBuffers(1)
			b[0].BufferData(uint32(formatToCode[p.t.format]), buf[:n], int32(p.t.rate))
			p.bufs = append(p.bufs, b[0])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	p.size = size
	if len(p.bufs) > 0 {
		p.source.QueueBuffers(p.bufs)
	}
	p.prep = true
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
		return time.Duration(0)
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
	return byteOffsetToDur(p.t, p.size)
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
	p.muPrep.Lock()
	if len(p.bufs) > 0 {
		al.DeleteBuffers(p.bufs)
	}
	p.muPrep.Unlock()
}

func byteOffsetToDur(t *track, offset int64) time.Duration {
	size := float64(offset)
	if t.format == Mono16 || t.format == Stereo16 {
		size /= 2
	}
	if t.format == Stereo8 || t.format == Stereo16 {
		size /= 2
	}
	size /= float64(t.rate)
	// Casting size back to int64. Work in milliseconds,
	// so that size doesn't get rounded excessively.
	return time.Duration(size*1000) * time.Duration(time.Millisecond)
}

func durToByteOffset(t *track, dur time.Duration) int64 {
	size := int64(dur/time.Second) * t.rate
	// Each sample is represented by 16-bits. Move twice further.
	if t.format == Mono16 || t.format == Stereo16 {
		size *= 2
	}
	if t.format == Stereo8 || t.format == Stereo16 {
		size *= 2
	}
	return size
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
