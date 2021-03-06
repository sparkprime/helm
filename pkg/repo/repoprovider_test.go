/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package repo

import (
	"github.com/kubernetes/helm/pkg/chart"

	"reflect"
	"testing"
)

var (
	TestShortReference = "helm:gs/" + TestRepoBucket + "/" + TestChartName + "#" + TestChartVersion
	TestLongReference  = TestRepoURL + "/" + TestArchiveName
)

var ValidChartReferences = []string{
	TestShortReference,
	TestLongReference,
}

var InvalidChartReferences = []string{
	"gs://missing-chart-segment",
	"https://not-a-gcs-url",
	"file://local-chart-reference",
}

func TestRepoProvider(t *testing.T) {
	rp := NewRepoProvider(nil, nil, nil)
	haveRepo, err := rp.GetRepoByURL(GCSPublicRepoURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := validateRepo(haveRepo, GCSPublicRepoURL, "", GCSRepoFormat, GCSRepoType); err != nil {
		t.Fatal(err)
	}

	castRepo, ok := haveRepo.(IStorageRepo)
	if !ok {
		t.Fatalf("invalid repo type, want: IStorageRepo, have: %T.", haveRepo)
	}

	wantBucket := GCSPublicRepoBucket
	haveBucket := castRepo.GetBucket()
	if haveBucket != wantBucket {
		t.Fatalf("unexpected bucket; want: %s, have %s.", wantBucket, haveBucket)
	}

	wantRepo := haveRepo
	URL := GCSPublicRepoURL + "/" + TestArchiveName
	haveRepo, err = rp.GetRepoByChartURL(URL)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantRepo, haveRepo) {
		t.Fatalf("retrieved invalid repo; want: %#v, have %#v.", haveRepo, wantRepo)
	}
}

func TestGetRepoByURLWithInvalidURL(t *testing.T) {
	var invalidURL = "https://valid.url/wrong/scheme"
	rp := NewRepoProvider(nil, nil, nil)
	_, err := rp.GetRepoByURL(invalidURL)
	if err == nil {
		t.Fatalf("found repo using invalid URL: %s", invalidURL)
	}
}

func TestGetRepoByChartURLWithInvalidChartURL(t *testing.T) {
	var invalidURL = "https://valid.url/wrong/scheme"
	rp := NewRepoProvider(nil, nil, nil)
	_, err := rp.GetRepoByChartURL(invalidURL)
	if err == nil {
		t.Fatalf("found repo using invalid chart URL: %s", invalidURL)
	}
}

func TestGetChartByReferenceWithValidReferences(t *testing.T) {
	rp := getTestRepoProvider(t)
	wantFile, err := chart.LoadChartfile(TestChartFile)
	if err != nil {
		t.Fatal(err)
	}

	for _, vcr := range ValidChartReferences {
		t.Logf("getting chart by reference: %s", vcr)
		tc, _, err := rp.GetChartByReference(vcr)
		if err != nil {
			t.Error(err)
			continue
		}

		haveFile := tc.Chartfile()
		if !reflect.DeepEqual(wantFile, haveFile) {
			t.Fatalf("retrieved invalid chart\nwant:%#v\nhave:\n%#v\n", wantFile, haveFile)
		}
	}
}

func getTestRepoProvider(t *testing.T) IRepoProvider {
	rp := newRepoProvider(nil, nil, nil)
	rs := rp.GetRepoService()
	tr, err := newRepo(TestRepoURL, TestRepoCredentialName, TestName, TestRepoFormat, TestRepoType)
	if err != nil {
		t.Fatalf("cannot create test repository: %s", err)
	}

	if err := rs.CreateRepo(tr); err != nil {
		t.Fatalf("cannot initialize repository service: %s", err)
	}

	return rp
}

func TestGetChartByReferenceWithInvalidReferences(t *testing.T) {
	rp := NewRepoProvider(nil, nil, nil)
	for _, icr := range InvalidChartReferences {
		_, _, err := rp.GetChartByReference(icr)
		if err == nil {
			t.Fatalf("found chart using invalid reference: %s", icr)
		}
	}
}

func TestIsGCSChartReferenceWithValidReferences(t *testing.T) {
	for _, vcr := range ValidChartReferences {
		if !IsGCSChartReference(vcr) {
			t.Fatalf("valid chart reference %s not accepted", vcr)
		}
	}
}

func TestIsGCSChartReferenceWithInvalidReferences(t *testing.T) {
	for _, icr := range InvalidChartReferences {
		if IsGCSChartReference(icr) {
			t.Fatalf("invalid chart reference %s accepted", icr)
		}
	}
}

func TestParseGCSChartReferences(t *testing.T) {
	for _, vcr := range ValidChartReferences {
		if _, err := ParseGCSChartReference(vcr); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseGCSChartReferenceWithInvalidReferences(t *testing.T) {
	for _, icr := range InvalidChartReferences {
		if _, err := ParseGCSChartReference(icr); err == nil {
			t.Fatalf("invalid chart reference %s parsed correctly", icr)
		}
	}
}
