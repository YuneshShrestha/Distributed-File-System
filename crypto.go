package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	// "fmt"
	"io"
)

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

func copyDecrypt(dst io.Writer, src io.Reader, key []byte) (int64, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// Read the IV from the IO reader which in our case should be the
	// block.BlockSize() bytes we read.
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(src, iv); err != nil {
		return 0, err
	}
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		rw     = block.BlockSize()
	)
	for {

		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			rw += nn

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

	}
	return int64(rw), nil
}

func copyEncrypt(dst io.Writer, src io.Reader, key []byte) (int64, error) {
	// Initialize the AES cipher block with the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // This line of Go code is creating a byte slice to be used as an initialization vector (IV) for AES encryption. IV is crucial for security of the encryption process.

	// Generate a random IV.
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// buf is used to temporarily store the data read from the source before it is encrypted and written to the destination.
	// stream is an instance of the cipher.Stream interface that is used to encrypt the data read from the source.

	// prepend the IV to the destination which is used for decryption.
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		rw    = block.BlockSize() 
	)
	for {
		n, err := src.Read(buf) // n = 5 for Hello

		if n > 0 {
			stream.XORKeyStream(buf, buf[:n]) // This line of Go code is encrypting the data read from the source using the CTR mode of operation. CTR mode is a stream cipher mode of operation that allows encryption of individual bytes of data.
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			rw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

	}
	return int64(rw), nil

}
