package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

type byteReader interface {
	io.Reader
	ReadByte() (byte, error)
}

func readByte(r io.Reader) (byte, error) {
	if br, ok := r.(byteReader); ok {
		return br.ReadByte()
	}

	bs := make([]byte, 1)
	if _, err := io.ReadFull(r, bs); err != nil {
		return 0, err
	}
	return bs[0], nil
}

func readBool(r io.Reader) (bool, error) {
	b, err := readByte(r)
	if err == nil {
		return (b == 1), nil
	}
	return false, err
}

func readInt(r io.Reader) (int32, error) {
	var v int
	var b byte
	var err error
	if br, ok := r.(byteReader); ok {
		for shift := uint(0); ; shift += 7 {
			if b, err = br.ReadByte(); err != nil {
				return 0, err
			}
			v |= int(b&127) << shift
			if b&128 == 0 {
				break
			}
		}
	} else {
		buf := make([]byte, 1)
		for shift := uint(0); ; shift += 7 {
			if _, err := io.ReadFull(r, buf); err != nil {
				return 0, err
			}
			b = buf[0]
			v |= int(b&127) << shift
			if b&128 == 0 {
				break
			}
		}
	}
	datum := (int32(v>>1) ^ -int32(v&1))
	return datum, nil
}

func readLong(r io.Reader) (int64, error) {
	var v uint64
	var b byte
	var err error
	if br, ok := r.(byteReader); ok {
		for shift := uint(0); ; shift += 7 {
			if b, err = br.ReadByte(); err != nil {
				return 0, err
			}
			v |= uint64(b&127) << shift
			if b&128 == 0 {
				break
			}
		}
	} else {
		buf := make([]byte, 1)
		for shift := uint(0); ; shift += 7 {
			if _, err = io.ReadFull(r, buf); err != nil {
				return 0, err
			}
			b = buf[0]
			v |= uint64(b&127) << shift
			if b&128 == 0 {
				break
			}
		}
	}
	datum := (int64(v>>1) ^ -int64(v&1))
	return datum, nil
}

func readFloat(r io.Reader) (float32, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint32(buf)
	val := math.Float32frombits(bits)
	return val, nil
}

func readDouble(r io.Reader) (float64, error) {
	buf := make([]byte, 8)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(buf)
	val := math.Float64frombits(bits)
	return val, nil
}

func readString(r io.Reader) (string, error) {
	len, err := readLong(r)
	if err != nil {
		return "", err
	}

	// makeslice can fail depending on available memory.
	// We arbitrarily limit string size to sane default (~2.2GB).
	if len < 0 || len > math.MaxInt32 {
		return "", fmt.Errorf("string length out of range: %d", len)
	}

	if len == 0 {
		return "", nil
	}

	bb := make([]byte, len)
	_, err = io.ReadFull(r, bb)
	if err != nil {
		return "", err
	}
	return string(bb), nil
}

func readBytes(r io.Reader) ([]byte, error) {
	size, err := readLong(r)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return []byte{}, nil
	}
	bb := make([]byte, size)
	_, err = io.ReadFull(r, bb)
	return bb, err
}

func readFixed(r io.Reader, size int) ([]byte, error) {
	bb := make([]byte, size)
	_, err := io.ReadFull(r, bb)
	return bb, err
}
