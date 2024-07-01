package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// CAN arayüzünün durumunu kontrol et
func checkCANStatus() (bool, error) {
	cmd := exec.Command("ip", "link", "show", "can0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("CAN arayüzü durumu kontrol edilemedi: %v, output: %s", err, string(output))
	}
	return strings.Contains(string(output), "state UP"), nil
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

	// CAN arayüzünü kullanarak mesaj alma döngüsü
	for {
		// Bu örnekte sadece bir süre bekleyip yeniden başlatma işlemini simüle ediyoruz
		time.Sleep(0 * time.Second)
		log.Println("CAN arayüzü yeniden başlatılıyor...")
		if err := restartCAN(); err != nil {
			log.Printf("CAN arayüzü yeniden başlatılamadı: %v", err)
		} else {
			log.Println("CAN arayüzü başarıyla yeniden başlatıldı.")
		}
	}
}
