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
const targetID = "123"

// Hata limiti
const maxErrors = 5

// CAN arayüzünün durumunu kontrol et
func checkCANStatus() (bool, error) {
	cmd := exec.Command("ip", "link", "show", "can0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("CAN arayüzü durumu kontrol edilemedi: %v, output: %s", err, string(output))
	}
	return strings.Contains(string(output), "state UP"), nil
}

// CAN arayüzünü başlat
func startCAN() error {
	cmd := exec.Command("sudo", "ip", "link", "set", "can0", "up", "type", "can", "bitrate", "500000")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CAN arayüzü başlatılamadı: %v, output: %s", err, string(output))
	}
	cmd = exec.Command("sudo", "ifconfig", "can0", "up")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CAN arayüzü etkinleştirilemedi: %v, output: %s", err, string(output))
	}
	return nil
}

// CAN arayüzünün durumunu kontrol et ve yeniden başlat
func restartCAN() error {
	if err := stopCAN(); err != nil {
		return err
	}
	time.Sleep(2 * time.Second) // Kısa bir bekleme süresi ekleyin
	if err := startCAN(); err != nil {
		return err
	}
	return nil
}

// CAN arayüzünü kapat
func stopCAN() error {
	cmd := exec.Command("sudo", "ifconfig", "can0", "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CAN arayüzü kapatılamadı: %v, output: %s", err, string(output))
	}
	return nil
}

func main() {
	errorCount := 0

	// CAN arayüzünün durumu kontrol et
	canStatus, err := checkCANStatus()
	if err != nil {
		log.Fatalf("CAN arayüzü durumu kontrol edilemedi: %v", err)
	}

	if canStatus {
		log.Println("CAN arayüzü zaten açık, işlem yapılmadı.")
	} else {
		log.Println("CAN arayüzü kapalı, açılıyor...")
		if err := startCAN(); err != nil {
			log.Fatalf("Başlangıçta CAN arayüzü başlatılamadı: %v", err)
		}
		log.Println("CAN arayüzü başarıyla başlatıldı.")
	}

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
		fmt.Println("Received Message:", line) // Debugging için satırı yazdır

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
		dataField := strings.Join(fields[5:], "") // Veri alanı için tüm hex verileri birleştir

		// ID ve veri alanlarını ayrıştır
		if idField == targetID {
			data, err := hex.DecodeString(dataField)
			if err != nil {
				log.Printf("Geçersiz veri: %s", dataField)
				errorCount++
				continue
			}

			// Veri uzunluğu 8 byte olduğunda CRC ekleyin ve doğrulama yapın
			if len(data) == 8 {
				dataWithCRC := append(data, calculateCRC(data)...)
				if validateMessage(dataWithCRC) {
					fmt.Println("Mesaj doğrulandı:", data)
					errorCount = 0
				} else {
					fmt.Println("Mesaj doğrulama hatası:", data)
					errorCount++
				}
			}
		} else {
			fmt.Println("ID eşleşmedi:", idField) // Debugging için yazdır
			errorCount++
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

// CRC hesaplama fonksiyonu
func calculateCRC(data []byte) []byte {
	crc32q := crc32.MakeTable(crc32.IEEE) // CRC-32 IEEE polinomu
	crc := crc32.Checksum(data, crc32q)
	crcBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcBytes, crc)
	return crcBytes
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
