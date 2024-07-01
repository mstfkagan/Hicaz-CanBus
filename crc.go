package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"log"
	"os/exec"
	"strings"
)

// Belirli bir ID'ye sahip mesajları filtrelemek için
const targetID = "223"

// Hata limiti
const maxErrors = 5

func main() {
	errorCount := 0

	// candump komutunu çalıştır
	cmd := exec.Command("candump", "can0")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("candump komutu çalıştırılamadı: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("candump komutu başlatılamadı: %v", err)
	}

	// candump çıktısını okuma
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Received line:", line) // Debugging için satırı yazdır

		// CAN mesajını ayrıştır
		fields := strings.Fields(line)
		if len(fields) < 5 {
			log.Printf("Geçersiz CAN mesajı: %s", line)
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

		idField := fields[1] // ID alanı için indeks 1 olarak güncellendi
		dataFields := fields[5:] // Veri alanı için indeks 5'ten sonrasını alın

		// Veri alanını birleştir
		dataField := strings.Join(dataFields, "")

		// ID ve veri alanlarını ayrıştır
		if idField == targetID {
			data, err := hex.DecodeString(dataField)
			if err != nil {
				log.Printf("Geçersiz veri: %s", dataField)
				errorCount++
				continue
			}

			if validateMessage(data) {
				fmt.Println("Mesaj doğrulandı:", data)
				errorCount = 0
			} else {
				fmt.Println("Mesaj doğrulama hatası:", data)
				errorCount++
			}
		} else {
			fmt.Println("ID eşleşmedi:", idField) // Debugging için yazdır
		}

		if errorCount >= maxErrors {
			log.Println("Hata limiti aşıldı, restart.go dosyası çalıştırılıyor...")
			if err := runRestartScript(); err != nil {
				log.Fatalf("restart.go dosyası çalıştırılamadı: %v", err)
			}
			errorCount = 0
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("candump çıktısı okunurken hata: %v", err)
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
