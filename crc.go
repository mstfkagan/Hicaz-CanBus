package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"os/exec"
	"time"

	"github.com/knieriem/gocan"
)

// Belirli bir ID'ye sahip mesajları filtrelemek için
const targetID = 0x123

// Hata limiti
const maxErrors = 5

func main() {
	// CAN arayüzünü aç
	dev := "can0"
	s, err := gocan.NewDevSocket(dev)
	if err != nil {
		log.Fatalf("CAN arayüzü açılırken hata: %v", err)
	}
	defer s.Close()

	errorCount := 0

	// Mesaj alma ve doğrulama
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var msg gocan.CANFrame
		if err := s.RecvFrame(&msg); err != nil {
			log.Printf("Mesaj alınırken hata: %v", err)
			errorCount++
			if errorCount >= maxErrors {
				log.Println("Hata limiti aşıldı, restart.go dosyası çalıştırılıyor...")
				if err := runRestartScript(); err != nil {
					log.Fatalf("restart.go dosyası çalıştırılamadı: %v", err)
				}
				errorCount = 0
			}
			continue
		}

		// Hatalı mesajları sıfırla
		errorCount = 0

		// Belirli bir ID'ye sahip mesajları filtrele
		if msg.ID == targetID {
			if validateMessage(msg.Data[:]) {
				fmt.Println("Mesaj doğrulandı:", msg.Data)
			} else {
				fmt.Println("Mesaj doğrulama hatası:", msg.Data)
			}
		}
	}
}

// CRC doğrulama fonksiyonu
func validateMessage(data []byte) bool {
	if len(data) < 12 {
		return false
	}

	// CRC'yi datadan ayır
	receivedData := data[:8]
	receivedCRC := binary.LittleEndian.Uint32(data[8:12])

	// CRC hesapla
	crc32q := crc32.MakeTable(crc32.IEEE) // CRC-32 IEEE polinomu
	calculatedCRC := crc32.Checksum(receivedData, crc32q)

	return receivedCRC == calculatedCRC
}

// restart.go dosyasını çalıştıran fonksiyon
func runRestartScript() error {
	cmd := exec.Command("go", "run", "/path/to/restart.go") // restart.go dosyasının yolunu güncelleyin
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}
