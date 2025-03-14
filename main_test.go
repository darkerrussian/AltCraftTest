package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateShortCode(t *testing.T) {
	url := "http://example.com"
	code1 := generateShortCode(url)
	code2 := generateShortCode(url)

	//Проверяем одинаковый ли код
	if code1 != code2 {
		t.Errorf("Expected %s, got %s", code1, code2)
	}

	if len(code1) != 8 {
		t.Errorf("ShortLink more than 8 characters, now %d", len(code1))
	}

}

func TestShortLinkHandle(t *testing.T) {
	//Тестовый запрос...
	req, err := http.NewRequest("GET", "/a/?url=http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	//Response Recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(shortLinkHandle)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, go %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if len(body) != 8 {
		t.Errorf("ShortLink more than 8 characters, now %d", len(body))
	}
}

func TestHandleRedirect(t *testing.T) {
	testCode := "test322228"
	testUrl := "http://example.com"
	err := rdb.Set(ctx, testCode, testUrl, 0).Err()
	if err != nil {
		t.Fatal(err)
	}

	//Тестовый запрос
	req, err := http.NewRequest("GET", "/s/"+testCode, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRedirect)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusFound, status)
	}
	//Проверяем правильность конечного url
	location := rr.Header().Get("Location")
	if location != testUrl {
		t.Errorf("Expected location %s, but got %s", testUrl, location)
	}
}
