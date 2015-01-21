// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

// Package al provides OpenAL Soft bindings for Go.
//
// More information about OpenAL Soft is available at
// http://www.openal.org/documentation/openal-1.1-specification.pdf.
package al

/*
// TODO(jbd,crawshaw): Add android, linux and windows support.

#cgo darwin   CFLAGS: -DGOOS_darwin
#cgo darwin  LDFLAGS: -framework OpenAL

#ifdef GOOS_darwin
#include <stdlib.h>
#include <OpenAL/al.h>
#endif

*/
import "C"
import "unsafe"

// Enable enables a capability.
func Enable(capability int32) {
	C.alEnable(C.ALenum(capability))
}

// Disable disables a capability.
func Disable(capability int32) {
	C.alDisable(C.ALenum(capability))
}

// Enabled returns true if the specified capability is enabled.
func Enabled(capability int32) bool {
	return C.alIsEnabled(C.ALenum(capability)) == 1
}

// Vector represents an vector in a Cartesian coordinate system.
type Vector [3]float32

// Orientation represents the angular position of an object in a
// right-handed Cartesian coordinate system.
// A cross product between the forward and up vector returns a vector
// that points to the right.
type Orientation struct {
	// Forward vector is the direction that the object is looking at.
	Forward Vector
	// Up vector represents the rotation of the object.
	Up Vector
}

func orientationFromSlice(v []float32) Orientation {
	return Orientation{
		Forward: Vector{v[0], v[1], v[2]},
		Up:      Vector{v[3], v[4], v[5]},
	}
}

func (v Orientation) slice() []float32 {
	return []float32{v.Forward[0], v.Forward[1], v.Forward[2], v.Up[0], v.Up[1], v.Up[2]}
}

func geti(param int) int32 {
	return int32(C.alGetInteger(C.ALenum(param)))
}

func getf(param int) float32 {
	return float32(C.alGetFloat(C.ALenum(param)))
}

func getString(param int) string {
	v := C.alGetString(C.ALenum(param))
	return C.GoString((*C.char)(v))
}

// DistanceModel returns the distance model.
func DistanceModel() int32 {
	return geti(paramDistanceModel)
}

// SetDistanceModel sets the distance model.
func SetDistanceModel(v int32) {
	C.alDistanceModel(C.ALenum(v))
}

// DopplerFactor returns the doppler factor.
func DopplerFactor() float32 {
	return getf(paramDopplerFactor)
}

// SetDopplerFactor sets the doppler factor.
func SetDopplerFactor(v float32) {
	C.alDopplerFactor(C.ALfloat(v))
}

// DopplerVelocity returns the doppler velocity.
func DopplerVelocity() float32 {
	return getf(paramDopplerVelocity)
}

// SetDopplerVelocity sets the doppler velocity.
func SetDopplerVelocity(v float32) {
	C.alDopplerVelocity(C.ALfloat(v))
}

// SpeedOfSound is the speed of sound in meters per second (m/s).
func SpeedOfSound() float32 {
	return getf(paramSpeedOfSound)
}

// SetSpeedOfSound sets the speed of sound, its unit should be meters per second (m/s).
func SetSpeedOfSound(v float32) {
	C.alSpeedOfSound(C.ALfloat(v))
}

// Vendor returns the vendor.
func Vendor() string {
	return getString(paramVendor)
}

// Version returns the version string.
func Version() string {
	return getString(paramVersion)
}

// Renderer returns the renderer information.
func Renderer() string {
	return getString(paramRenderer)
}

// Extensions returns the enabled extensions.
func Extensions() string {
	return getString(paramExtensions)
}

// Error returns the most recently generated error.
func Error() int32 {
	return int32(C.alGetError())
}

// Source represents an individual sound source in 3D-space.
// They take PCM data, apply modifications and then submit them to
// be mixed according to their spatial location.
type Source uint32

// GenSources generates n new sources. These sources should be deleted
// once they are not in use.
func GenSources(n int) []Source {
	s := make([]Source, n)
	C.alGenSources(C.ALsizei(n), (*C.ALuint)(unsafe.Pointer(&s[0])))
	return s
}

// PlaySources plays the sources.
func PlaySources(source ...Source) {
	C.alSourcePlayv(C.ALsizei(len(source)), (*C.ALuint)(unsafe.Pointer(&source[0])))
}

// PauseSources pauses the sources.
func PauseSources(source ...Source) {
	C.alSourcePausev(C.ALsizei(len(source)), (*C.ALuint)(unsafe.Pointer(&source[0])))
}

