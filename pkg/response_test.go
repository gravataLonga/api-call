package pkg

import (
	"testing"
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
