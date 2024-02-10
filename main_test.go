package main

import (
	"io"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_arrayVar_Set(t *testing.T) {
	exampleHttpUrl, _ := url.Parse("http://www.example.com")
	exampleHttpsUrl, _ := url.Parse("https://www.example.com")
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		a       *arrayVar
		args    args
		wantArr *arrayVar
		wantErr bool
	}{
		{"Set HTTP URL", &arrayVar{}, args{"http://www.example.com"}, &arrayVar{*exampleHttpUrl}, false},
		{"Set HTTPS URL", &arrayVar{}, args{"https://www.example.com"}, &arrayVar{*exampleHttpsUrl}, false},
		{"Add another URL", &arrayVar{*exampleHttpUrl}, args{"https://www.example.com"}, &arrayVar{*exampleHttpUrl, *exampleHttpsUrl}, false},
		{"Set mailto URL", &arrayVar{}, args{"mailto:test@example.com"}, &arrayVar{}, true},
		{"Set relative URL", &arrayVar{}, args{"/"}, &arrayVar{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("arrayVar.Set() error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.a, tt.wantArr) {
				t.Errorf("extractLinks() = %#v, want %#v", tt.a, tt.wantArr)
			}
		})
	}
}

func Test_arrayVar_String(t *testing.T) {
	exampleHttpUrl, _ := url.Parse("http://www.example.com")
	exampleHttpsUrl, _ := url.Parse("https://www.example.com")
	tests := []struct {
		name string
		a    *arrayVar
		want string
	}{
		{"Empty array", &arrayVar{}, "URLs: []"},
		{"Single URL", &arrayVar{*exampleHttpUrl}, "URLs: [http://www.example.com]"},
		{"Multiple URLs", &arrayVar{*exampleHttpUrl, *exampleHttpsUrl}, "URLs: [http://www.example.com https://www.example.com]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.String(); got != tt.want {
				t.Errorf("arrayVar.String() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_isHttpURL(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *url.URL
		wantErr bool
	}{
		{"HTTP URL", args{"http://www.example.com"}, &url.URL{Scheme: "http", Host: "www.example.com"}, false},
		{"HTTPS URL", args{"https://www.example.com"}, &url.URL{Scheme: "https", Host: "www.example.com"}, false},
		{"Mailto URL", args{"mailto:test@example.com"}, nil, true},
		{"Relative URL", args{"/"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isHttpURL(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("isHttpURL() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isHttpURL() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_extractLinks(t *testing.T) {
	type args struct {
		url  *url.URL
		body io.Reader
	}
	tests := []struct {
		name      string
		args      args
		wantLinks []*url.URL
		wantErr   bool
	}{
		{"No links", args{&url.URL{Scheme: "http", Host: "www.example.com"}, strings.NewReader("")}, []*url.URL(nil), false},
		{"Single link", args{&url.URL{Scheme: "http", Host: "www.example.com"}, strings.NewReader("<a href=\"http://www.example.com\">Example</a>")}, []*url.URL{{Scheme: "http", Host: "www.example.com"}}, false},
		{"Multiple links", args{&url.URL{Scheme: "http", Host: "www.example.com"}, strings.NewReader("<a href=\"http://www.example.com\">Example</a><a href=\"https://www.example.com\">Example</a>")}, []*url.URL{{Scheme: "http", Host: "www.example.com"}, {Scheme: "https", Host: "www.example.com"}}, false},
		{"Invalid link", args{&url.URL{Scheme: "http", Host: "www.example.com"}, strings.NewReader("<a href=\"mailto:test@example.com\">Example</a>")}, []*url.URL(nil), false},
		{"Relative link", args{&url.URL{Scheme: "http", Host: "www.example.com"}, strings.NewReader("<a href=\"/\">Example</a>")}, []*url.URL{{Scheme: "http", Host: "www.example.com", Path: "/"}}, false},
		{"Link with path", args{&url.URL{Scheme: "http", Host: "www.example.com", Path: "/hello"}, strings.NewReader("<a href=\"http://www.example.com/hello\">Example</a>")}, []*url.URL{{Scheme: "http", Host: "www.example.com", Path: "/hello"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks, err := extractLinks(tt.args.url, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractLinks() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("extractLinks() = %#v, want %#v", gotLinks, tt.wantLinks)
			}
		})
	}
}
