package spring83

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func Put(server string, keypair ed25519.PrivateKey, tm time.Time, body []byte) error {
	if len(body) > MaxBoardLen {
		return errors.New("content + date longer than 2217 bytes")
	}

	pub := hex.EncodeToString(keypair[len(keypair)-ed25519.PublicKeySize:])
	sig := hex.EncodeToString(ed25519.Sign(keypair, body))

	req, err := http.NewRequest("PUT", "https://"+server+"/"+pub, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header["User-Agent"] = []string{"ahoy/0.1"}
	req.Header["If-Unmodified-Since"] = []string{tm.Format(http.TimeFormat)}

	req.Header["Content-Type"] = []string{"text/html;charset=utf-8"}
	req.Header["Content-Length"] = []string{strconv.Itoa(len(body))}

	req.Header["Spring-Version"] = []string{"83"}
	req.Header["Spring-Signature"] = []string{sig}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return fmt.Errorf("non-OK response: %s", resp.Status)
	}

	return nil
}

func Get(client http.Client, server string, pubkey string) ([]byte, error) {
	req, err := http.NewRequest("GET", "https://"+server+"/"+pubkey, nil)
	if err != nil {
		return nil, err
	}

	req.Header["User-Agent"] = []string{"ahoy/0.1"}
	req.Header["Spring-Version"] = []string{"83"}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-OK response: %s", resp.Status)
	}

	board, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return board, nil
}
