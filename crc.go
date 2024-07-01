package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"os/exec"
	"time"

	"github.com/brutella/can"
	"github.com/brutella/can/pkg/adapter/socketcan"
)

// Belirli bir ID'ye sahip mesajları filtrelemek için
const targetID = 0x123

// Hata limiti
const maxErrors = 5

func main() {
	// CAN arayüzünü aç
	config := socketcan.NewConfig()
	config.Name = "can0"
	config.Bitrate = 500000

	adapter, err := socketcan.New(config)
	if err != nil {
		log.Fatalf("CAN arayüzü açılırken hata: %v", err)
	}

	bus := can.NewBus(adapter)
	defer bus.Disconnect()

	errorCount := 0

	// Mesaj alma ve doğrulama
	bus.SubscribeFunc(func(frm can.Frame) {
		if frm.ID == targetID {
			if validateMessage(frm.Data) {
				fmt.Println("Mesaj doğrulandı:", frm.Data)
			} else {
				fmt.Println("Mesaj doğrulama hatası:", frm.Data)
			}
		}
	})

	bus.ConnectAndPublish()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Hata sayısını kontrol et
		if errorCount >= maxErrors {
			log.Println("Hata limiti aşıldı, restart.go dosyası çalıştırılıyor...")
			if err := runRestartScript(); err != nil {
				log.Fatalf("restart.go dosyası çalıştırılamadı: %v", err)
			}
			errorCount = 0
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
	cmd := exec.Command("go", "run", "restart.go") // restart.go dosyasının adını belirtin
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}
