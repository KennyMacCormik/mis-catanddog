package handlers

import "testing"

func TestGetDocTypeValidateUrl(t *testing.T) {
	var query = []map[string][]string{
		{
			"id": {"2"},
		},
		{
			"id": {"2", "3"},
		},
		{
			"id": {""},
		},
		{
			"doc": {"2"},
		},
		{
			"doc": {"2", "3"},
		},
		{
			"doc": {""},
		},
		{
			"id":  {"2"},
			"doc": {"2"},
		},
	}
	for _, val := range query {
		err := getDocTypeValidateUrl(val)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}
