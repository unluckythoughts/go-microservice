package bus

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Message struct {
	ID           string
	CorelationID string
	RoutingKeys  []string
	Type         string
	PublishTime  time.Time
	Body         []byte
}

// From sets the message fields from the given body and other parameters
func (m *Message) From(body any, corelationID string, msgType string) error {
	m.ID = uuid.Must(uuid.NewV4()).String()
	m.CorelationID = corelationID
	m.Type = msgType
	m.PublishTime = time.Now()
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("could not marshal message body: %w", err)
	}
	m.Body = bodyBytes
	return nil
}
