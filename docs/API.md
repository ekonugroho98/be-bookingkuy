# Bookingkuy API Documentation

## Base URL
- **Development**: `http://localhost:8080`
- **Production**: `https://api.bookingkuy.com`

## Authentication

Most endpoints require JWT token in Authorization header:

```http
Authorization: Bearer <token>
```

### Getting a Token

Register or login to receive a JWT token:

```bash
# Register
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "phone": "+628123456789"
}

# Login
POST /api/v1/auth/login
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

Response includes token:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

---

## Endpoints

### Auth Endpoints (Public)

#### Register User
**POST** `/api/v1/auth/register`

Register a new user account.

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "phone": "+628123456789"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+628123456789",
    "role": "user"
  }
}
```

**Error Response (409 Conflict):**
```json
{
  "error": "Email already registered"
}
```

---

#### Login
**POST** `/api/v1/auth/login`

Authenticate user and receive JWT token.

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Invalid email or password"
}
```

---

### Hotel Endpoints (Public)

#### Search Hotels
**POST** `/api/v1/search/hotels`

Search for available hotels.

**Query Parameters:**
- `page` (optional): Page number, default: 1
- `per_page` (optional): Items per page, default: 20, max: 100
- `sort_by` (optional): Sort field (price, rating, name)

**Request Body:**
```json
{
  "check_in": "2025-01-15T00:00:00Z",
  "check_out": "2025-01-17T00:00:00Z",
  "city": "Bali",
  "guests": 2
}
```

**Response (200 OK):**
```json
{
  "hotels": [
    {
      "id": "hotel-123",
      "name": "Grand Hotel Bali",
      "description": "Luxury beachfront hotel...",
      "city": "Bali",
      "country": "Indonesia",
      "rating": 4.5,
      "category": "5STAR",
      "price_per_night": 1500000,
      "currency": "IDR",
      "image_url": "https://..."
    }
  ],
  "total": 50,
  "page": 1,
  "per_page": 20
}
```

---

#### Get Hotel Details
**GET** `/api/v1/hotels/{id}`

Get complete hotel information including images, amenities, and policies.

**Path Parameters:**
- `id`: Hotel ID

**Response (200 OK):**
```json
{
  "id": "hotel-123",
  "name": "Grand Hotel Bali",
  "description": "Luxury beachfront hotel with stunning ocean views...",
  "country": "Indonesia",
  "city": "Bali",
  "address": "Jl. Beach No. 123",
  "rating": 4.5,
  "category": "5STAR",
  "images": [
    {
      "id": "img-1",
      "url": "https://...",
      "type": "exterior",
      "caption": "Hotel exterior",
      "sort_order": 1
    }
  ],
  "amenities": ["pool", "wifi", "spa", "restaurant", "gym"],
  "location": {
    "latitude": -8.3405,
    "longitude": 115.0920
  },
  "policies": {
    "check_in_time": "14:00",
    "check_out_time": "12:00",
    "cancellation": "Free cancellation until 24h before check-in"
  },
  "rooms": [
    {
      "id": "room-123",
      "name": "Deluxe Room",
      "max_guests": 2,
      "beds": "1 King Bed"
    }
  ]
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Hotel not found"
}
```

---

#### Get Available Rooms
**GET** `/api/v1/hotels/{hotelId}/rooms`

Check room availability and pricing for specific dates.

**Path Parameters:**
- `hotelId`: Hotel ID

**Query Parameters:**
- `checkIn` (required): Check-in date (YYYY-MM-DD format)
- `checkOut` (required): Check-out date (YYYY-MM-DD format)
- `guests` (required): Number of guests

**Example:**
```http
GET /api/v1/hotels/hotel-123/rooms?checkIn=2025-01-15&checkOut=2025-01-17&guests=2
```

**Response (200 OK):**
```json
{
  "hotel_id": "hotel-123",
  "hotel_name": "Grand Hotel Bali",
  "check_in": "2025-01-15",
  "check_out": "2025-01-17",
  "guests": 2,
  "rooms": [
    {
      "room_id": "room-123",
      "room_name": "Deluxe Room",
      "available": true,
      "price": 1500000,
      "currency": "IDR",
      "max_guests": 2,
      "beds": "1 King Bed"
    }
  ]
}
```

---

#### Get Hotel Images
**GET** `/api/v1/hotels/{id}/images`

Get all hotel images.

**Path Parameters:**
- `id`: Hotel ID

**Response (200 OK):**
```json
{
  "hotel_id": "hotel-123",
  "images": [
    {
      "id": "img-1",
      "url": "https://...",
      "type": "exterior",
      "caption": "Hotel exterior",
      "sort_order": 1
    }
  ]
}
```

---

### Booking Endpoints (Protected)

#### Create Booking
**POST** `/api/v1/bookings`

Create a new booking.

**Headers:**
```http
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "hotel_id": "hotel-123",
  "room_id": "room-123",
  "check_in": "2025-01-15T00:00:00Z",
  "check_out": "2025-01-17T00:00:00Z",
  "guests": 2,
  "payment_type": "PAY_NOW"
}
```

**Response (201 Created) - Frontend Compatible:**
```json
{
  "id": "booking-123",
  "user_id": "user-123",
  "hotel_id": "hotel-123",
  "hotelName": "Grand Hyatt Bali",
  "hotelImage": "https://example.com/hotel.jpg",
  "city": "Bali",
  "room_id": "room-123",
  "roomName": "Deluxe Ocean View",
  "booking_reference": "BKG-ABC123",
  "checkIn": "Jan 15, 2025",
  "checkOut": "Jan 17, 2025",
  "guests": 2,
  "guestsFormatted": "2 Adults",
  "totalPrice": 3000000,
  "total_amount": 3000000,
  "currency": "IDR",
  "status": "Pending",
  "payment_type": "PAY_NOW",
  "created_at": "2025-01-10T10:00:00Z"
}
```

**Note:** Response includes both camelCase (for frontend) and snake_case (for API) field names for compatibility.

**Error Response (400 Bad Request):**
```json
{
  "error": "check-out date must be after check-in date"
}
```

---

#### Get My Bookings
**GET** `/api/v1/bookings/my`

Get current user's bookings with pagination. Returns frontend-compatible format with hotel and room details.

**Headers:**
```http
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (optional): Page number, default: 1
- `per_page` (optional): Items per page, default: 20, max: 100

