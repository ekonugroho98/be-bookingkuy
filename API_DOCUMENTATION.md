# Bookingkuy API Documentation

**Base URL**: `http://localhost:8080` (Development)
**API Version**: v1
**Content-Type**: `application/json`

## Table of Contents

- [Authentication](#authentication)
- [Search Hotels](#search-hotels)
- [Hotel Details](#hotel-details)
- [Bookings](#bookings)
- [Payments](#payments)
- [User Profile](#user-profile)
- [Health Check](#health-check)

---

## Authentication

Sebagian besar endpoint memerlukan authentication menggunakan JWT token. Include token di header:

```
Authorization: Bearer <your-jwt-token>
```

### Register New User

**POST** `/api/v1/auth/register`

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "phone": "+628123456789"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550b8b66-746f-4c65-8e0b-7fcb7c0c20d5",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+628123456789",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

### Login

**POST** `/api/v1/auth/login`

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550b8b66-746f-4c65-8e0b-7fcb7c0c20d5",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+628123456789",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

---

## Search Hotels

### Autocomplete Search Suggestions

**GET** `/api/v1/search/autocomplete?q={query}&limit={limit}`

Endpoint untuk mendapatkan sugesti pencarian saat user mengetik. Mencari city dan hotel names secara bersamaan.

**Query Parameters:**
- `q` (string, required) - Query search (minimal 2 karakter)
- `limit` (integer, optional) - Jumlah hasil (default: 10, max: 20)

**Example:** `/api/v1/search/autocomplete?q=bal&limit=10`

**Response (200 OK):**
```json
{
  "query": "bal",
  "results": [
    {
      "type": "city",
      "id": "Bali",
      "name": "Bali",
      "country": "ID"
    },
    {
      "type": "hotel",
      "id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
      "name": "Bali Paradise Hotel",
      "city": "Bali",
      "country": "ID"
    },
    {
      "type": "hotel",
      "id": "0fc05305-9c90-412b-bf49-dc9b46a4138b",
      "name": "Kuta Beach Resort",
      "city": "Bali",
      "country": "ID"
    }
  ]
}
```

**Field Penjelasan:**
- `type`: Jenis result ("city" atau "hotel")
- `id`: Untuk city, ini adalah nama city yang bisa dipakai langsung di search. Untuk hotel, ini adalah hotel_id
- `name`: Nama untuk ditampilkan di UI
- `city` & `country`: Lokasi (opsional untuk city type)

### Get Popular Destinations

**GET** `/api/v1/search/destinations?limit={limit}`

Endpoint untuk mendapatkan daftar destinasi populer (default saat search box kosong).

**Query Parameters:**
- `limit` (integer, optional) - Jumlah hasil (default: 10, max: 20)

**Example:** `/api/v1/search/destinations?limit=10`

**Response (200 OK):**
```json
{
  "destinations": [
    {
      "type": "city",
      "id": "Bali",
      "name": "Bali",
      "country": "ID"
    },
    {
      "type": "city",
      "id": "Jakarta",
      "name": "Jakarta",
      "country": "ID"
    },
    {
      "type": "city",
      "id": "Surabaya",
      "name": "Surabaya",
      "country": "ID"
    }
  ]
}
```

### Search Hotels by City

**POST** `/api/v1/search/hotels`

**Request Body:**
```json
{
  "check_in": "2025-01-15T00:00:00Z",
  "check_out": "2025-01-17T00:00:00Z",
  "city": "Bali",
  "guests": 2,
  "min_price": 50,
  "max_price": 500
}
```

**Response (200 OK):**
```json
{
  "hotels": [
    {
      "id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
      "name": "Bali Paradise Hotel",
      "country_code": "ID",
      "city": "Bali",
      "rating": 4.5
    }
  ],
  "total": 25,
  "page": 1,
  "per_page": 20,
  "total_pages": 2
}
```

---

## Hotel Details

### Get Hotel Details

**GET** `/api/v1/hotels/{id}`

**URL Parameters:**
- `id` (string) - Hotel ID

**Response (200 OK):**
```json
{
  "id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
  "name": "Bali Paradise Hotel",
  "description": "Luxurious beachfront resort",
  "country_code": "ID",
  "city": "Bali",
  "address": "Jalan Beach 123",
  "rating": 4.5,
  "category": "4 Star",
  "images": [
    {
      "id": "img-001",
      "hotel_id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
      "url": "https://example.com/image.jpg",
      "type": "exterior",
      "caption": "Hotel exterior view",
      "sort_order": 1
    }
  ],
  "amenities": ["WiFi", "Pool", "Spa"],
  "location": {
    "latitude": -8.409518,
    "longitude": 115.188919
  },
  "policies": {
    "check_in_time": "14:00",
    "check_out_time": "12:00",
    "cancellation": "Free cancellation until 24h before check-in"
  },
  "rooms": []
}
```

### Get Available Rooms

**GET** `/api/v1/hotels/{id}/rooms`

**URL Parameters:**
- `id` (string) - Hotel ID
- `check_in` (string) - Check-in date (YYYY-MM-DD format)
- `check_out` (string) - Check-out date (YYYY-MM-DD format)

**Example:** `/api/v1/hotels/60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe/rooms?check_in=2025-01-15&check_out=2025-01-17`

**Response (200 OK):**
```json
{
  "hotel_id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
  "hotel_name": "Bali Paradise Hotel",
  "check_in": "2025-01-15",
  "check_out": "2025-01-17",
  "guests": 2,
  "rooms": [
    {
      "room_id": "room-001",
      "room_name": "Deluxe Ocean View",
      "available": true,
      "price": 150000,
      "currency": "IDR",
      "max_guests": 2,
      "beds": "1 King Bed"
    }
  ]
}
```

### Get Hotel Images

**GET** `/api/v1/hotels/{id}/images`

**URL Parameters:**
- `id` (string) - Hotel ID

**Response (200 OK):**
```json
{
  "images": [
    {
      "id": "img-001",
      "hotel_id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
      "url": "https://example.com/image.jpg",
      "type": "exterior",
      "caption": "Hotel exterior view",
      "sort_order": 1
    }
  ]
}
```

---

## Bookings

Semua endpoint booking memerlukan authentication.

### Create Booking

**POST** `/api/v1/bookings`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Request Body:**
```json
{
  "hotel_id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
  "room_id": "room-001",
  "check_in": "2025-01-15T00:00:00Z",
  "check_out": "2025-01-17T00:00:00Z",
  "guests": 2,
  "payment_type": "PAY_NOW"
}
```

**Payment Type Options:**
- `PAY_NOW` - Pay immediately via payment gateway
- `PAY_AT_HOTEL` - Pay when you arrive at the hotel

**Response (201 Created):**
```json
{
  "id": "bk-001",
  "user_id": "user-001",
  "hotel_id": "60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe",
  "hotelName": "Bali Paradise Hotel",
  "hotelImage": "https://example.com/hotel.jpg",
  "city": "Bali",
  "room_id": "room-001",
  "roomName": "Deluxe Ocean View",
  "booking_reference": "BK20250115001",
  "checkIn": "Jan 15, 2025",
  "checkOut": "Jan 17, 2025",
  "guests": 2,
  "guestsFormatted": "2 Adults",
  "totalPrice": 300000,
  "total_amount": 300000,
  "currency": "IDR",
  "status": "Confirmed",
  "payment_type": "PAY_NOW",
  "created_at": "2025-01-10T10:30:00Z"
}
```

### Get Booking Details

**GET** `/api/v1/bookings/{id}`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**URL Parameters:**
- `id` (string) - Booking ID

**Response (200 OK):** Same as Create Booking response

### Get My Bookings

**GET** `/api/v1/bookings/my`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Query Parameters:**
- `page` (integer, optional) - Page number, default: 1
- `per_page` (integer, optional) - Items per page, default: 20

**Example:** `/api/v1/bookings/my?page=1&per_page=20`

**Response (200 OK):**
```json
{
  "bookings": [
    // Array of booking objects
  ],
  "page": 1,
  "per_page": 20
}
```

### Cancel Booking

**POST** `/api/v1/bookings/{id}/cancel`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**URL Parameters:**
- `id` (string) - Booking ID

**Response (200 OK):**
```json
{
  "message": "Booking cancelled successfully",
  "booking": {
    // Booking object with updated status
  }
}
```

---

## Payments

### Create Payment

**POST** `/api/v1/payments`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Request Body:**
```json
{
  "booking_id": "bk-001",
  "amount": 300000
}
```

**Response (201 Created):**
```json
{
  "id": "pay-001",
  "booking_id": "bk-001",
  "amount": 300000,
  "currency": "IDR",
  "status": "Pending",
  "payment_type": "MIDTRANS",
  "payment_url": "https://app.midtrans.com/payment-link/xxx",
  "midtrans_tx_id": "midtrans-001",
  "created_at": "2025-01-10T10:30:00Z",
  "updated_at": "2025-01-10T10:35:00Z"
}
```

### Get Payment Details

**GET** `/api/v1/payments/{id}`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**URL Parameters:**
- `id` (string) - Payment ID

**Response (200 OK):** Same as Create Payment response

---

## User Profile

### Get User Profile

**GET** `/api/v1/users/me`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):**
```json
{
  "id": "550b8b66-746f-4c65-8e0b-7fcb7c0c20d5",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+628123456789",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Update User Profile

**PUT** `/api/v1/users/me`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Request Body:**
```json
{
  "name": "John Updated",
  "phone": "+628123456789"
}
```

**Response (200 OK):**
```json
{
  "id": "550b8b66-746f-4c65-8e0b-7fcb7c0c20d5",
  "name": "John Updated",
  "email": "john@example.com",
  "phone": "+628123456789",
  "created_at": "2025-01-15T10:30:00Z"
}
```

---

## Health Check

### Check API Health

**GET** `/health`

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-10T10:30:00Z"
}
```

### Check Readiness

**GET** `/health/ready`

**Response (200 OK):**
```json
{
  "status": "ready",
  "database": "up",
  "redis": "up"
}
```

### Check Liveness

**GET** `/health/live`

**Response (200 OK):**
```json
{
  "status": "alive"
}
```

---

## Error Responses

Semua error responses mengikuti format ini:

```json
{
  "error": "Error message description"
}
```

**Common HTTP Status Codes:**
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required or invalid token
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists (e.g., email already registered)
- `500 Internal Server Error` - Server error

---

## Swagger UI

Interactive API documentation tersedia di:
**`http://localhost:8080/swagger/index.html`**

Anda bisa mencoba API langsung dari Swagger UI dengan fitur "Try it out".

---

## Notes

1. **Date Format**: Gunakan ISO 8601 format untuk tanggal (contoh: `2025-01-15T00:00:00Z`)
2. **Authentication Token**: Setelah login, simpan token dan sertakan di semua request yang memerlukan authentication
3. **Currency**: Semua harga dalam IDR (Indonesian Rupiah)
4. **Pagination**: Untuk endpoint yang mendukung pagination, gunakan parameter `page` dan `per_page`
5. **Booking Status**: Status booking bisa berupa:
   - `Confirmed` - Booking confirmed
   - `Pending` - Waiting for payment
   - `Cancelled` - Booking cancelled

---

## Support

Untuk pertanyaan atau issues, hubungi:
- Email: support@bookingkuy.com
- Documentation: https://docs.bookingkuy.com
