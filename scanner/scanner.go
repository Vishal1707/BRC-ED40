// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package scanner provides functions for reading barcode scans from
// usb-connected barcode scanner devices as if they were keyboards, i.e.,
// by using the corresponding '/dev/input/event' device, inspired by this
// post on linuxquestions.org:
//
// http://www.linuxquestions.org/questions/programming-9/read-from-a-usb-barcode-scanner-that-simulates-a-keyboard-495358/#post2767643
//
// Also found important Go-specific information by reviewing the code from
// this repo on github:
//
// https://github.com/gvalkov/golang-evdev

package scanner

import (
	"bytes"
	"encoding/binary"
	"github.com/gvalkov/golang-evdev"
	"os"
	"syscall"
	"unsafe"
)

const (
	EVENT_BUFFER   = 64
	EVENT_CAPTURES = 16
)

var (
	scanned        bool
	scanner_device string = "/dev/input/event3"
)

// InputEvent is a Go implementation of the native linux device
// input_event struct, as described in the kernel documentation
// (https://www.kernel.org/doc/Documentation/input/input.txt),
// with a big assist from https://github.com/gvalkov/golang-evdev
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

var EVENT_SIZE = int(unsafe.Sizeof(InputEvent{}))

// KEYCODES is the map of hex found in the InputEvent.Code field, and
// its corresponding char (string) representation
// [source: Vojtech Pavlik (author of the Linux Input Drivers project),
// via linuxquestions.org user bricedebrignaisplage]
var KEYCODES = map[byte]string{
	0x02: "1",
	0x03: "2",
	0x04: "3",
	0x05: "4",
	0x06: "5",
	0x07: "6", 0x08: "7", 0x09: "8", 0x0a: "9",
	0x0b: "0",
	0x0c: "-", 0x10: "q",
	0x11: "w", 0x12: "e",
	0x13: "r",
	0x14: "t", 0x15: "y",
	0x16: "u",
	0x17: "i",
	0x18: "o", 0x19: "p",
	0x1e: "a",
	0x1f: "s",
	0x20: "d",
	0x21: "f",
	0x22: "g",
	0x23: "h",
	0x24: "j",
	0x25: "k",
	0x26: "l",
	0x2c: "z",
	0x2d: "x",
	0x2e: "c",
	0x2f: "v",
	0x30: "b",
	0x31: "n",
	0x32: "m",
}

// lookupKeyCode finds the corresponding string for the given hex byte,
// returning "-" as the default if not found
func lookupKeyCode(b byte) string {
	val, exists := KEYCODES[b]
	if exists {
		return val
	} else {
		return "-"
	}
}

// read takes the open scanner device pointer and returns a list of
// InputEvent captures, corresponding to input (scan) events
func read(dev *os.File) ([]InputEvent, error) {
	events := make([]InputEvent, EVENT_CAPTURES)
	buffer := make([]byte, EVENT_SIZE*EVENT_CAPTURES)
	_, err := dev.Read(buffer)
	if err != nil {
		return events, err
	}
	b := bytes.NewBuffer(buffer)
	err = binary.Read(b, binary.LittleEndian, &events)
	if err != nil {
		return events, err
	}

	// remove trailing structures
	for i := range events {
		if events[i].Time.Sec == 0 {
			events = append(events[:i])
			break
		}
	}
	return events, err
}

// DecodeEvents iterates through the list of InputEvents and decodes
// the barcode data into a string, along with a boolean to indicate if this
// particular input sequence is done
func decodeEvents(events []InputEvent) (string, bool) {
	var buffer bytes.Buffer
	for i := range events {
		if events[i].Type == 1 && events[i].Value == 1 {
			if events[i].Code == 28 {
				// carriage return detected: the barcode sequence ends here
				return buffer.String(), true
			} else {
				if events[i].Code != 0 {
					// this is barcode data we want to capture
					buffer.WriteString(lookupKeyCode(byte(events[i].Code)))
				}
			}
		}
	}
	// return what has been collected so far,
	// even though the barcode is not yet complete
	return buffer.String(), false
}

// FindDevicePath find all the input devices connected,
// Scanner being used has Vendor code "1662" and product code "2049"
// It is checked to the devices obtained and device path is found
func findDevicePath() {
	devices, _ := evdev.ListInputDevices()

	for _, dev := range devices {
		if dev.Vendor == 1662 && dev.Product == 2049 {
			scanner_device = dev.Fn
		}
	}

}

// ScanForever takes a linux input device string pointing to the scanner
// to read from, invokes the given function on the resulting barcode string
// when complete, or the errfn on error, then goes back to read/scan again
func ScanForever(fn func(string)) {
	findDevicePath()
	scanner, err := os.Open(scanner_device)
	if err != nil {
		// invoke the function which handles scanner errors
	}
	defer scanner.Close()

	var scanBuffer bytes.Buffer
	for {
		scanEvents, scanErr := read(scanner)
		if scanErr != nil {
			// invoke the function which handles scanner errors
		}
		scannedData, endOfScan := decodeEvents(scanEvents)
		if endOfScan {
			// invoke the function which handles the scan result
			fn(scanBuffer.String())
			scanBuffer.Reset() // clear the buffer and start again
		} else {
			scanBuffer.WriteString(scannedData)
		}
	}
}