**Example:**
```http
GET /api/v1/bookings/my?page=1&per_page=20
```

**Response (200 OK) - Frontend Compatible:**
```json
{
  "bookings": [
    {
      "id": "booking-123",
      "user_id": "user-123",
      "hotel_id": "hotel-123",
      "hotelName": "Grand Hyatt Bali",
      "hotelImage": "https://example.com/hotel.jpg",
      "city": "Bali",
      "room_id": "room-123",
      "roomName": "Deluxe Ocean View",
      "booking_reference": "BKG-ABC123",
      "checkIn": "Jan 15, 2025",
      "checkOut": "Jan 17, 2025",
      "guests": 2,
      "guestsFormatted": "2 Adults",
      "totalPrice": 3000000,
      "total_amount": 3000000,
      "currency": "IDR",
      "status": "Confirmed",
      "payment_type": "PAY_NOW",
      "created_at": "2025-01-10T10:00:00Z"
    }
  ],
  "page": 1,
  "per_page": 20
}
```

---

#### Get Booking Details
**GET** `/api/v1/bookings/{id}`

Get details of a specific booking. Returns frontend-compatible format with hotel and room details.

**Headers:**
```http
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Booking ID

**Response (200 OK):**
```json
{
  "id": "booking-123",
  "user_id": "user-123",
  "hotel_id": "hotel-123",
  "room_id": "room-123",
  "booking_reference": "BKG-ABC123",
  "supplier_reference": "HB-123456",
  "check_in": "2025-01-15T00:00:00Z",
  "check_out": "2025-01-17T00:00:00Z",
  "guests": 2,
  "status": "CONFIRMED",
  "total_amount": 3000000,
  "currency": "IDR",
  "payment_type": "PAY_NOW",
  "created_at": "2025-01-10T10:00:00Z",
  "updated_at": "2025-01-10T10:05:00Z"
}
```

---

#### Cancel Booking
**POST** `/api/v1/bookings/{id}/cancel`

Cancel a booking and process refund if applicable.

**Headers:**
```http
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Booking ID

**Response (200 OK):**
```json
{
  "message": "Booking cancelled successfully",
  "booking": {
    "id": "booking-123",
    "status": "CANCELLED"
  }
}
```

---

### Payment Endpoints (Protected + Webhook)

#### Create Payment
**POST** `/api/v1/payments`

Create a payment and get Midtrans payment URL.

**Headers:**
```http
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "booking_id": "booking-123",
  "payment_method_type": "GOPAY"
}
```

