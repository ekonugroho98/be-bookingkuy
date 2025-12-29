# Frontend - Backend Data Contract Analysis

**Date:** 2025-12-28
**Status:** ⚠️ **CRITICAL INCONSISTENCIES FOUND**
**Risk:** **HIGH** - Data contract mismatches between FE and BE

---

## 🔴 Critical Issues Found

### Issue 1: Booking Field Name Mismatch

**Frontend Expectation** (`fe-bookingkuy/types.ts:47`):
```typescript
export interface Booking {
  id: string;
  hotelId: string;        // ✅ camelCase
  hotelName: string;      // ✅ Expected
  hotelImage: string;     // ✅ Expected
  city: string;           // ✅ Expected
  checkIn: string;        // ✅ camelCase
  checkOut: string;       // ✅ camelCase
  guests: string;         // ❌ Type mismatch (string vs number)
  totalPrice: number;     // ❌ WRONG FIELD NAME
  status: 'Confirmed' | 'Pending' | 'Cancelled';  // ❌ Value mismatch
  roomName: string;       // ✅ Expected
}
```

**Backend Response** (`be-bookingkuy/internal/booking/model.go:31-51`):
```go
type Booking struct {
  ID                string        `json:"id"`                    // ✅
  UserID            string        `json:"user_id"`               // ❌ snake_case
  HotelID           string        `json:"hotel_id"`              // ❌ snake_case (FE expects hotelId)
  RoomID            string        `json:"room_id"`               // ❌ snake_case (FE expects roomName)
  BookingReference  string        `json:"booking_reference"`     // ❌ snake_case
  SupplierReference string        `json:"supplier_reference"`    // ❌ snake_case
  CheckIn           time.Time     `json:"check_in"`              // ❌ snake_case (FE expects checkIn)
  CheckOut          time.Time     `json:"check_out"`             // ❌ snake_case (FE expects checkOut)
  Guests            int           `json:"guests"`                // ❌ Type: number (FE expects string)
  GuestName         string        `json:"guest_name"`            // ❌ snake_case
  GuestEmail        string        `json:"guest_email"`           // ❌ snake_case
  GuestPhone        string        `json:"guest_phone"`           // ❌ snake_case
  SpecialRequests   string        `json:"special_requests"`      // ❌ snake_case
  Status            BookingStatus `json:"status"`                // ⚠️ Value mismatch
  TotalAmount       int           `json:"total_amount"`          // ❌ WRONG! FE expects totalPrice
  Currency          string        `json:"currency"`              // ✅
  PaymentType       PaymentType   `json:"payment_type"`          // ❌ snake_case
  CreatedAt         time.Time     `json:"created_at"`            // ❌ snake_case
  UpdatedAt         time.Time     `json:"updated_at"`            // ❌ snake_case
}
```

**Impact:** 🔴 **CRITICAL** - Frontend will break when displaying bookings

---

### Issue 2: Booking Status Value Mismatch

**Frontend Expectation**:
```typescript
status: 'Confirmed' | 'Pending' | 'Cancelled'
```

**Backend Values**:
```go
const (
  StatusInit          BookingStatus = "INIT"                  // ❌ Not in FE
  StatusAwaitingPayment BookingStatus = "AWAITING_PAYMENT"   // ❌ Not in FE
  StatusPaid          BookingStatus = "PAID"                 // ❌ Not in FE
  StatusConfirmed     BookingStatus = "CONFIRMED"            // ✅ Matches
  StatusCompleted     BookingStatus = "COMPLETED"            // ❌ Not in FE
  StatusCancelled     BookingStatus = "CANCELLED"            // ✅ Matches
)
```

**Impact:** 🔴 **CRITICAL** - Status display will fail, wrong UI states

---

### Issue 3: Missing Fields in Backend Response

**Frontend Expects but Backend Doesn't Provide:**
```typescript
hotelName: string;      // ❌ Backend only sends hotel_id
hotelImage: string;     // ❌ Backend only sends hotel_id
city: string;           // ❌ Backend only sends hotel_id
roomName: string;       // ❌ Backend only sends room_id
```

**Backend Only Sends IDs:**
```json
{
  "hotel_id": "hotel-123",   // ❌ FE needs hotel name & image
  "room_id": "room-123"      // ❌ FE needs room name
}
```

**Impact:** 🔴 **CRITICAL** - FE cannot display booking details without joining hotel/room data

---

### Issue 4: Date Format Mismatch

**Frontend Expects**:
```typescript
checkIn: string;    // "Oct 24, 2024" (formatted string)
checkOut: string;   // "Oct 25, 2024" (formatted string)
```

