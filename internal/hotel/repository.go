package hotel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines hotel repository interface
type Repository interface {
	GetByID(ctx context.Context, id string) (*Hotel, error)
	GetImages(ctx context.Context, hotelID string) ([]Image, error)
	GetRooms(ctx context.Context, hotelID string) ([]RoomInfo, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new hotel repository
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Hotel, error) {
	query := `
		SELECT id, name, description, country_code, city, address,
		       overall_rating, star_rating, latitude, longitude,
		       created_at, updated_at
		FROM hotels
		WHERE id = $1 AND deleted_at IS NULL
	`

	var hotel Hotel
	var starRating float64
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&hotel.ID, &hotel.Name, &hotel.Description,
		&hotel.CountryCode, &hotel.City, &hotel.Address,
		&hotel.Rating, &starRating,
		&hotel.Location.Latitude, &hotel.Location.Longitude,
		&hotel.CreatedAt, &hotel.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get hotel: %w", err)
	}

	// Set category from star_rating
	hotel.Category = fmt.Sprintf("%.0f Star", starRating)

	// Get amenities
	amenitiesQuery := `SELECT amenity FROM hotel_amenities WHERE hotel_id = $1`
	rows, _ := r.pool.Query(ctx, amenitiesQuery, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var amenity string
			if rows.Scan(&amenity) == nil {
				hotel.Amenities = append(hotel.Amenities, amenity)
			}
		}
	}

	return &hotel, nil
}

func (r *repository) GetImages(ctx context.Context, hotelID string) ([]Image, error) {
	query := `
		SELECT id, hotel_id, url, type, caption, sort_order
		FROM hotel_images
		WHERE hotel_id = $1
		ORDER BY sort_order ASC
	`

	rows, err := r.pool.Query(ctx, query, hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hotel images: %w", err)
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.ID, &img.HotelID, &img.URL, &img.Type, &img.Caption, &img.SortOrder); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *repository) GetRooms(ctx context.Context, hotelID string) ([]RoomInfo, error) {
	query := `
		SELECT id, hotel_id, name, max_guests, beds, size
		FROM rooms
		WHERE hotel_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	defer rows.Close()

	var rooms []RoomInfo
	for rows.Next() {
		var room RoomInfo
		if err := rows.Scan(&room.ID, &room.HotelID, &room.Name, &room.MaxGuests, &room.Beds, &room.Size); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}
