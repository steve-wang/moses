package moses

import (
	"errors"
	"io"
)

const (
	PACK_MAXSIZE = 16 * 1024
)

var (
	ERR_EXIT = errors.New("EXIT")
)

type PrivateReadWriter struct {
	io.Reader
	io.Writer
	key  [4]byte
	data []byte
}

func NewPrivateReadWriter(rw io.ReadWriter, key [4]byte) *PrivateReadWriter {
	return &PrivateReadWriter{
		Reader: rw,
		Writer: rw,
		key:    key}
}

func (p *PrivateReadWriter) unsize() (int, error) {
	n, err := readUint16(p.Reader)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (p *PrivateReadWriter) unpack(data []byte) error {
	if err := read(p.Reader, data); err != nil {
		return err
	}
	code(data, p.key)
	return nil
}

func (p *PrivateReadWriter) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	index := 0
	if len(p.data) > 0 {
		index += copy(data, p.data)
		p.data = p.data[index:]
		if index == len(data) {
			return index, nil
		}
	}
	// p.data is empty now
	size, err := p.unsize()
	if err != nil {
		return 0, err
	}
	if len(data[index:]) >= size {
		if err := p.unpack(data[index:index+size]); err != nil {
			return index, err
		}
		index += size
	} else {
		buff := make([]byte, size)
		if err := p.unpack(buff); err != nil {
			return index, err
		}
		n := len(data[index:])
		copy(data[index:], buff[:n])
		p.data = buff[n:]
		index += n
	}
	return index, nil
}

func (p *PrivateReadWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	index := 0
	for index < len(data) {
		size := len(data[index:])
		if size > PACK_MAXSIZE {
			size = PACK_MAXSIZE
		}
		if err := writeUint16(p.Writer, uint16(size)); err != nil {
			return 0, err
		}
		chip := data[index : index+size]
		code(chip, p.key)
		if err := write(p.Writer, chip); err != nil {
			return 0, err
		}
		index += size
	}
	return len(data), nil
}
