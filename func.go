package moses

import (
	"errors"
	"io"
	"net"
)

func read(r io.Reader, data []byte) error {
	index := 0
	for index < len(data) {
		n, err := r.Read(data[index:])
		if n > 0 {
			index += n
		}
		if err != nil {
			if err == io.EOF && index == len(data) {
				break
			}
			return err
		}
	}
	return nil
}

func write(w io.Writer, data []byte) error {
	index := 0
	for index < len(data) {
		n, err := w.Write(data[index:])
		if err != nil {
			if nerr, ok := err.(net.Error); !ok || !nerr.Temporary() {
				return err
			}
		}
		index += n
	}
	return nil
}

func code(data []byte, key [4]byte) {
	for len(data) >= 4 {
		data[0] ^= key[0]
		data[1] ^= key[1]
		data[2] ^= key[2]
		data[3] ^= key[3]
		data = data[4:]
	}
	for i := 0; i < len(data); i++ {
		data[i] ^= key[i]
	}
}

func readUint16(r io.Reader) (uint16, error) {
	var data [2]byte
	if err := read(r, data[:]); err != nil {
		return 0, err
	}
	n := uint16(data[0])
	n |= uint16(data[1]) << 8
	return n, nil
}

func writeUint16(w io.Writer, n uint16) error {
	var data [2]byte
	data[0] = byte(n & 0xff)
	data[1] = byte((n >> 8) & 0xff)
	return write(w, data[:])
}

func readUint32(r io.Reader) (uint32, error) {
	var data [4]byte
	if err := read(r, data[:]); err != nil {
		return 0, err
	}
	n := uint32(data[0])
	n |= uint32(data[1]) << 8
	n |= uint32(data[1]) << 16
	n |= uint32(data[1]) << 24
	return n, nil
}

func writeUint32(w io.Writer, n uint32) error {
	var data [4]byte
	data[0] = byte(n & 0xff)
	data[1] = byte((n >> 8) & 0xff)
	data[2] = byte((n >> 16) & 0xff)
	data[3] = byte((n >> 24) & 0xff)
	return write(w, data[:])
}

func readString(r io.Reader) (string, error) {
	size, err := readUint16(r)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", nil
	}
	data := make([]byte, int(size))
	if err := read(r, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func writeString(w io.Writer, str string) error {
	if len(str) > 65536 {
		return errors.New("too long string")
	}
	size := uint16(len(str))
	if err := writeUint16(w, size); err != nil {
		return err
	}
	if size == 0 {
		return nil
	}
	return write(w, []byte(str))
}
