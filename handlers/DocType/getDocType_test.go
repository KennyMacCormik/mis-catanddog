package DocType

import (
	"context"
	"log/slog"
	"mis-catanddog/controllers"
	"net/url"
	"testing"
)

type fakeDB struct {
	docTypeId  map[int]string
	docTypeDoc map[string]int
}

func (f *fakeDB) DocTypeGetById(ctx context.Context, id int, l *slog.Logger) controllers.DocType {
	val, ok := f.docTypeId[id]
	if !ok {
		return controllers.DocType{Id: 0, Doc: "", Err: ""}
	}
	return controllers.DocType{Id: id, Doc: val, Err: ""}
}

func (f *fakeDB) DocTypeGetByDoc(ctx context.Context, doc string, l *slog.Logger) controllers.DocType {
	val, ok := f.docTypeDoc[doc]
	if !ok {
		return controllers.DocType{Id: 0, Doc: "", Err: ""}
	}
	return controllers.DocType{Id: val, Doc: doc, Err: ""}
}

/*
func (f *fakeDB) New(uri string, timeout time.Duration) error {
	return nil
}
func (f *fakeDB) Get(ctx context.Context, r repos.DbReq) (*sql.Rows, error) {
	return nil, nil
}
func (f *fakeDB) Exec(ctx context.Context, rs []repos.DbReq) error {
	return nil
}
func (f *fakeDB) Close() {}
*/

func TestGetDocTypeValidateUrl(t *testing.T) {
	type urlTest struct {
		Url     url.Values
		Err     string
		Message string
		Crit    bool
	}
	var fail bool
	var arr = []urlTest{
		{
			Url: map[string][]string{
				"id": {"1", "2"},
			},
			Err:     "",
			Message: "positive test [id 1 2] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"doc": {"passport", "veterenary pasword"},
			},
			Err:     "",
			Message: "positive test [doc 'passport' 'veterenary pasword'] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"id":  {"1", "2"},
				"doc": {"passport", "veterenary pasword"},
			},
			Err:     "ambiguous query; 'doc' and 'id' either together or not present; query [map[doc:[passport veterenary pasword] id:[1 2]]]",
			Message: "negative test ['doc' and 'id' both preset] failed",
			Crit:    true,
		},
	}

	for _, val := range arr {
		err := getDocTypeValidateUrl(val.Url)
		if err != nil && err.Error() != val.Err {
			if val.Crit {
				fail = true
			}
			t.Logf("crit: %t; %s; %s", val.Crit, val.Message, err.Error())
		}
	}

	if fail {
		t.Fatalf("Critical tests failed")
	}
}

func compare(a, b []controllers.DocType) bool {
	var ln = len(a)
	if ln != len(b) {
		return false
	}

	for i := 0; i < ln; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestGetDocTypeQueryData(t *testing.T) {
	type queryTest struct {
		Url     url.Values
		Result  []controllers.DocType
		Message string
		Crit    bool
	}

	var ctx = context.TODO()
	var log = &slog.Logger{}
	var fail = false
	var db = &fakeDB{
		docTypeId: map[int]string{
			1: "passport",
			2: "veterinary passport",
		},
		docTypeDoc: map[string]int{
			"passport":            1,
			"veterinary passport": 2,
		},
	}
	var arr = []queryTest{
		{
			Url: map[string][]string{
				"id": {"1"},
			},
			Result: []controllers.DocType{
				{
					Id:  1,
					Doc: "passport",
					Err: "",
				},
			},
			Message: "positive test [ Url map[id:[1]] Result [{1 passport }] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"id": {"1", "2"},
			},
			Result: []controllers.DocType{
				{
					Id:  1,
					Doc: "passport",
					Err: "",
				},
				{
					Id:  2,
					Doc: "veterinary passport",
					Err: "",
				},
			},
			Message: "positive test [ Url map[id:[1 2]] Result [{1 passport } {2 veterinary passport }] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"doc": {"passport"},
			},
			Result: []controllers.DocType{
				{
					Id:  1,
					Doc: "passport",
					Err: "",
				},
			},
			Message: "positive test [ Url map[doc:[passport]] Result [{1 passport }] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"doc": {"passport", "veterinary passport"},
			},
			Result: []controllers.DocType{
				{
					Id:  1,
					Doc: "passport",
					Err: "",
				},
				{
					Id:  2,
					Doc: "veterinary passport",
					Err: "",
				},
			},
			Message: "positive test [ Url map[doc:[passport veterinary passport]] Result [{1 passport } {2 veterinary passport }] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"doc": {"document"},
			},
			Result: []controllers.DocType{
				{
					Id:  0,
					Doc: "",
					Err: "",
				},
			},
			Message: "negative test [ Url map[doc:[document]] Result [{0  Empty result}] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"id": {"3"},
			},
			Result: []controllers.DocType{
				{
					Id:  0,
					Doc: "",
					Err: "",
				},
			},
			Message: "negative test [ Url map[id:[3]] Result [{0  Empty result}] ] failed",
			Crit:    true,
		},
		{
			Url: map[string][]string{
				"id": {"fail"},
			},
			Result: []controllers.DocType{
				{
					Id:  0,
					Doc: "",
					Err: "failed to convert id [fail] to an integer1",
				},
			},
			Message: "negative test [ Url map[id:[fail]] Result [{0  failed to convert id [fail] to an integer}] ] failed",
			Crit:    true,
		},
	}

	for _, val := range arr {
		result := getDocTypeQueryData(ctx, val.Url, db, log)
		if !compare(result, val.Result) {
			if val.Crit {
				fail = true
			}
			//t.Logf("%v || %v || %v", result, val.Result, val.Url)
			t.Logf("crit: %t; %s", val.Crit, val.Message)
		}
	}

	if fail {
		t.Fatalf("Critical tests failed")
	}
}

func TestGetDocTypeHideInternals(t *testing.T) {
	var fail = false
	var log = &slog.Logger{}
	var arrMutated = []controllers.DocType{
		{
			Id:  1,
			Doc: "passport",
			Err: "",
		},
		{
			Id:  0,
			Doc: "",
			Err: "",
		},
		{
			Id:  5,
			Doc: "",
			Err: "error",
		},
	}
	var arrInitial = make([]controllers.DocType, len(arrMutated))
	var arrResult = []controllers.DocType{
		{
			Id:  1,
			Doc: "passport",
			Err: "",
		},
		{
			Id:  0,
			Doc: "",
			Err: "Empty result",
		},
		{
			Id:  5,
			Doc: "",
			Err: "Bad request",
		},
	}
	if len(arrMutated) != len(arrResult) && len(arrMutated) != len(arrInitial) {
		t.Fatalf("All three initial arrays must be of the same len")
	}
	copy(arrInitial, arrMutated)
	getDocTypeHideInternals(arrMutated, log)
	for i := 0; i < len(arrInitial); i++ {
		if arrMutated[i] != arrResult[i] {
			fail = true
			t.Logf("input: %v; output: %v; expected: %v", arrInitial[i], arrMutated[i], arrResult[i])
		}
	}

	if fail {
		t.Fatalf("Critical tests failed")
	}
}

func TestGetDocType(t *testing.T) {
	// TODO: test
}
