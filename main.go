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
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./tde filename")
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Error reading image: ", err)
	}

	MagicNumbers := make([]byte, 8)

	if _, err = io.ReadFull(f, MagicNumbers); err != nil {
		log.Fatal("File couldn't be read: ", err)
	}

	if !bytes.Equal([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, MagicNumbers) {
		log.Fatal("Image is NOT a PNG!")
	}
	//log.Println("Image Read Succesfully")

	buffer := make([]byte, 4)
	blobLen := make([]byte, 4)
	
	for {
		if _, err := io.ReadFull(f, buffer); err != nil {
			break
		}
		if bytes.Equal([]byte{'I', 'H', 'D', 'R'}, buffer) {
			ihdr_data := make([]byte, 13)
			io.ReadFull(f, ihdr_data)
		}
		if bytes.Equal([]byte{'t', 'E', 'X', 't'}, buffer) {
			var identifier []byte
			bufferone := make([]byte, 1)
			for {
				if _, err = io.ReadFull(f, bufferone); err != nil {
					log.Fatal(err)
				}
				if bytes.Equal([]byte{0x00}, bufferone) {
					break
				}
				identifier = append(identifier, bufferone...)
			}
			if bytes.Equal([]byte{'c', 'h', 'a', 'r', 'a'}, identifier) {
				//log.Println("Character definitions found")
				content := make([]byte, binary.BigEndian.Uint32(blobLen) - 6) // 6 here is the len of chara + null separator
				//log.Println(binary.BigEndian.Uint32(blobLen))
				if _, err := io.ReadFull(f, content); err != nil { log.Fatal(err) }

				crcchecksum := make([]byte, 4)
				if _, err := io.ReadFull(f, crcchecksum); err != nil { log.Fatal(err) }		
				
				if crc32.ChecksumIEEE(append([]byte{'t', 'E', 'X', 't', 'c', 'h', 'a', 'r', 'a', 0x00}, content...)) != binary.BigEndian.Uint32(crcchecksum) {
					log.Fatal("CRC Checksums do not match. File corrupted or incorrect")
				}
				
				decoded, err := base64.StdEncoding.DecodeString(string(content))
				if err != nil { log.Fatal(err)  }
				fmt.Printf("%s\n", decoded);
			}
		}
		copy(blobLen, buffer)
	}
}