// StopSources stops the sources.
func StopSources(source ...Source) {
	C.alSourceStopv(C.ALsizei(len(source)), (*C.ALuint)(unsafe.Pointer(&source[0])))
}

// RewindSources rewinds the sources to their beginning positions.
func RewindSources(source ...Source) {
	C.alSourceRewindv(C.ALsizei(len(source)), (*C.ALuint)(unsafe.Pointer(&source[0])))
}

// DeleteSources deletes the sources.
func DeleteSources(source ...Source) {
	C.alDeleteSources(C.ALsizei(len(source)), (*C.ALuint)(unsafe.Pointer(&source[0])))
}

// Gain returns the source gain.
func (s Source) Gain() float32 {
	return getSourcef(s, paramGain)
}

// SetGain sets the source gain.
func (s Source) SetGain(v float32) {
	setSourcef(s, paramGain, v)
}

// MinGain returns the source's minimum gain setting.
func (s Source) MinGain() float32 {
	return getSourcef(s, paramMinGain)
}

// SetMinGain sets the source's minimum gain setting.
func (s Source) SetMinGain(v float32) {
	setSourcef(s, paramMinGain, v)
}

// MaxGain returns the source's maximum gain setting.
func (s Source) MaxGain() float32 {
	return getSourcef(s, paramMaxGain)
}

// SetMaxGain sets the source's maximum gain setting.
func (s Source) SetMaxGain(v float32) {
	setSourcef(s, paramMaxGain, v)
}

// Position returns the position of the source.
func (s Source) Position() Vector {
	v := Vector{}
	getSourcefv(s, paramPosition, v[:])
	return v
}

// SetPosition sets the position of the source.
func (s Source) SetPosition(v Vector) {
	setSourcefv(s, paramPosition, v[:])
}

// Velocity returns the source's velocity.
func (s Source) Velocity() Vector {
	v := Vector{}
	getSourcefv(s, paramVelocity, v[:])
	return v
}

// SetVelocity sets the source's velocity.
func (s Source) SetVelocity(v Vector) {
	setSourcefv(s, paramVelocity, v[:])
}

// Orientation returns the orientation of the source.
func (s Source) Orientation() Orientation {
	v := make([]float32, 6)
	getSourcefv(s, paramOrientation, v)
	return orientationFromSlice(v)
}

// SetOrientation sets the orientation of the source.
func (s Source) SetOrientation(o Orientation) {
	setSourcefv(s, paramOrientation, o.slice())
}

// State returns the playing state of the source.
func (s Source) State() int32 {
	return getSourcei(s, paramSourceState)
}

// BuffersQueued returns the number of the queued buffers.
func (s Source) BuffersQueued() int32 {
	return getSourcei(s, paramBuffersQueued)
}

// BuffersProcessed returns the number of the processed buffers.
func (s Source) BuffersProcessed() int32 {
	return getSourcei(s, paramBuffersProcessed)
}

// OffsetSeconds returns the current playback position of the source in seconds.
func (s Source) OffsetSeconds() int32 {
	return getSourcei(s, paramSecOffset)
}

// OffsetSample returns the sample offset of the current playback position.
func (s Source) OffsetSample() int32 {
	return getSourcei(s, paramSampleOffset)
}

// OffsetByte returns the byte offset of the current playback position.
func (s Source) OffsetByte() int32 {
	return getSourcei(s, paramByteOffset)
}

func getSourcei(s Source, param int) int32 {
	var v C.ALint
	C.alGetSourcei(C.ALuint(s), C.ALenum(param), &v)
	return int32(v)
}

func getSourcef(s Source, param int) float32 {
	var v C.ALfloat
	C.alGetSourcef(C.ALuint(s), C.ALenum(param), &v)
	return float32(v)
}

func getSourcefv(s Source, param int, v []float32) {
	C.alGetSourcefv(C.ALuint(s), C.ALenum(param), (*C.ALfloat)(unsafe.Pointer(&v[0])))
}

func setSourcei(s Source, param int, v int32) {
	C.alSourcei(C.ALuint(s), C.ALenum(param), C.ALint(v))
}

func setSourcef(s Source, param int, v float32) {
	C.alSourcef(C.ALuint(s), C.ALenum(param), C.ALfloat(v))
}

func setSourcefv(s Source, param int, v []float32) {
	C.alSourcefv(C.ALuint(s), C.ALenum(param), (*C.ALfloat)(unsafe.Pointer(&v[0])))
}

