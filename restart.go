package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// CAN arayüzünü kapat
func stopCAN() error {
	cmd := exec.Command("sudo", "ifconfig", "can0", "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CAN arayüzü kapatılamadı: %v, output: %s", err, string(output))
	}
	return nil
}

// CAN arayüzünü aç
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

func main() {
	// CAN arayüzünü başlat
	if err := startCAN(); err != nil {
		log.Fatalf("Başlangıçta CAN arayüzü başlatılamadı: %v", err)
	}

	// CAN arayüzünü kullanarak mesaj alma döngüsü
	for {
		// Bu örnekte sadece bir süre bekleyip yeniden başlatma işlemini simüle ediyoruz
		time.Sleep(10 * time.Second)
		log.Println("CAN arayüzü yeniden başlatılıyor...")
		if err := restartCAN(); err != nil {
			log.Printf("CAN arayüzü yeniden başlatılamadı: %v", err)
		} else {
			log.Println("CAN arayüzü başarıyla yeniden başlatıldı.")
		}

		// Aynı dizindeki main.go dosyasını çalıştır
		if err := runMainGo(); err != nil {
			log.Printf("main.go dosyası çalıştırılamadı: %v", err)
		} else {
			log.Println("main.go dosyası başarıyla çalıştırıldı.")
		}
	}
}

// main.go dosyasını çalıştıran fonksiyon
func runMainGo() error {
	cmd := exec.Command("go", "run", "main.go")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}
