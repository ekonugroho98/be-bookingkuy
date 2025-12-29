# Admin Service API Documentation

## Overview

The Admin Service provides complete administrative functionality for managing the Bookingkuy platform, including user management, booking oversight, provider configuration, analytics, and system settings.

---

## Authentication

### Admin Login

**Endpoint:** `POST /api/v1/admin/login`

**Request Body:**
```json
{
  "username": "admin@bookingkuy.com",
  "password": "Admin@123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "admin": {
    "id": "uuid-here",
    "email": "admin@bookingkuy.com",
    "role": "super_admin",
    "first_name": "Super",
    "last_name": "Admin",
    "is_active": true,
    "created_at": "2025-01-15T10:00:00Z"
  }
}
```

**Default Credentials:**
- Email: `admin@bookingkuy.com`
- Password: `Admin@123` ⚠️ **Change immediately after first login!**

---

## Dashboard

### Get Dashboard Statistics

**Endpoint:** `GET /api/v1/admin/dashboard`

**Authorization:** Required (Admin role or higher)

**Response (200 OK):**
```json
{
  "stats": {
    "total_users": 1523,
    "total_bookings": 4521,
    "total_revenue": 125000000,
    "today_bookings": 47,
    "today_revenue": 1500000,
    "today_users": 12,
    "active_providers": 3,
    "pending_bookings": 15,
    "confirmed_bookings": 4200,
    "cancelled_bookings": 306
  }
}
```

---

## User Management

### List All Users

**Endpoint:** `GET /api/v1/admin/users`

**Authorization:** Required (`users:read` permission)

**Query Parameters:**
- `page` (optional): Page number, default: 1
- `per_page` (optional): Items per page, default: 20, max: 100
- `search` (optional): Search by name or email
- `role` (optional): Filter by role

