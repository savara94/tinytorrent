package torrent

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"example.com/bencode"
)

type AnnounceRequest struct {
	AnnounceURL string
	PeerId      []byte
	InfoHash    []byte
	Port        int
	Uploaded    int
	Downloaded  int
	Left        int
}

type PeerInfo struct {
	PeerId []byte
	IP     net.IP
	Port   int
}

type AnnounceResponse struct {
	Interval    int
	Peers       []PeerInfo
	RawResponse []byte
}

func buildTrackerRequest(request *AnnounceRequest) (*http.Request, error) {
	httpReq, err := http.NewRequest("GET", request.AnnounceURL, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}

	q.Add("info_hash", string(request.InfoHash))
	q.Add("peer_id", string(request.PeerId))
	// IP is optional
	// q.Add("ip", "")
	q.Add("port", strconv.Itoa(request.Port))
	q.Add("uploaded", strconv.Itoa(request.Uploaded))
	q.Add("downloaded", strconv.Itoa(request.Downloaded))
	q.Add("left", strconv.Itoa(request.Left))
	// Event is optional -> event=started|completed|stopped
	// q.Add("event")

	httpReq.URL.RawQuery = q.Encode()

	return httpReq, nil
}

func readAnnounceResponse(response *http.Response) (*AnnounceResponse, error) {
	if response.StatusCode < 200 || response.StatusCode > 300 {
		errMsg := fmt.Sprintf("Server responded with NOK %#v", response)
		return nil, errors.New(errMsg)
	}

	return ParseAnnounceResponse(response.Body)
}

func parseStandardAnnounceResponse(data []byte) (*AnnounceResponse, error) {
	type peerInfo struct {
		PeerId string `bencode:"peer id"`
		IP     string `bencode:"ip"`
		Port   int    `bencode:"port"`
	}

	type standardAnnounceResponse struct {
		Interval      *int        `bencode:"interval"`
		Peers         *[]peerInfo `bencode:"peers"`
		FailureReason *string     `bencode:"failure reason"`
	}

	var response standardAnnounceResponse

	err := bencode.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	if response.FailureReason != nil {
		return nil, errors.New(*response.FailureReason)
	}

	if response.Interval == nil || response.Peers == nil {
		// This should not happen
		errMsg := fmt.Sprintf("Unexpected error, tracker response invalid: %#v", response)
		return nil, errors.New(errMsg)
	}

	announceResponse := AnnounceResponse{Interval: *response.Interval}

	for i := range *response.Peers {
		pInfo := PeerInfo{PeerId: []byte((*response.Peers)[i].PeerId), IP: net.ParseIP(((*response.Peers)[i].IP)), Port: (*response.Peers)[i].Port}

		if pInfo.IP == nil {
			continue
		}

		announceResponse.Peers = append(announceResponse.Peers, pInfo)
	}

	return &announceResponse, nil
}

func parseCompactAnnounceResponse(data []byte) (*AnnounceResponse, error) {
	type compactAnnounceResponse struct {
		Interval      *int    `bencode:"interval"`
		Peers         *string `bencode:"peers"`
		FailureReason *string `bencode:"failure reason"`
	}

	var response compactAnnounceResponse

	err := bencode.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	if response.FailureReason != nil {
		return nil, errors.New(*response.FailureReason)
	}

	if response.Interval == nil || response.Peers == nil {
		// This should not happen
		errMsg := fmt.Sprintf("Unexpected error, tracker response invalid: %#v", response)
		return nil, errors.New(errMsg)
	}

	announceResponse := AnnounceResponse{Interval: *response.Interval}

	peerBytes := []byte(*(response.Peers))
	// 4 bytes IP address, 2 bytes port
	for i := 0; i < len(peerBytes); i += 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peerBytes[i], peerBytes[i+1], peerBytes[i+2], peerBytes[i+3])

		var port uint16

		port = uint16(peerBytes[i+4]) << 8
		port = port | uint16(peerBytes[i+5])

		// PeerId is not supplied in compact format
		pInfo := PeerInfo{IP: net.ParseIP(ip), Port: int(port)}
		announceResponse.Peers = append(announceResponse.Peers, pInfo)
	}

	return &announceResponse, nil
}

func ParseAnnounceResponse(reader io.Reader) (*AnnounceResponse, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try with standard format first
	announceResponse, err := parseStandardAnnounceResponse(bytes)
	if err != nil {
		slog.Info("Standard parser failed " + err.Error())

		// ...proceed with compact format
		announceResponse, err = parseCompactAnnounceResponse(bytes)
		if err != nil {
			slog.Info("Compact parser failed " + err.Error())

			return nil, errors.New("Could not parse announce response.")
		}
	}

	announceResponse.RawResponse = bytes

	return announceResponse, nil
}

func Announce(request *AnnounceRequest) (*AnnounceResponse, error) {
	httpRequest, err := buildTrackerRequest(request)
	if err != nil {
		return nil, err
	}

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	trackerResponse, err := readAnnounceResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	return trackerResponse, nil
}