**Backend Sends**:
```go
CheckIn  time.Time `json:"check_in"`   // "2025-01-15T00:00:00Z"
CheckOut time.Time `json:"check_out"`  // "2025-01-17T00:00:00Z"
```

**Impact:** 🟡 **MEDIUM** - FE needs to parse ISO 8601 dates

---

### Issue 5: Guests Field Type Mismatch

**Frontend Expects**:
```typescript
guests: string;  // "2 Adults"
```

**Backend Sends**:
```go
Guests int `json:"guests"`  // 2 (number)
```

**Impact:** 🟡 **MEDIUM** - FE needs to format number to string

---

### Issue 6: Missing API Integration in FE

**Frontend Checkout** (`fe-bookingkuy/pages/Checkout.tsx:10-14`):
```typescript
const handleConfirm = () => {
  // Generate a mock booking ID
  const bookingId = `BK-${Math.floor(Math.random() * 900000) + 100000}`;
  navigate('/confirmation', { state: { hotel, room, bookingId } });
};
```

**Problems:**
- ❌ No API call to backend
- ❌ Mock booking ID generation
- ❌ No payment processing
- ❌ No booking creation

**Impact:** 🔴 **CRITICAL** - Checkout flow is completely broken

---

## 📊 Complete Field Mapping Analysis

### Booking Object

| FE Field | FE Type | BE Field | BE Type | Match? | Issue |
|----------|---------|----------|---------|--------|-------|
| id | string | id | string | ✅ | - |
| hotelId | string | hotel_id | string | ⚠️ | Naming: camelCase vs snake_case |
| hotelName | string | - | - | ❌ | **MISSING** |
| hotelImage | string | - | - | ❌ | **MISSING** |
| city | string | - | - | ❌ | **MISSING** |
| checkIn | string | check_in | time.Time | ⚠️ | Format: string vs ISO8601 |
| checkOut | string | check_out | time.Time | ⚠️ | Format: string vs ISO8601 |
| guests | string | guests | int | ❌ | **Type mismatch** |
| totalPrice | number | total_amount | int | ❌ | **Wrong field name** |
| status | enum | status | BookingStatus | ❌ | **Value mismatch** |
| roomName | string | - | - | ❌ | **MISSING** |

---

### Hotel Object

**Frontend** (`fe-bookingkuy/types.ts:2-21`):
```typescript
export interface Hotel {
  id: string;                  // ✅
  name: string;                // ✅
  location: string;            // ✅
  city: string;                // ✅
  country: string;             // ✅
  rating: number;              // ✅
  reviewCount: number;         // ❓ Backend doesn't have this
  pricePerNight: number;       // ❓ Backend doesn't provide this directly
  originalPrice?: number;      // ❓ Backend doesn't have this
  image: string;               // ✅ (from hotelbeds)
  amenities: string[];         // ✅ (from hotelbeds)
  description: string;         // ✅ (from hotelbeds)
  isPopular?: boolean;         // ❓ Backend doesn't have this
  isBestValue?: boolean;       // ❓ Backend doesn't have this
  distanceFromBeach?: string;  // ❓ Backend doesn't have this
  distanceFromCenter?: string; // ❓ Backend doesn't have this
  aiReasoning?: string;        // ❓ Backend doesn't have this
  matchPercentage?: number;    // ❓ Backend doesn't have this
}
```

**Backend** (from HotelBeds integration):
- Most core fields match (id, name, location, image, etc.)
- Missing: AI recommendation fields, distance fields, popularity flags

---

## 🎯 Summary of Inconsistencies

### Critical (Must Fix Before Integration) 🔴

1. **`total_amount` vs `totalPrice`** - Field name mismatch
2. **`hotel_id` vs `hotelName`** - BE sends ID, FE needs name
3. **`room_id` vs `roomName`** - BE sends ID, FE needs name
4. **Status values** - "CONFIRMED" vs "Confirmed" (capitalization)
5. **No API calls in FE checkout** - Mock implementation

### High Priority (Should Fix) 🟠

6. **Date format** - BE sends ISO8601, FE expects formatted strings
7. **Guests type** - BE sends number, FE expects string
8. **Missing hotel/room details** in booking response

### Medium Priority (Nice to Have) 🟡

9. **Snake_case vs camelCase** - Inconsistent naming convention
10. **Missing metadata** - reviewCount, distances, AI fields

---

## 💡 Recommended Solutions

### Option 1: Fix Backend to Match Frontend (Recommended) ✅

**Changes Required:**