**Example:**
```http
GET /api/v1/admin/users?page=1&per_page=20&search=john
```

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "user-123",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+628123456789",
      "role": "user",
      "is_active": true,
      "created_at": "2025-01-15T10:00:00Z",
      "total_bookings": 5,
      "total_spent": 5000000
    }
  ],
  "total": 1523,
  "page": 1,
  "per_page": 20
}
```

### Get User Details

**Endpoint:** `GET /api/v1/admin/users/{id}`

**Authorization:** Required (`users:read` permission)

**Path Parameters:**
- `id`: User ID

**Response (200 OK):**
```json
{
  "user": {
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+628123456789",
    "role": "user",
    "is_active": true,
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-20T15:30:00Z",
    "bookings": [
      {
        "id": "booking-456",
        "hotel_name": "Grand Hotel Bali",
        "status": "CONFIRMED",
        "total_amount": 1500000,
        "check_in": "2025-02-01T00:00:00Z",
        "check_out": "2025-02-03T00:00:00Z"
      }
    ],
    "total_bookings": 5,
    "total_spent": 5000000
  }
}
```

### Update User

**Endpoint:** `PUT /api/v1/admin/users/{id}`

**Authorization:** Required (`users:write` permission)

**Request Body:**
```json
{
  "name": "John Updated",
  "phone": "+628987654321",
  "is_active": true
}
```

**Response (200 OK):**
```json
{
  "message": "User updated successfully",
  "user": {
    "id": "user-123",
    "name": "John Updated",
    "email": "john@example.com",
    "phone": "+628987654321",
    "is_active": true
  }
}
```

### Delete User

**Endpoint:** `DELETE /api/v1/admin/users/{id}`

**Authorization:** Required (`users:delete` permission)

**Path Parameters:**
- `id`: User ID

**Response (200 OK):**
```json
{
  "message": "User deleted successfully"
}
```

---

## Booking Management

### List All Bookings

**Endpoint:** `GET /api/v1/admin/bookings`

**Authorization:** Required (`bookings:read` permission)

**Query Parameters:**
- `page` (optional): Page number, default: 1
- `per_page` (optional): Items per page, default: 20
- `status` (optional): Filter by status (INIT, AWAITING_PAYMENT, PAID, CONFIRMED, COMPLETED, CANCELLED)
- `user_id` (optional): Filter by user ID
- `hotel_id` (optional): Filter by hotel ID
- `start_date` (optional): Filter bookings from this date (YYYY-MM-DD)
- `end_date` (optional): Filter bookings until this date (YYYY-MM-DD)

**Example:**
```http
GET /api/v1/admin/bookings?status=CONFIRMED&page=1&per_page=20
```

**Response (200 OK):**
```json
{
  "bookings": [
    {
      "id": "booking-123",
      "booking_reference": "BKG-ABC123",
      "user_id": "user-456",
      "user_name": "John Doe",
      "user_email": "john@example.com",
      "hotel_id": "hotel-789",
      "hotel_name": "Grand Hotel Bali",
      "check_in": "2025-02-01T00:00:00Z",
      "check_out": "2025-02-03T00:00:00Z",
      "guests": 2,
      "status": "CONFIRMED",
      "total_amount": 1500000,
      "currency": "IDR",
      "created_at": "2025-01-15T10:00:00Z"
    }
  ],
  "total": 4521,
  "page": 1,
  "per_page": 20
}
```

### Get Booking Details

**Endpoint:** `GET /api/v1/admin/bookings/{id}`

**Authorization:** Required (`bookings:read` permission)

**Response (200 OK):**
```json
{
  "booking": {
    "id": "booking-123",
    "booking_reference": "BKG-ABC123",
    "supplier_reference": "HB-123456",
    "user": {
      "id": "user-456",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+628123456789"
    },
    "hotel": {
      "id": "hotel-789",
      "name": "Grand Hotel Bali",
      "city": "Bali",
      "country": "Indonesia"
    },
    "check_in": "2025-02-01T00:00:00Z",
    "check_out": "2025-02-03T00:00:00Z",
    "guests": 2,
    "status": "CONFIRMED",
    "total_amount": 1500000,
    "currency": "IDR",
    "payment": {
      "id": "payment-789",
      "status": "PAID",
      "payment_method": "GOPAY",
      "amount": 1500000
    },
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

### Update Booking

**Endpoint:** `PUT /api/v1/admin/bookings/{id}`

**Authorization:** Required (`bookings:write` permission)

**Request Body:**
```json
{
  "status": "CONFIRMED",
  "notes": "Manually confirmed by support"
}
```

**Response (200 OK):**
```json
{
  "message": "Booking updated successfully",
  "booking": {
    "id": "booking-123",
    "status": "CONFIRMED"
  }
}
```

### Get Booking Statistics

**Endpoint:** `GET /api/v1/admin/bookings/stats`

**Authorization:** Required (`bookings:read` permission)

**Query Parameters:**
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)

**Response (200 OK):**
```json
{
  "stats": {
    "total_bookings": 1523,
    "by_status": {
      "INIT": 15,
      "AWAITING_PAYMENT": 47,
      "PAID": 23,
      "CONFIRMED": 1200,
      "COMPLETED": 200,
      "CANCELLED": 38
    },
    "cancellation_rate": 0.025,
    "total_revenue": 225000000,
    "avg_booking_value": 147700
  }
}
```

---

## Provider Management

### List Providers

**Endpoint:** `GET /api/v1/admin/providers`

**Authorization:** Required (`providers:read` permission)

**Response (200 OK):**
```json
{
  "providers": [
    {
      "code": "HOTELBEDS",
      "name": "HotelBeds",
      "is_active": true,
      "is_healthy": true,
      "metrics": {
        "total_calls": 15234,
        "successful_calls": 15012,
        "failed_calls": 222,
        "success_rate": 0.985,
        "avg_response_time_ms": 234
      }
    },
    {
      "code": "EXPEDIA",
      "name": "Expedia",
      "is_active": false,
      "is_healthy": false,
      "metrics": {
        "total_calls": 0,
        "successful_calls": 0,
        "failed_calls": 0,
        "success_rate": 0,
        "avg_response_time_ms": 0
      }
    }
  ]
}
```

### Get Provider Details

**Endpoint:** `GET /api/v1/admin/providers/{code}`

**Authorization:** Required (`providers:read` permission)

**Path Parameters:**
- `code`: Provider code (HOTELBEDS, EXPEDIA, etc.)

**Response (200 OK):**
```json
{
  "provider": {
    "code": "HOTELBEDS",
    "name": "HotelBeds",
    "base_url": "https://api.test.hotelbeds.com",
    "is_active": true,
    "is_healthy": true,
    "config": {
      "rate_limit_rpm": 100,
      "timeout_seconds": 30,
      "retry_attempts": 3
    },
    "metrics": {
      "today": {
        "total_calls": 523,
        "successful_calls": 518,
        "failed_calls": 5,
        "success_rate": 0.990,
        "avg_response_time_ms": 198
      },
      "this_week": {
        "total_calls": 3521,
        "successful_calls": 3478,
        "failed_calls": 43,
        "success_rate": 0.988,
        "avg_response_time_ms": 212
      }
    }
  }
}
```

### Update Provider

**Endpoint:** `PUT /api/v1/admin/providers/{code}`

**Authorization:** Required (`providers:write` permission)

**Request Body:**
```json
{
  "is_active": true,
  "config": {
    "rate_limit_rpm": 150,
    "timeout_seconds": 45
  }
}
```

**Response (200 OK):**
```json
{
  "message": "Provider updated successfully",
  "provider": {
    "code": "HOTELBEDS",
    "is_active": true
  }
}
```

---

## Analytics

### Get Revenue Statistics

**Endpoint:** `GET /api/v1/admin/analytics/revenue`

**Authorization:** Required (`analytics:read` permission)

**Query Parameters:**
- `start_date` (required): Start date (YYYY-MM-DD)
- `end_date` (required): End date (YYYY-MM-DD)
- `group_by` (optional): Group by `day`, `week`, `month`, `provider`, `payment_method`

**Example:**
```http
GET /api/v1/admin/analytics/revenue?start_date=2025-01-01&end_date=2025-01-31&group_by=day
```

**Response (200 OK):**
```json
{
  "period": {
    "start_date": "2025-01-01",
    "end_date": "2025-01-31"
  },
  "summary": {
    "total_revenue": 325000000,
    "total_bookings": 1523,
    "avg_revenue_per_booking": 213456,
    "growth_vs_previous_period": 0.15
  },
  "breakdown": [
    {
      "date": "2025-01-01",
      "revenue": 10500000,
      "bookings": 52
    },
    {
      "date": "2025-01-02",
      "revenue": 11200000,
      "bookings": 55
    }
  ]
}
```

### Get User Statistics

**Endpoint:** `GET /api/v1/admin/analytics/users`

**Authorization:** Required (`analytics:read` permission)

**Query Parameters:**
- `start_date` (required): Start date
- `end_date` (required): End date

**Response (200 OK):**
```json
{
  "period": {
    "start_date": "2025-01-01",
    "end_date": "2025-01-31"
  },
  "summary": {
    "new_users": 152,
    "active_users": 892,
    "total_users": 2534,
    "growth_rate": 0.08
  },
  "daily_breakdown": [
    {
      "date": "2025-01-01",
      "new_users": 12,
      "active_users": 785
    }
  ]
}
```

### Get Provider Statistics

**Endpoint:** `GET /api/v1/admin/analytics/providers`

**Authorization:** Required (`analytics:read` permission)

**Query Parameters:**
- `start_date` (required): Start date
- `end_date` (required): End date

**Response (200 OK):**
```json
{
  "period": {
    "start_date": "2025-01-01",
    "end_date": "2025-01-31"
  },
  "providers": [
    {
      "code": "HOTELBEDS",
      "total_calls": 15234,
      "successful_calls": 15012,
      "failed_calls": 222,
      "success_rate": 0.985,
      "avg_response_time_ms": 234,
      "total_revenue": 287000000
    }
  ]
}
```

---

## Audit Logs

### Get Audit Logs

**Endpoint:** `GET /api/v1/admin/audit-logs`

**Authorization:** Required (Super Admin only)

**Query Parameters:**
- `admin_id` (optional): Filter by admin ID
- `entity_type` (optional): Filter by entity type (user, booking, provider, config)
- `entity_id` (optional): Filter by entity ID
- `page` (optional): Page number
- `per_page` (optional): Items per page

**Example:**
```http
GET /api/v1/admin/audit-logs?entity_type=booking&page=1&per_page=50
```

**Response (200 OK):**
```json
{
  "logs": [
    {
      "id": "log-123",
      "admin_id": "admin-456",
      "admin_email": "admin@bookingkuy.com",
      "action": "booking.updated",
      "entity_type": "booking",
      "entity_id": "booking-789",
      "old_values": {
        "status": "PAID"
      },
      "new_values": {
        "status": "CONFIRMED"
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 15234,
  "page": 1,
  "per_page": 50
}
```

---

## Admin Management

### List All Admins

**Endpoint:** `GET /api/v1/admin/admins`

**Authorization:** Required (`admins:read` permission - Super Admin only)

**Query Parameters:**
- `page` (optional): Page number
- `per_page` (optional): Items per page
- `role` (optional): Filter by role

**Response (200 OK):**
```json
{
  "admins": [
    {
      "id": "admin-001",
      "email": "admin@bookingkuy.com",
      "role": "super_admin",
      "first_name": "Super",
      "last_name": "Admin",
      "is_active": true,
      "last_login_at": "2025-01-28T15:30:00Z",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "per_page": 20
}
```

### Create Admin

**Endpoint:** `POST /api/v1/admin/admins`

**Authorization:** Required (`admins:write` permission - Super Admin only)

**Request Body:**
```json
{
  "email": "newadmin@bookingkuy.com",
  "password": "SecurePass123!",
  "first_name": "New",
  "last_name": "Admin",
  "role": "admin"
}
```

**Response (201 Created):**
```json
{
  "message": "Admin created successfully",
  "admin": {
    "id": "admin-002",
    "email": "newadmin@bookingkuy.com",
    "role": "admin",
    "first_name": "New",
    "last_name": "Admin",
    "is_active": true
  }
}
```

### Update Admin

**Endpoint:** `PUT /api/v1/admin/admins/{id}`

**Authorization:** Required (`admins:write` permission - Super Admin only)

**Request Body:**
```json
{
  "first_name": "Updated",
  "role": "support",
  "is_active": true
}
```

**Response (200 OK):**
```json
{
  "message": "Admin updated successfully",
  "admin": {
    "id": "admin-002",
    "first_name": "Updated",
    "role": "support"
  }
}
```

### Delete Admin

**Endpoint:** `DELETE /api/v1/admin/admins/{id}`

**Authorization:** Required (`admins:write` permission - Super Admin only)

**Response (200 OK):**
```json
{
  "message": "Admin deleted successfully"
}
```

---

## Role-Based Access Control (RBAC)

### Roles and Permissions

#### Super Admin
- Full access to all resources
- Can manage other admins
- Can modify system configuration

#### Admin
- All user and booking operations
- Provider management (read-only)
- Analytics access
- Cannot manage other admins

#### Moderator
- Booking management (read/write)
- Review moderation
- User operations (read-only)

#### Support
- Read-only access to users, bookings, reviews
- No write permissions

### Permission Check Middleware

All protected endpoints automatically check permissions based on the admin's role. The middleware returns `403 Forbidden` if the admin lacks the required permission.

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
| 401  | Unauthorized - Invalid or missing admin token |
| 403  | Forbidden - Insufficient permissions |
| 404  | Not Found - Resource not found |
| 500  | Internal Server Error - Server error |

---

## Testing with cURL

### Admin Login
```bash
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin@bookingkuy.com",
    "password": "Admin@123"
  }'
```

### Get Dashboard (with token)
```bash
curl -X GET http://localhost:8080/api/v1/admin/dashboard \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### List Users
```bash
curl -X GET "http://localhost:8080/api/v1/admin/users?page=1&per_page=20" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Update Booking Status
```bash
curl -X PUT http://localhost:8080/api/v1/admin/bookings/booking-123 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "CONFIRMED",
    "notes": "Manually confirmed by admin"
  }'
```

---

## Security Best Practices

1. **Change Default Password**
   - First admin login: immediately change password
   - Use strong passwords (min 8 characters, mixed case, numbers)

2. **Use HTTPS Only**
   - Never expose admin endpoints over HTTP
   - Use valid SSL certificates in production

3. **Limit Admin Access**
   - Only grant necessary permissions
   - Use role-based access control
   - Regularly review admin user list

4. **Audit Trail**
   - All admin actions are logged
   - Review audit logs regularly
   - Monitor for suspicious activity

5. **Token Management**
   - Admin tokens expire after 24 hours
   - Store tokens securely (httpOnly cookies)
   - Implement logout functionality

---

## Rate Limiting

Admin endpoints have rate limiting:
- Default: 60 requests per minute
- Burst: 10 requests per second
- Per-admin rate limiting

Configure via environment variables:
```bash
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST=10
```

---

**Last Updated:** 2025-12-28
**Version:** 1.0.0
