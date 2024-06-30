package main

import (
	"fmt"
	"hash/crc32"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	spiDev     = "/dev/spidev0.1"
	maxRetries = 3 // Maksimum tekrar denemesi
)
var spiSpeed uint32 = 500000
func openSPI() (*os.File, error) {
	file, err := os.OpenFile(spiDev, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	// SPI hızı ayarla
	if err := ioctl(file.Fd(), 0x40046b04, uintptr(unsafe.Pointer(&spiSpeed))); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func ioctl(fd, cmd, arg uintptr) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, arg); err != 0 {
		return err
	}
	return nil
}

func readSPI(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		return err
	}
	if _, err := file.Read(data); err != nil {
		return err
	}
	return nil
}

func main() {
	file, err := openSPI()
	if err != nil {
		fmt.Println("SPI açma hatası:", err)
		return
	}
	defer file.Close()

	for {
		time.Sleep(1 * time.Second)

		success := false

		for retries := 0; retries < maxRetries; retries++ {
			// Veri alma
			data := make([]byte, 4)
			if err := readSPI(file, data); err != nil {
				fmt.Println("Veri alma hatası:", err)
				continue
			}

			// CRC alma
			crcReceived := make([]byte, 4)
			if err := readSPI(file, crcReceived); err != nil {
				fmt.Println("CRC alma hatası:", err)
				continue
			}

			// CRC hesaplama
			crcCalculated := crc32.ChecksumIEEE(data)

			// CRC doğrulama
			if crc32.ChecksumIEEE(data) == crcCalculated {
				fmt.Println("Veri doğru:", data)
				success = true
				break
			} else {
				fmt.Println("Veri hatalı, tekrar deneme:", retries+1)
			}
		}

		if !success {
			fmt.Println("Veri alımı başarısız oldu.")
		}
	}
}
