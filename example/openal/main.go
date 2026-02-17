package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	"golang.org/x/mobile/exp/audio/al"
)

// newSourceBufferChan
func newSourceBufferChan(format uint32, samplerate int32, jitter time.Duration) chan<- []byte {
	var bytesPerSample int
	switch format {
	case al.FormatMono8:
		bytesPerSample = 1
	case al.FormatMono16:
		bytesPerSample = 2
	case al.FormatStereo8:
		bytesPerSample = 2
	case al.FormatStereo16:
		bytesPerSample = 4
	default:
		panic("unknown format")
	}

	ch := make(chan []byte)
	go func() {
		var (
			source        = al.GenSources(1)[0]
			freeBuffers   []al.Buffer
			queuedBuffers []al.Buffer
			queuedSamples []int
		)
		defer func() {
			al.DeleteBuffers(freeBuffers...)
			source.UnqueueBuffers(queuedBuffers...)
			al.DeleteSources(source)
			al.DeleteBuffers(queuedBuffers...)
		}()

		for bts := range ch {
			// Reuse an unqueued buffer or create a new one to put bts in
			var buf al.Buffer
			if len(freeBuffers) > 0 {
				buf = freeBuffers[len(freeBuffers)-1]
				freeBuffers = freeBuffers[:len(freeBuffers)-1]
			} else {
				buf = al.GenBuffers(1)[0]
			}

			// Unqueue processed buffers.
			// When a buffer has finished playing, it should be unqueued
			// from the source unless the intention is to play it in a loop.
			n := source.BuffersProcessed()
			if n > 0 {
				processedBuffers := queuedBuffers[:n]
				queuedBuffers = queuedBuffers[n:]
				queuedSamples = queuedSamples[n:]
				source.UnqueueBuffers(processedBuffers...)
				freeBuffers = append(freeBuffers, processedBuffers...)
			}

			// Set buffer bytes and queue the buffer
			buf.BufferData(format, bts, samplerate)
			source.QueueBuffers(buf)
			queuedBuffers = append(queuedBuffers, buf)
			queuedSamples = append(queuedSamples, len(bts)/bytesPerSample)

			// Play the source.
			// If a source runs out of buffers to play, it will not start
			// playing automatically if a new buffer is queued, and
			// PlaySources should be called again.
			if source.State() != al.Playing {
				al.PlaySources(source)
			}

			// Wait for the buffers to be almost processed minus the jitter
			// duration before accepting another byte slice.
			var samplesRemaining = -uint64(source.OffsetSample())
			for _, samples := range queuedSamples {
				samplesRemaining += uint64(samples)
			}
			durRemaining := time.Duration(samplesRemaining) * time.Second / time.Duration(samplerate)
			time.Sleep(durRemaining - jitter)
		}
	}()
	return ch
}

const (
	samplerate = 44100
	basefreq   = 80.0
	semitone   = 1.0594630943592953 // math.Pow(2, 1./12)
)

func main() {
	al.OpenDevice()
	pcmCh := newSourceBufferChan(al.FormatMono16, samplerate, time.Second/10)

	// the bass line
	pitches := []float64{0, 0, 0, 1, 0, 6, 5, 0.8}
	for i := range pitches {
		pitches[i] = basefreq * math.Pow(semitone, pitches[i])
	}

	// play two full sections of 6ths
	var wubs float64 = 6
	var nRows = 2
introLoop:
	for k := 0; k < nRows; k++ {
		for p, pitch := range pitches {
			pcmCh <- pcm16bit(note(pitch, samplerate, samplerate, wubs))
			if k == nRows-1 && p == len(pitches)-2 {
				break introLoop
			}

			if k != 0 && p == 0 {
				fmt.Println()
			}
			fmt.Printf("%.0fHz ", pitch)
		}
	}

	// 3, 2, 1, let's jam
	fmt.Print("3 ")
	time.Sleep(500 * time.Millisecond)
	fmt.Print("2 ")
	pcmCh <- pcm16bit(note(pitches[len(pitches)-1], samplerate, samplerate, wubs))
	fmt.Print("1 ")
	time.Sleep(250 * time.Millisecond)
	fmt.Print("\nlet's ")
	time.Sleep(250 * time.Millisecond)
	fmt.Print("jam")

	// like a bass drop, but with dots
	t := time.After(500 * time.Millisecond)
dotLoop:
	for i := 1; ; i++ {
		select {
		case <-t:
			break dotLoop
		default:
			fmt.Print(".")
			time.Sleep(100 * time.Millisecond / time.Duration(i))
		}
	}

	// take it away!
	fmt.Println()
	for section := 0; ; section++ {
		// remember rhythm motifs that played in the current section
		// we'll replay some of them at random for pleasant repetition
		motifs := map[string][]int{}

		for _, pitch := range pitches {
			// pick a rhythm for the pitch
			var rhythm []int
			doRepeat := rand.Intn(len(motifs)+1)-1 > 0
			if doRepeat {
				// repeat a random previous motif
				for _, rhythm = range motifs {
					break
				}
			} else {
				// create a new random motif
				rhythm = newRhythm(section)
				motifs[fmt.Sprint(rhythm)] = rhythm
			}

			// generate and play every beat
			for q, quotient := range rhythm {
				samples := samplerate / quotient
				pcmCh <- pcm16bit(note(pitch, samples, samplerate, float64(quotient)))
				if q == 0 {
					fmt.Printf("\n%.0fHz ", pitch)
				}
				fmt.Printf("1/%d ", quotient)
			}
		}
	}
}

func newRhythm(measure int) []int {
	// start with one full note
	rhythm := []int{1}
	// convert a beat selected at random into a tuplet 1-2 times
	n := rand.Intn(2) + 1
	for i := 0; i < n; i++ {
		beat := rand.Intn(len(rhythm))
		beatQuotient := rhythm[beat]

		// duplet, triplet, quadruplet? etc.
		tupletType := rand.Intn(measure+2) + 2

		// copy rhythm to newRhythm, leaving a gap of size tupletType at beat
		newRhythm := make([]int, len(rhythm)+tupletType-1)
		copy(newRhythm[0:beat], rhythm[0:beat])
		copy(newRhythm[beat+tupletType:len(newRhythm)], rhythm[beat:len(rhythm)])

		// divide the beat into a tuplet
		for j := 0; j < tupletType; j++ {
			newRhythm[beat+j] = beatQuotient * tupletType
		}
		rhythm = newRhythm
	}
	return rhythm
}

func note(pitch float64, samples, samplerate int, wubFreq float64) []float64 {
	note := make([]float64, samples)
	for i := 0; i < samples; i++ {
		t := float64(i) / float64(samplerate)
		note[i] = tone(t, pitch) * sine(t, wubFreq/2)
	}
	return note
}

func tone(t float64, freq float64) float64 {
	return .92*sine(t, freq) +
		.04*square(t, freq*1.001) +
		.04*triangle(t, freq*1.003)
}

func sine(t float64, freq float64) float64 {
	return math.Sin(freq * t * (2 * math.Pi))
}

func triangle(t float64, freq float64) float64 {
	return (math.Mod(t, 1/freq)*freq)*2 - 1
}

func square(t float64, freq float64) float64 {
	if math.Mod(t*freq, 1) > 0.5 {
		return 1.0
	}
	return -1.0
}

func pcm16bit(samples []float64) []byte {
	ret := make([]byte, len(samples)*2)
	for i, f := range samples {
		binary.LittleEndian.PutUint16(ret[i*2:], uint16((f*.5+.5)*math.MaxInt16))
	}
	return ret
}
