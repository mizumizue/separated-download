package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	mockdownloader "separated-download/mock/downloader"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNewDownloader(t *testing.T) {
	type args struct {
		c  IHttpClient
		cc []DownloadConfig
	}
	tests := []struct {
		name string
		args args
		want *Downloader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDownloader(tt.args.c, tt.args.cc...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDownloader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloader_Download(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		config DownloadConfig
	}
	type args struct {
		url string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		prepare     func(client *mockdownloader.MockIHttpClient)
		want        *File
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Case1: Unknown Accept-Range unit",
			args: args{
				url: "https://example.com",
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				header := http.Header{}
				header.Add("Accept-Ranges", "MB")
				url := "https://example.com"
				mock.EXPECT().Head(url).Return(&http.Response{
					Header: header,
				}, nil)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: ErrUnknownAcceptRange,
		},
		{
			name: "Case2: Head Request failed...",
			args: args{
				url: "hoge",
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				errUrl := "hoge"
				mock.EXPECT().Head(errUrl).Return(nil, errors.New("request failed... "))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: ErrHeadRequestFailed,
		},
		{
			name: "Case3: Content-Length parse error",
			args: args{
				url: "https://example2.com",
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				url := "https://example2.com"
				header := http.Header{}
				header.Add("Accept-Ranges", "bytes")
				header.Add("Content-Length", "mogemoge")
				mock.EXPECT().Head(url).Return(&http.Response{
					Header: header,
				}, nil)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: ErrUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockdownloader.NewMockIHttpClient(ctrl)
			tt.prepare(mock)
			d := &Downloader{
				c:      mock,
				config: tt.fields.config,
			}
			got, err := d.Download(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Download() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloader_fetchChunk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		config DownloadConfig
	}
	type args struct {
		url string
		r   Range
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		prepare     func(mock *mockdownloader.MockIHttpClient)
		want        []byte
		want1       int
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Case1: ErrGenerateHttpRequest",
			args: args{
				url: "https://example.com",
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				url := "https://example.com"
				mock.EXPECT().GenerateRequest(url).Return(nil, http.ErrServerClosed)
			},
			want:        nil,
			want1:       0,
			wantErr:     true,
			expectedErr: ErrGenerateHttpRequest,
		},
		{
			name: "Case2: ErrHttpRequestFailed",
			args: args{
				url: "https://example.com",
				r: Range{
					start: 0,
					end:   500,
				},
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				url := "https://example.com"
				req, _ := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", 0, 500))
				mock.EXPECT().GenerateRequest(url).Return(req, nil)
				mock.EXPECT().Do(req).Return(nil, ErrHttpRequestFailed)
			},
			want:        nil,
			want1:       0,
			wantErr:     true,
			expectedErr: ErrHttpRequestFailed,
		},
		{
			name: "Case3: ErrHttpBadStatusCode",
			args: args{
				url: "https://example.com",
				r: Range{
					start: 0,
					end:   500,
				},
			},
			prepare: func(mock *mockdownloader.MockIHttpClient) {
				url := "https://example.com"
				req, _ := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", 0, 500))
				mock.EXPECT().GenerateRequest(url).Return(req, nil)
				mock.EXPECT().Do(req).Return(&http.Response{
					Status:     "BadRequest",
					StatusCode: http.StatusBadRequest,
				}, nil)
			},
			want:        nil,
			want1:       400,
			wantErr:     true,
			expectedErr: ErrHttpBadStatusCode,
		},
		// TODO Add lack of test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockdownloader.NewMockIHttpClient(ctrl)
			tt.prepare(mock)
			d := &Downloader{
				c:      mock,
				config: tt.fields.config,
			}
			got, got1, err := d.fetchChunk(tt.args.url, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchChunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("fetchChunk() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchChunk() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("fetchChunk() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