// QueueBuffers adds the buffers to the buffer queue.
func (s Source) QueueBuffers(buffers []Buffer) {
	C.alSourceQueueBuffers(C.ALuint(s), C.ALsizei(len(buffers)), (*C.ALuint)(unsafe.Pointer(&buffers[0])))
}

// UnqueueBuffers removes the specified buffers from the buffer queue.
func (s Source) UnqueueBuffers(buffers []Buffer) {
	C.alSourceUnqueueBuffers(C.ALuint(s), C.ALsizei(len(buffers)), (*C.ALuint)(unsafe.Pointer(&buffers[0])))
}

// ListenerGain returns the total gain applied to the final mix.
func ListenerGain() float32 {
	return getListenerf(paramGain)
}

// ListenerPosition returns the position of the listener.
func ListenerPosition() Vector {
	v := Vector{}
	getListenerfv(paramPosition, v[:])
	return v
}

// ListenerVelocity returns the velocity of the listener.
func ListenerVelocity() Vector {
	v := Vector{}
	getListenerfv(paramVelocity, v[:])
	return v
}

// ListenerOrientation returns the orientation of the listener.
func ListenerOrientation() Orientation {
	v := make([]float32, 6)
	getListenerfv(paramOrientation, v)
	return orientationFromSlice(v)
}

// SetListenerGain sets the total gain that will be applied to the final mix.
func SetListenerGain(v float32) {
	setListenerf(paramGain, v)
}

// SetListenerPosition sets the position of the listener.
func SetListenerPosition(v Vector) {
	setListenerfv(paramPosition, v[:])
}

// SetListenerVelocity sets the velocity of the listener.
func SetListenerVelocity(v Vector) {
	setListenerfv(paramVelocity, v[:])
}

// SetListenerOrientation sets the orientation of the listener.
func SetListenerOrientation(v Orientation) {
	setListenerfv(paramOrientation, v.slice())
}

func getListenerf(param int) float32 {
	var v C.ALfloat
	C.alGetListenerf(C.ALenum(param), &v)
	return float32(v)
}

func getListenerfv(param int, v []float32) {
	C.alGetListenerfv(C.ALenum(param), (*C.ALfloat)(unsafe.Pointer(&v[0])))
}

func setListenerf(param int, v float32) {
	C.alListenerf(C.ALenum(param), C.ALfloat(v))
}

func setListenerfv(param int, v []float32) {
	C.alListenerfv(C.ALenum(param), (*C.ALfloat)(unsafe.Pointer(&v[0])))
}

// A buffer represents a chunk of PCM audio data that could be buffered to an audio
// source. A single buffer could be shared between multiple sources.
type Buffer uint32

// GenBuffers generates n new buffers. The generated buffers should be deleted
// once they are no longer in use.
func GenBuffers(n int) []Buffer {
	s := make([]Buffer, n)
	C.alGenBuffers(C.ALsizei(n), (*C.ALuint)(unsafe.Pointer(&s[0])))
	return s
}

// DeleteBuffers deletes the buffers.
func DeleteBuffers(buffers []Buffer) {
	C.alDeleteBuffers(C.ALsizei(len(buffers)), (*C.ALuint)(unsafe.Pointer(&buffers[0])))
}

func getBufferi(b Buffer, param int) int32 {
	var v C.ALint
	C.alGetBufferi(C.ALuint(b), C.ALenum(param), &v)
	return int32(v)
}

// Frequency returns the frequency of the buffer data in Hertz (Hz).
func (b Buffer) Frequency() int32 {
	return getBufferi(b, paramFreq)
}

// Bits return the number of bits used to represent a sample.
func (b Buffer) Bits() int32 {
	return getBufferi(b, paramBits)
}

// Channels return the number of the audio channels.
func (b Buffer) Channels() int32 {
	return getBufferi(b, paramChannels)
}

// Size returns the size of the data.
func (b Buffer) Size() int32 {
	return getBufferi(b, paramSize)
}

// BufferData buffers PCM data to the current buffer.
func (b Buffer) BufferData(format uint32, data []byte, freq int32) {
	C.alBufferData(C.ALuint(b), C.ALenum(format), unsafe.Pointer(&data[0]), C.ALsizei(len(data)), C.ALsizei(freq))
}

// Valid returns true if the buffer exists and is valid.
func (b Buffer) Valid() bool {
	return C.alIsBuffer(C.ALuint(b)) == 1
}
