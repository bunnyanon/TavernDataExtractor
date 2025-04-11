package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./tde filename")
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("couldn't open file: ", err)
	}
	data, err := extractData(f)
	if err != nil {
		log.Fatal("couldn't extract data: ", err)
	}
	fmt.Println(string(data))
}

func extractData(f io.Reader) ([]byte, error) {
	MagicNumbers := make([]byte, 8)

	if _, err := io.ReadFull(f, MagicNumbers); err != nil {
		return nil, err
	}

	if !bytes.Equal([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, MagicNumbers) {
		return nil, errors.New("file is NOT a PNG")
	}

	blobLen := make([]byte, 4)
	header := make([]byte, 4)
	crcSum := make([]byte, 4)

	for {
		if _, err := io.ReadFull(f, blobLen); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if _, err := io.ReadFull(f, header); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		content := make([]byte, binary.BigEndian.Uint32(blobLen))
		if _, err := io.ReadFull(f, content); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if _, err := io.ReadFull(f, crcSum); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if crc32.ChecksumIEEE(append(header, content...)) != binary.BigEndian.Uint32(crcSum) {
			return nil, errors.New("CRC32 checksums don't match. Possible sign of file corruption")
		}
		if bytes.Equal([]byte{'t', 'E', 'X', 't'}, header) && bytes.Equal([]byte{'c', 'h', 'a', 'r', 'a', 0x0}, content[:6]) {
			decoded, err := base64.StdEncoding.DecodeString(string(content[6:]))
			if err != nil {
				return nil, err
			}
			return decoded, nil
		}
	}
	return nil, errors.New("no tavern data found")
}
