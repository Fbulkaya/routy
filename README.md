# Routy API

Routy, başlangıç ve varış noktaları arasında kullanıcı ilgi alanlarına göre önerilen durakları (kafeler, parklar, manzaralar vb.) bulan bir rota öneri API'sidir.

## 🚀 Özellikler

- Başlangıç ve varış noktası girerek rota oluşturma
- Kullanıcının mevcut konumuna göre öneri alma
- İlgi alanlarına göre filtreleme (örn: "cafe", "park", "viewpoint")
- Redis ile önbellekleme (performans artışı için)
- OpenStreetMap ve Overpass API kullanımı

## 📦 Kurulum

### Gereksinimler

- Go 1.20+  
- Docker (opsiyonel ama önerilir)
- Redis

### 1. Kaynak kodu klonlayın

```bash
git clone https://github.com/Fbulkaya/routy.git
cd routy

#Docker ile çalıştırın
docker build -t routy-api .
docker run -p 8080:8080 routy-api

#Alternatif olarak doğrudan Go ile çalıştırabilirsiniz:

go run main.go


# API Kullanımı

# 1. Rota oluşturma

curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "start": "Taksim, Istanbul",
    "end": "Kadikoy, Istanbul",
    "interests": ["cafe", "park"]
  }'

# 2. Mevcut konuma göre öneri alma
curl -X POST http://localhost:8080/route_with_current \
  -H "Content-Type: application/json" \
  -d '{
    "start": "Taksim, Istanbul",
    "end": "Kadikoy, Istanbul",
    "current": {
      "lat": 41.0369,
      "lon": 28.9861
    },
    "interests": ["cafe", "viewpoint"]
  }'

 #  📁 Proje Yapısı

 .
├── main.go
├── Dockerfile
├── .dockerignore
├── go.mod
├── models/
├── utils/
├── handlers/

📜 Lisans

MIT © [Fbulkaya]

🤝 Katkı

Pull request’ler ve öneriler memnuniyetle karşılanır! Katkı sunmak için önce bir issue açabilirsin.

🌍 API Ne İşe Yarar?

Bu API, örneğin bir mobil uygulamada yolculuk sırasında ilginç yerler keşfetmek isteyen kullanıcılar için idealdir. Turistler, gezginler veya sürücüler için rota üstü öneriler sağlar.


