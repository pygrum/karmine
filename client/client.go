package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pygrum/karmine/krypto/kryptor"
)

func NewMTLSClient(certFile, pemFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, pemFile)
	if err != nil {
		return nil, err
	}
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // testing
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
			},
		},
		Timeout: 5 * time.Second,
	}, nil
}

func TLSDialConfig(certdata, keydata, x1, x2 string) (*tls.Config, error) {
	cbytes, err := kryptor.Decrypt(certdata, x1, x2)
	if err != nil {
		return nil, err
	}
	kbytes, err := kryptor.Decrypt(keydata, x1, x2)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(cbytes, kbytes)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cbytes)

	return &tls.Config{
		InsecureSkipVerify: true, // testing
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

func MTLsClientByKryptor(certdata, keydata, x1, x2 string) (*http.Client, error) {
	cbytes, err := kryptor.Decrypt(certdata, x1, x2)
	if err != nil {
		return nil, err
	}
	kbytes, err := kryptor.Decrypt(keydata, x1, x2)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(cbytes, kbytes)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cbytes)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // testing
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
			},
		},
		Timeout: 5 * time.Second,
	}, nil
}
