# Ticket #007: Frontend-Backend Data Contract Fixes

**Created:** 2025-12-28
**Status:** ⚠️ **IN PROGRESS**
**Priority:** 🔴 **CRITICAL** - Blocker for frontend integration
**Estimated Time:** 2-3 hours

---

## 📋 Problem Statement

Frontend expects different data structure than backend provides. Critical mismatches found that will break booking flow.

### Critical Issues Found:
1. Field naming: `total_amount` vs `totalPrice`, `check_in` vs `checkIn`
2. Missing fields: `hotelName`, `hotelImage`, `city`, `roomName`
3. Status values: "CONFIRMED" vs "Confirmed"
4. Type mismatches: guests (int vs string), dates (ISO8601 vs formatted)
5. Mock checkout implementation in FE (no API calls)

---

## 🎯 Deliverables

### Backend Changes
- [x] Create `BookingResponse` DTO with FE-compatible fields
- [ ] Update `CreateBooking` handler to return FE-compatible response
- [ ] Update `GetBooking` handler to return FE-compatible response
- [ ] Update `ListUserBookings` handler to return FE-compatible response
- [ ] Add hotel/room data fetching in booking handlers
- [ ] Update API documentation with new response format

### Testing
- [ ] Update booking integration test to validate new response format
- [ ] Add test for FE-compatible response structure
- [ ] Verify all booking endpoints return correct format

### Documentation
- [ ] Update API.md with new response examples
- [ ] Document field mapping between BE and FE
- [ ] Update FRONTEND_INTEGRATION.md with contract details

---

## 📝 Implementation Notes

**Files to Modify:**
- `internal/booking/handler.go` - Update handlers
- `internal/booking/service.go` - Add hotel/room fetching
- `internal/booking/booking_integration_test.go` - Update tests
- `docs/API.md` - Update documentation

**Files Created:**
- `internal/booking/dto.go` - Response DTOs (✅ DONE)
- `docs/FE_BE_DATA_CONTRACT_ANALYSIS.md` - Analysis report (✅ DONE)

---

**Last Updated:** 2025-12-28
