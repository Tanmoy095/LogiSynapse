package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Tanmoy095/LogiSynapse/shipment-service/internal/models"
	_ "github.com/lib/pq"
)

// PostgresStore manages database operations for the Shipment Service
type PostgresStore struct {
	db *sql.DB // Holds the database connection
}

// NewPostgresStore creates a new PostgresStore instance with a database connection
// connStr is the PostgreSQL connection string (e.g., postgres://user:pass@host:port/dbname)
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	// Open a connection to the PostgreSQL database using the provided connection string
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		// Return an error if the connection cannot be opened
		return nil, fmt.Errorf("failed to open postgres db: %v", err)
	}

	// Test the database connection to ensure it's valid
	if err := db.Ping(); err != nil {
		// Return an error if the database cannot be reached
		return nil, fmt.Errorf("failed to ping postgres db: %v", err)
	}

	// Return a new PostgresStore instance with the database connection
	return &PostgresStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	// Close the database connection and return any error
	return s.db.Close()
}

// CreateShipment inserts a new shipment into the database
// Takes a context for cancellation/timeout and a Shipment struct with data
func (s *PostgresStore) CreateShipment(ctx context.Context, shipment models.Shipment) (models.Shipment, error) {
	// Define the SQL query to insert a shipment and return the generated ID
	query := `
        INSERT INTO shipments (origin, destination, status, eta, carrier_name, carrier_tracking_url)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

	// Execute the query with the shipment data and scan the returned ID into shipment.ID
	err := s.db.QueryRowContext(ctx, query,
		shipment.Origin,              // Shipment origin (e.g., "New York")
		shipment.Destination,         // Shipment destination (e.g., "London")
		shipment.Status,              // Shipment status (e.g., "In Transit")
		shipment.ETA,                 // Estimated time of arrival (nullable)
		shipment.Carrier.Name,        // Carrier name (e.g., "FedEx")
		shipment.Carrier.TrackingURL, // Carrier tracking URL (nullable)
	).Scan(&shipment.ID)

	// Check for errors during the query execution
	if err != nil {
		// Return an empty Shipment and an error if the insert fails
		return models.Shipment{}, fmt.Errorf("failed to insert shipment: %v", err)
	}

	// Return the shipment with the newly assigned ID
	return shipment, nil
}

// GetShipments retrieves shipments from the database with optional filtering and pagination
// Filters by origin, status, and destination (empty string means no filter)
// Uses limit and offset for pagination
func (s *PostgresStore) GetShipments(ctx context.Context, origin, status, destination string, limit, offset int32) ([]models.Shipment, error) {
	// Define the SQL query to select shipments with filters and pagination
	query := `
        SELECT id, origin, destination, status, eta, carrier_name, carrier_tracking_url
        FROM shipments
        WHERE ($1 = '' OR origin = $1)
          AND ($2 = '' OR status = $2)
          AND ($3 = '' OR destination = $3)
        ORDER BY eta ASC
        LIMIT $4 OFFSET $5`

	// Execute the query with the provided filters and pagination parameters
	rows, err := s.db.QueryContext(ctx, query, origin, status, destination, limit, offset)
	if err != nil {
		// Return an error if the query fails
		return nil, err
	}
	// Ensure rows are closed to free resources
	defer rows.Close()

	// Initialize a slice to store the retrieved shipments
	var shipments []models.Shipment

	// Iterate over the query results
	for rows.Next() {
		// Create a new Shipment struct for each row
		var s models.Shipment
		// Use sql.NullString for nullable fields (eta, carrier_name, carrier_tracking_url)
		var eta, carrierName, trackingURL sql.NullString

		// Scan the row data into the Shipment struct and nullable fields
		if err := rows.Scan(
			&s.ID,          // Shipment ID (UUID)
			&s.Origin,      // Shipment origin
			&s.Destination, // Shipment destination
			&s.Status,      // Shipment status
			&eta,           // Nullable ETA
			&carrierName,   // Nullable carrier name
			&trackingURL,   // Nullable carrier tracking URL
		); err != nil {
			// Return an error if scanning fails
			return nil, err
		}

		// Assign nullable fields to the Shipment struct, using empty string if null
		s.ETA = eta.String
		s.Carrier = models.Carrier{
			Name:        carrierName.String,
			TrackingURL: trackingURL.String,
		}

		// Append the shipment to the results slice
		shipments = append(shipments, s)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return the list of shipments
	return shipments, nil
}