**Response (201 Created):**
```json
{
  "id": "payment-123",
  "booking_id": "booking-123",
  "amount": 3000000,
  "status": "PENDING",
  "payment_url": "https://app.sandbox.midtrans.com/payment-link/...",
  "created_at": "2025-01-10T10:00:00Z"
}
```

---

#### Get Payment Status
**GET** `/api/v1/payments/{id}`

Get payment status by ID.

**Headers:**
```http
Authorization: Bearer <token>
```

**Path Parameters:**
- `id`: Payment ID

**Response (200 OK):**
```json
{
  "id": "payment-123",
  "booking_id": "booking-123",
  "amount": 3000000,
  "status": "PAID",
  "payment_method_type": "GOPAY",
  "created_at": "2025-01-10T10:00:00Z",
  "updated_at": "2025-01-10T10:05:00Z"
}
```

---

#### Payment Webhook
**POST** `/api/v1/payments/webhook`

Midtrans webhook endpoint (public, no authentication required).

**Request Body:** (Sent by Midtrans)
```json
{
  "transaction_time": "2025-01-10 10:05:00",
  "transaction_status": "settlement",
  "transaction_id": "abc123",
  "status_message": "Success",
  "signature_key": "...",
  "payment_type": "gopay",
  "order_id": "payment-123",
  "gross_amount": "3000000"
}
```

**Response (200 OK):**
```json
{
  "message": "Webhook processed successfully"
}
```

---

### User Endpoints (Protected)

#### Get Profile
**GET** `/api/v1/users/me`

Get current user's profile.

**Headers:**
```http
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "user-123",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+628123456789",
  "role": "user",
  "created_at": "2025-01-01T00:00:00Z"
}
```

---

#### Update Profile
**PUT** `/api/v1/users/me`

Update current user's profile.

**Headers:**
```http
Authorization: Bearer <token>
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
  "message": "Profile updated successfully",
  "user": {
    "id": "user-123",
    "name": "John Updated",
    "email": "john@example.com",
    "phone": "+628123456789"
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message here"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200  | OK - Request successful |
| 201  | Created - Resource created successfully |
| 400  | Bad Request - Invalid input or validation error |
| 401  | Unauthorized - Invalid or missing token |
| 404  | Not Found - Resource not found |
| 409  | Conflict - Duplicate resource or state conflict |
| 500  | Internal Server Error - Server error |

### Common Error Messages

| Error | Status Code | Description |
|-------|-------------|-------------|
| `Invalid request body` | 400 | Malformed JSON |
| `Booking not found` | 404 | Booking doesn't exist |
| `Invalid email or password` | 401 | Authentication failed |
| `Email already registered` | 409 | User exists |
| `check-out date must be after check-in date` | 400 | Validation error |
| `room is not available` | 400 | No availability |

---

## Data Models

### Booking Status

Possible booking status values:
- `INIT` - Initial state
- `AWAITING_PAYMENT` - Waiting for payment
- `PAID` - Payment received
- `CONFIRMED` - Confirmed with supplier
- `COMPLETED` - Stay completed
- `CANCELLED` - Booking cancelled

### Payment Type

- `PAY_NOW` - Pay in full now
- `PAY_AT_HOTEL` - Pay at hotel (coming soon)

### Payment Methods (Midtrans)

- `GOPAY` - GoPay e-wallet
- `BCA_KLIKPAY` - BCA KlikPay
- `SHOPEEPAY` - ShopeePay
- `DANA` - DANA e-wallet
- `CREDIT_CARD` - Credit/Debit card

---

## Testing with cURL

### Register and Login
```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "phone": "+628123456789"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

### Search Hotels
```bash
curl -X POST http://localhost:8080/api/v1/search/hotels \
  -H "Content-Type: application/json" \
  -d '{
    "check_in": "2025-01-15T00:00:00Z",
    "check_out": "2025-01-17T00:00:00Z",
    "city": "Bali",
    "guests": 2
  }'
```

### Get Hotel Details
```bash
curl -X GET http://localhost:8080/api/v1/hotels/hotel-123
```

### Create Booking (Authenticated)
```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hotel_id": "hotel-123",
    "room_id": "room-123",
    "check_in": "2025-01-15T00:00:00Z",
    "check_out": "2025-01-17T00:00:00Z",
    "guests": 2,
    "payment_type": "PAY_NOW"
  }'
```

---

**Last Updated:** 2025-12-27