1. **Update Booking Response** - Add computed fields:
```go
type BookingResponse struct {
  // Existing fields
  ID                string        `json:"id"`
  HotelID           string        `json:"hotel_id"`
  RoomID            string        `json:"room_id"`

  // NEW: Add joined fields
  HotelName         string        `json:"hotelName"`
  HotelImage        string        `json:"hotelImage"`
  City              string        `json:"city"`
  RoomName          string        `json:"roomName"`

  // Format dates as strings
  CheckIn           string        `json:"checkIn"`
  CheckOut          string        `json:"checkOut"`

  // Add both formats for compatibility
  TotalAmount       int           `json:"total_amount"`
  TotalPrice        int           `json:"totalPrice"` // Alias

  // Format guests as string
  Guests            int           `json:"guests"`
  GuestsFormatted   string        `json:"guestsFormatted"`

  Status            BookingStatus `json:"status"`

  // ... other fields
}
```

2. **Create Response Builder**:
```go
func BuildBookingResponse(booking *booking.Booking, hotel *hotelbeds.Hotel, room *hotelbeds.Room) BookingResponse {
    return BookingResponse{
        ID:                booking.ID,
        HotelID:           booking.HotelID,
        HotelName:         hotel.Name,
        HotelImage:        hotel.Image,
        City:              hotel.City,
        RoomID:            booking.RoomID,
        RoomName:          room.Name,
        CheckIn:           booking.CheckIn.Format("Jan 2, 2006"),
        CheckOut:          booking.CheckOut.Format("Jan 2, 2006"),
        Guests:            booking.Guests,
        GuestsFormatted:   fmt.Sprintf("%d Adults", booking.Guests),
        TotalAmount:       booking.TotalAmount,
        TotalPrice:        booking.TotalAmount,
        Status:            booking.Status,
        // ... map other fields
    }
}
```

3. **Update Booking Handler**:
```go
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    // ... create booking

    // Fetch hotel and room details
    hotel, _ := h.hotelbedsClient.GetHotel(ctx, booking.HotelID)
    room, _ := h.hotelbedsClient.GetRoom(ctx, booking.RoomID)

    // Build response with joined data
    response := BuildBookingResponse(booking, hotel, room)

    respondWithJSON(w, http.StatusCreated, response)
}
```

### Option 2: Fix Frontend to Match Backend (Not Recommended) ❌

**Why Not:**
- Requires extensive FE changes
- Breaks existing UI components
- More complex date parsing on client
- Need to fetch hotel/room details separately

---

## 📋 Implementation Checklist

### Backend Changes Required

- [ ] Create `BookingResponse` struct with FE-compatible fields
- [ ] Add `hotelName`, `hotelImage`, `city`, `roomName` fields
- [ ] Format dates as readable strings (add `checkIn`, `checkOut`)
- [ ] Add `totalPrice` alias for `total_amount`
- [ ] Add `guestsFormatted` for display
- [ ] Update booking handlers to join hotel/room data
- [ ] Update GET /api/v1/bookings/:id endpoint
- [ ] Update GET /api/v1/users/:id/bookings endpoint
- [ ] Add API response examples to documentation

### Frontend Changes Required

- [ ] Implement actual API call in checkout flow
- [ ] Add payment processing integration
- [ ] Handle ISO8601 date parsing (or use formatted strings from BE)
- [ ] Update Booking interface to match backend response
- [ ] Add error handling for booking creation
- [ ] Add loading states during booking creation
- [ ] Update MyBookings page to fetch from API
- [ ] Update Confirmation page to show real booking data

---

## 🚀 Risk Assessment

**Current Risk Level:** 🔴 **HIGH**

**Why:**
1. Checkout flow is completely broken (mock implementation)
2. Booking display will fail due to field mismatches
3. Status values will break conditional rendering
4. Missing hotel/room details prevents proper display

**Impact:**
- Users cannot complete bookings
- Booking history page will be empty or broken
- Confirmation page shows incorrect data
- Payment processing is not integrated

**Mitigation:**
- ✅ Implement Backend changes first (Option 1)
- ✅ Add comprehensive integration tests
- ✅ Document API contracts clearly
- ✅ Test end-to-end booking flow

---

## 📝 Conclusion

**Status:** ⚠️ **NOT READY FOR FRONTEND INTEGRATION**

**Critical Blockers:**
1. Backend response doesn't match frontend expectations
2. Missing hotel/room data in booking response
3. No actual API integration in FE checkout
4. Status value mismatches

**Recommendation:**
✅ **FIX BACKEND FIRST** - Update booking responses to match frontend expectations
✅ **Then integrate FE** - Connect checkout to real API endpoints

**Estimated Effort:**
- Backend changes: 2-3 hours
- Frontend integration: 3-4 hours
- Testing & debugging: 2-3 hours
- **Total: 7-10 hours**

---

**Created:** 2025-12-28
**Status:** 🔴 CRITICAL INCONSISTENCIES FOUND
**Priority:** **HIGH** - Must fix before frontend integration
