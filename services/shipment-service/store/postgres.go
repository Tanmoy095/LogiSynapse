package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Tanmoy095/LogiSynapse/shared/contracts"
	"github.com/Tanmoy095/LogiSynapse/shared/proto"
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

// Takes a context for cancellation/timeout and a Shipment struct with data
// CreateShipment inserts a new shipment into the database.
// Why: Persists shipment data, including dynamic dimensions, for real-world accuracy.
func (s *PostgresStore) CreateShipment(ctx context.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
	// Define the SQL query to insert a shipment and return the generated ID
	// SQL query to insert shipment and return generated ID
	// Why: Stores all fields, including package details and tracking
	query := `
		INSERT INTO shipments (origin, destination, status, eta, carrier_name, carrier_tracking_url, tracking_number, length, width, height, weight, unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	// Execute the query with the shipment data and scan the returned ID into shipment.ID
	// Execute query with shipment data
	// Why: Saves data and retrieves UUID
	// Convert proto enum to string for DB storage
	statusStr := shipment.Status.String()
	err := s.db.QueryRowContext(ctx, query,
		shipment.Origin,              // Shipment origin (e.g., "New York")
		shipment.Destination,         // Shipment destination (e.g., "London")
		statusStr,                    // Shipment status as string
		shipment.Eta,                 // Estimated time of arrival (nullable)
		shipment.Carrier.Name,        // Carrier name (e.g., "FedEx")
		shipment.Carrier.TrackingURL, // Carrier tracking URL (nullable)
		shipment.TrackingNumber,
		shipment.Length,
		shipment.Width,
		shipment.Height,
		shipment.Weight,
		shipment.Unit,
	).Scan(&shipment.ID)

	// Check for errors during the query execution
	if err != nil {
		// Return an empty Shipment and an error if the insert fails
		return contracts.Shipment{}, fmt.Errorf("failed to insert shipment: %v", err)
	}

	// Return the shipment with the newly assigned ID
	return shipment, nil
}

// CreateShipmentWithOutbox inserts a new shipment into the database and publishes an outbox event
// This function is used to create a shipment and publish an outbox event to the event store
// The outbox event is used to track the shipment and publish it to the event store
func (s *PostgresStore) CreateShipmentWithOutbox(ctx context.Context, shipment contracts.Shipment, eventKey string, eventPayload []byte) (contracts.Shipment, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return contracts.Shipment{}, fmt.Errorf("failed to start tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	insertShipment := `
		INSERT INTO shipments (origin, destination, status, eta, carrier_name, carrier_tracking_url, tracking_number, length, width, height, weight, unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`
	statusStr := shipment.Status.String()
	if err = tx.QueryRowContext(ctx, insertShipment,
		shipment.Origin,
		shipment.Destination,
		statusStr,
		shipment.Eta,
		shipment.Carrier.Name,
		shipment.Carrier.TrackingURL,
		shipment.TrackingNumber,
		shipment.Length,
		shipment.Width,
		shipment.Height,
		shipment.Weight,
		shipment.Unit,
	).Scan(&shipment.ID); err != nil {
		return contracts.Shipment{}, fmt.Errorf("failed to insert shipment in tx: %w", err)
	}

	insertOutbox := `
		INSERT INTO shipment_outbox (aggregate_id, event_type, event_key, payload)
		VALUES ($1, $2, $3, $4)`
	if _, err = tx.ExecContext(ctx, insertOutbox, shipment.ID, "shipment.created", eventKey, eventPayload); err != nil {
		return contracts.Shipment{}, fmt.Errorf("failed to insert outbox event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return contracts.Shipment{}, fmt.Errorf("failed to commit tx: %w", err)
	}
	return shipment, nil
}

// GetShipment retrieves a shipment by ID.
// Why: Needed for UpdateShipment and DeleteShipment to check status and preserve data.
func (s *PostgresStore) GetShipment(ctx context.Context, id string) (contracts.Shipment, error) {
	// SQL query to fetch shipment with all fields
	// Why: Retrieves complete data, including dimensions
	query := `
		SELECT id, origin, destination, status, eta, carrier_name, carrier_tracking_url, tracking_number,
   		length, width, height, weight, unit
		FROM shipments WHERE id = $1`
	var shipment contracts.Shipment
	// Use sql.Null* for nullable fields
	// Why: Handles nullable database fields safely
	var statusStr, eta, carrierName, trackingURL, trackingNumber, unit sql.NullString
	var length, width, height, weight sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&shipment.ID, &shipment.Origin, &shipment.Destination, &statusStr,
		&eta, &carrierName, &trackingURL, &trackingNumber,
		&length, &width, &height, &weight, &unit,
	)
	// Handle not found error
	if err == sql.ErrNoRows {
		return contracts.Shipment{}, fmt.Errorf("shipment not found")
	}
	if err != nil {
		return contracts.Shipment{}, err
	}
	// Assign nullable fields
	// Why: Converts database nulls to Go zero values
	shipment.Eta = eta.String
	shipment.Carrier = contracts.Carrier{Name: carrierName.String, TrackingURL: trackingURL.String}
	shipment.TrackingNumber = trackingNumber.String
	shipment.Length = length.Float64
	shipment.Width = width.Float64
	shipment.Height = height.Float64
	shipment.Weight = weight.Float64
	shipment.Unit = unit.String
	// parse status string into proto enum
	shipment.Status = parseStatusStringToProto(statusStr.String)
	return shipment, nil
}

// GetShipments retrieves shipments from the database with optional filtering and pagination
// Filters by origin, status, and destination (empty string means no filter)
// Uses limit and offset for pagination
func (s *PostgresStore) GetShipments(ctx context.Context, origin string, status proto.ShipmentStatus, destination string, limit, offset int32) ([]contracts.Shipment, error) {
	// Define the SQL query to select shipments with filters and pagination
	//sql querry with filter and pagination
	query := `
        SELECT id, origin, destination, status, eta, carrier_name, carrier_tracking_url,
		tracking_number,length,width,height, weight, unit 
        FROM shipments
        WHERE ($1 = '' OR origin = $1)
          AND ($2 = '' OR status = $2)
          AND ($3 = '' OR destination = $3)
        ORDER BY eta ASC
        LIMIT $4 OFFSET $5`

	// Execute the query with the provided filters and pagination parameters
	// convert status proto enum to string for DB query
	statusStr := status.String()
	rows, err := s.db.QueryContext(ctx, query, origin, statusStr, destination, limit, offset)
	if err != nil {
		// Return an error if the query fails
		return nil, err
	}
	// Ensure rows are closed to free resources
	defer rows.Close()

	// Initialize a slice to store the retrieved shipments
	var shipments []contracts.Shipment

	// Iterate over the query results
	for rows.Next() {
		// Create a new Shipment struct for each row
		var sh contracts.Shipment
		// Use sql.NullString for nullable fields (eta, carrier_name, carrier_tracking_url)
		var statusStr, eta, carrierName, trackingURL, trackingNumber, unit sql.NullString
		var length, width, height, weight sql.NullFloat64

		// Scan the row data into the Shipment struct and nullable fields
		if err := rows.Scan(
			&sh.ID,          // Shipment ID (UUID)
			&sh.Origin,      // Shipment origin
			&sh.Destination, // Shipment destination
			&statusStr,      // Shipment status (string)
			&eta,            // Nullable ETA
			&carrierName,    // Nullable carrier name
			&trackingURL,    // Nullable carrier tracking URL
			&trackingNumber,
			&length,
			&width,
			&height,
			&weight,
			&unit,
		); err != nil {
			// Return an error if scanning fails
			return nil, err
		}

		// Assign nullable fields to the Shipment struct, using empty string if null
		sh.Eta = eta.String
		sh.Carrier = contracts.Carrier{
			Name:        carrierName.String,
			TrackingURL: trackingURL.String,
		}
		sh.TrackingNumber = trackingNumber.String
		sh.Length = length.Float64
		sh.Width = width.Float64
		sh.Height = height.Float64
		sh.Weight = weight.Float64
		sh.Unit = unit.String
		sh.Status = parseStatusStringToProto(statusStr.String)

		// Append the shipment to the results slice
		shipments = append(shipments, sh)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return the list of shipments
	return shipments, nil
}

//Update Shipment updates a shipment in a database
//Update details destination like dimensions and status

// Sql query to Update all fields
// Persists changes including status for deleteShipment
func (s *PostgresStore) UpdateShipment(ctx context.Context, shipment contracts.Shipment) error {
	//sql query to update all fields
	query := `
UPDATE shipments
SET origin = $1, destination = $2, status = $3, eta = $4,carrier_name = $5, carrier_tracking_url = $6, tracking_number =    $7,length = $8, width = $9, height = $10, weight = $11, unit = $12 
WHERE id = $13`
	//Execute update
	//Save updated shipment data
	// convert enum to string for DB
	statusStr := shipment.Status.String()
	_, err := s.db.ExecContext(ctx, query,
		shipment.Origin, shipment.Destination, statusStr, shipment.Eta,
		shipment.Carrier.Name, shipment.Carrier.TrackingURL, shipment.TrackingNumber,
		shipment.Length, shipment.Width, shipment.Height, shipment.Weight, shipment.Unit,
		shipment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update shipment: %v", err)
	}
	return nil

}

// PopPendingOutboxEvent fetches the oldest pending outbox event for a given aggregate ID.
// It returns the event ID, payload, and an error if the event is not found.
func (s *PostgresStore) PopPendingOutboxEvent(ctx context.Context, aggregateID string) (string, []byte, error) {
	query := `
		SELECT id, payload
		FROM shipment_outbox
		WHERE aggregate_id = $1 AND published_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1`

	var eventID string
	var payload []byte
	err := s.db.QueryRowContext(ctx, query, aggregateID).Scan(&eventID, &payload)
	if err == sql.ErrNoRows {
		return "", nil, nil
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch pending outbox event: %w", err)
	}
	return eventID, payload, nil
}

func (s *PostgresStore) MarkOutboxEventPublished(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	query := `UPDATE shipment_outbox SET published_at = NOW() WHERE id = $1`
	if _, err := s.db.ExecContext(ctx, query, eventID); err != nil {
		return fmt.Errorf("failed to mark outbox event published: %w", err)
	}
	return nil
}

// parseStatusStringToProto converts status string (stored in DB or from Shippo)
// into the proto.ShipmentStatus enum. Unknown values map to PENDING.
func parseStatusStringToProto(status string) proto.ShipmentStatus {
	switch status {
	case "PRE_TRANSIT":
		return proto.ShipmentStatus_PRE_TRANSIT
	case "IN_TRANSIT":
		return proto.ShipmentStatus_IN_TRANSIT
	case "DELIVERED":
		return proto.ShipmentStatus_DELIVERED
	case "PENDING":
		return proto.ShipmentStatus_PENDING
	case "CANCELLED":
		return proto.ShipmentStatus_CANCELLED
	default:
		return proto.ShipmentStatus_PENDING
	}
}
