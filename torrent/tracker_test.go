package torrent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type peerTestCase struct {
	metaInfoLocation string
	bodyLocation     string
	wantedError      error
}

func TestAnnounce(t *testing.T) {
	testCases := []peerTestCase{
		// Compact 0
		{"examples/ubuntu-22.04.3-desktop-amd64.iso.torrent", "examples/ubuntu-22.04.3-desktop-amd64.iso.torrent.compact0.announce", nil},
		// Compact 1
		{"examples/ubuntu-22.04.3-desktop-amd64.iso.torrent", "examples/ubuntu-22.04.3-desktop-amd64.iso.torrent.compact1.announce", nil},
	}

	for i := range testCases {
		metaInfoLocation := testCases[i].metaInfoLocation
		bodyLocation := testCases[i].bodyLocation
		wantedErr := testCases[i].wantedError

		metaInfoFileReader, err := os.Open(metaInfoLocation)
		if err != nil {
			t.Errorf("Could not open %s", metaInfoLocation)
		}
		defer metaInfoFileReader.Close()

		metaInfo, err := ParseMetaInfo(metaInfoFileReader)
		if err != nil {
			t.Errorf("Error parsing %s #%v", metaInfoLocation, err)
		}

		bodyResponseReader, err := os.Open(bodyLocation)
		if err != nil {
			t.Errorf("Could not open %s", bodyLocation)
		}
		defer bodyResponseReader.Close()

		announceUrl, err := url.Parse(metaInfo.Announce)
		if err != nil {
			t.Errorf("Could not parse URL %s", metaInfo.Announce)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == announceUrl.Path {
				bytes, err := io.ReadAll(bodyResponseReader)
				if err != nil {
					t.Errorf("Error reading body response from file")
				}

				if n, err := w.Write(bytes); n < len(bytes) || err != nil {
					t.Errorf("Error writing response")
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		}))
		defer server.Close()

		mockedUrl := server.URL + announceUrl.Path
		metaInfo.Announce = mockedUrl

		trackerAnnounceRequest := AnnounceRequest{
			AnnounceURL: mockedUrl,
			PeerId:      GenerateRandomProtocolId(),
			InfoHash:    metaInfo.GetInfoHash(),
			Port:        6699,
		}
		announceResponse, err := Announce(&trackerAnnounceRequest)

		if err == nil && wantedErr != nil {
			t.Errorf("Not expected, but got error %#v", err)
		}

		if err != nil && wantedErr != nil {
			if err != wantedErr {
				t.Errorf("got error %#v wanted %#v", err, wantedErr)
			}
		}

		if err == nil && announceResponse == nil {
			t.Errorf("Announce response is nil")
		}

		if err == nil && announceResponse != nil {
			if announceResponse.Interval == 0 || len(announceResponse.Peers) == 0 {
				t.Errorf("Not expected %#v", announceResponse)
			}
		}

	}
}
