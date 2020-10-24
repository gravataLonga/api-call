package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConvertErrorToString(t *testing.T) {
	i := Items{
		ItemsMeta{
			{

				"x011",
				"Testing",
			},
			{

				"x012",
				"Testing 2",
			},
		},
	}

	if i.String() != "[x011]: Testing, [x012]: Testing 2" {
		t.Errorf("Unable to convert to String items, got: %v", i.String())
	}
}

func TestGenerateOperationId(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	a := AuditInfo{ClientIP: "127.0.0.1", Host: "localhost", Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, location)}
	b := BaseStandard{AuditInfo: a}
	op, _ := b.NewOperationId()
	assert.Equal(t, "0b00fff8ca0e86cb772c7ef037c6713d", op)
}
