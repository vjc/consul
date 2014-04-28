package agent

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/consul/consul/structs"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestKVSEndpoint_PUT_GET_DELETE(t *testing.T) {
	dir, srv := makeHTTPServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer srv.agent.Shutdown()

	// Wait for a leader
	time.Sleep(100 * time.Millisecond)

	keys := []string{
		"baz",
		"bar",
		"foo/sub1",
		"foo/sub2",
		"zip",
	}

	for _, key := range keys {
		buf := bytes.NewBuffer([]byte("test"))
		req, err := http.NewRequest("PUT", "/v1/kv/"+key, buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); !res {
			t.Fatalf("should work")
		}
	}

	for _, key := range keys {
		req, err := http.NewRequest("GET", "/v1/kv/"+key, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertIndex(t, resp)

		res, ok := obj.(structs.DirEntries)
		if !ok {
			t.Fatalf("should work")
		}

		if len(res) != 1 {
			t.Fatalf("bad: %v", res)
		}

		if res[0].Key != key {
			t.Fatalf("bad: %v", res)
		}
	}

	for _, key := range keys {
		req, err := http.NewRequest("DELETE", "/v1/kv/"+key, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}

func TestKVSEndpoint_Recurse(t *testing.T) {
	dir, srv := makeHTTPServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer srv.agent.Shutdown()

	// Wait for a leader
	time.Sleep(100 * time.Millisecond)

	keys := []string{
		"bar",
		"baz",
		"foo/sub1",
		"foo/sub2",
		"zip",
	}

	for _, key := range keys {
		buf := bytes.NewBuffer([]byte("test"))
		req, err := http.NewRequest("PUT", "/v1/kv/"+key, buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); !res {
			t.Fatalf("should work")
		}
	}

	{
		// Get all the keys
		req, err := http.NewRequest("GET", "/v1/kv/?recurse", nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertIndex(t, resp)

		res, ok := obj.(structs.DirEntries)
		if !ok {
			t.Fatalf("should work")
		}

		if len(res) != len(keys) {
			t.Fatalf("bad: %v", res)
		}

		for idx, key := range keys {
			if res[idx].Key != key {
				t.Fatalf("bad: %v %v", res[idx].Key, key)
			}
		}
	}

	{
		req, err := http.NewRequest("DELETE", "/v1/kv/?recurse", nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	{
		// Get all the keys
		req, err := http.NewRequest("GET", "/v1/kv/?recurse", nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if obj != nil {
			t.Fatalf("bad: %v", obj)
		}
	}
}

func TestKVSEndpoint_CAS(t *testing.T) {
	dir, srv := makeHTTPServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer srv.agent.Shutdown()

	// Wait for a leader
	time.Sleep(100 * time.Millisecond)

	{
		buf := bytes.NewBuffer([]byte("test"))
		req, err := http.NewRequest("PUT", "/v1/kv/test?flags=50", buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); !res {
			t.Fatalf("should work")
		}
	}

	req, err := http.NewRequest("GET", "/v1/kv/test", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	resp := httptest.NewRecorder()
	obj, err := srv.KVSEndpoint(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	d := obj.(structs.DirEntries)[0]

	// Check the flags
	if d.Flags != 50 {
		t.Fatalf("bad: %v", d)
	}

	// Create a CAS request, bad index
	{
		buf := bytes.NewBuffer([]byte("zip"))
		req, err := http.NewRequest("PUT",
			fmt.Sprintf("/v1/kv/test?flags=42&cas=%d", d.ModifyIndex-1), buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); res {
			t.Fatalf("should NOT work")
		}
	}

	// Create a CAS request, good index
	{
		buf := bytes.NewBuffer([]byte("zip"))
		req, err := http.NewRequest("PUT",
			fmt.Sprintf("/v1/kv/test?flags=42&cas=%d", d.ModifyIndex), buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); !res {
			t.Fatalf("should work")
		}
	}

	// Verify the update
	req, _ = http.NewRequest("GET", "/v1/kv/test", nil)
	resp = httptest.NewRecorder()
	obj, _ = srv.KVSEndpoint(resp, req)
	d = obj.(structs.DirEntries)[0]

	if d.Flags != 42 {
		t.Fatalf("bad: %v", d)
	}
	if string(d.Value) != "zip" {
		t.Fatalf("bad: %v", d)
	}
}

func TestKVSEndpoint_ListKeys(t *testing.T) {
	dir, srv := makeHTTPServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer srv.agent.Shutdown()

	// Wait for a leader
	time.Sleep(100 * time.Millisecond)

	keys := []string{
		"bar",
		"baz",
		"foo/sub1",
		"foo/sub2",
		"zip",
	}

	for _, key := range keys {
		buf := bytes.NewBuffer([]byte("test"))
		req, err := http.NewRequest("PUT", "/v1/kv/"+key, buf)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if res := obj.(bool); !res {
			t.Fatalf("should work")
		}
	}

	{
		// Get all the keys
		req, err := http.NewRequest("GET", "/v1/kv/?keys&seperator=/", nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		obj, err := srv.KVSEndpoint(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertIndex(t, resp)

		res, ok := obj.([]string)
		if !ok {
			t.Fatalf("should work")
		}

		expect := []string{"bar", "baz", "foo/", "zip"}
		if !reflect.DeepEqual(res, expect) {
			t.Fatalf("bad: %v", res)
		}
	}
}
