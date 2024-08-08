package main

import (
	"hash/crc32"
	"bytes"
	"fmt"
	"encoding/binary"
	"io"
	"log"
	"os"
	"encoding/base64"
	"errors"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./tde filename")
		return
	}
	f, err := os.Open(os.Args[1])

	MagicNumbers := make([]byte, 8)

	if _, err = io.ReadFull(f, MagicNumbers); err != nil {
		log.Fatal("File couldn't be read: ", err)
	}

	if !bytes.Equal([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, MagicNumbers) {
		log.Fatal("Image is NOT a PNG!")
	}
	
	blobLen := make([]byte, 4)
	header := make([]byte, 4)
	crcSum := make([]byte, 4)
	
	for {
    if _, err := io.ReadFull(f, blobLen); err != nil {
      if errors.Is(err, io.EOF) { break }
      log.Fatal(err)
    }
    if _, err := io.ReadFull(f, header); err != nil {
      if errors.Is(err, io.EOF) { break }
      log.Fatal(err)
    }
    
    content := make([]byte, binary.BigEndian.Uint32(blobLen))
    if _, err := io.ReadFull(f, content); err != nil {
      if errors.Is(err, io.EOF) { break }
      log.Fatal(err)
    }
    if _, err := io.ReadFull(f, crcSum); err != nil {
      if errors.Is(err, io.EOF) { break }
      log.Fatal(err)
    }
    
    if crc32.ChecksumIEEE(append(header, content...)) != binary.BigEndian.Uint32(crcSum) {
      log.Fatalf("CRC32 checksums of header %s don't match. Possible sign of file corruption\n", header)
    }
    if bytes.Equal([]byte{'t', 'E', 'X', 't'}, header) {
      if bytes.Equal([]byte{'c', 'h', 'a', 'r', 'a', 0x0}, content[:6]) {
        decoded, err := base64.StdEncoding.DecodeString(string(content[6:]))
        if err != nil { log.Fatal(err) }
        fmt.Printf("%s\n", decoded)
        os.Exit(0)
      }
    }
  }
}
