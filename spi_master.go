package main

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

func main() {
	// Periph.io'yu başlat
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// SPI portunu aç
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// SPI bağlantı parametrelerini ayarla
	conn, err := p.Connect(1e6, spi.Mode0, 8)
	if err != nil {
		log.Fatal(err)
	}

	// Gönderilecek ve alınacak veriler
	tx := []byte{42} // Arduino'ya gönderilecek veri
	rx := make([]byte, len(tx))

	// SPI transferi
	if err := conn.Tx(tx, rx); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Gönderilen veri: %v\n", tx)
	fmt.Printf("Alınan veri: %v\n", rx)
}
