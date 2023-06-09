package acss

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/setherplatform/sether-node/common"
	"github.com/setherplatform/sether-node/common/sharing"
	"io"

	"github.com/coinbase/kryptology/pkg/core/curves"
	kryptsharing "github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

func Encrypt(share []byte, public curves.Point, priv curves.Scalar) ([]byte, error) {
	key := public.Mul(priv)
	keyHash := sha256.Sum256(key.ToAffineCompressed())
	cipher, err := encryptAES(keyHash[:], share) // 52 bytes
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	return cipher, nil
}

func encryptAES(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Infof("Key error: err=%v", err)
		return nil, fmt.Errorf("encrypt_aes: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("encrypt_aes: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func decryptAES(key []byte, ciphertext []byte) ([]byte, error) {
	input := make([]byte, len(ciphertext))
	copy(input, ciphertext)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt_aes: %w", err)
	}

	if len(input) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := input[:aes.BlockSize]
	input = input[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(input, input)
	return input, nil
}

func CompressCommitments(v *sharing.FeldmanVerifier) []byte {
	c := make([]byte, 0)
	for _, v := range v.Commitments {
		e := v.ToAffineCompressed() // 33 bytes
		c = append(c, e[:]...)
	}
	return c
}

func DecompressCommitments(k int, c []byte, curve *curves.Curve) ([]curves.Point, error) {
	commitment := make([]curves.Point, 0)
	for i := 0; i < k; i++ {
		cI, err := curve.Point.FromAffineCompressed(c[i*33 : (i*33)+33])
		if err == nil {
			commitment = append(commitment, cI)
		} else {
			return nil, err
		}
	}

	return commitment, nil
}

func verifierFromCommits(k int, c []byte, curve *curves.Curve) (*sharing.FeldmanVerifier, error) {

	commitment, err := DecompressCommitments(k, c, curve)
	if err != nil {
		return nil, err
	}
	verifier := new(sharing.FeldmanVerifier)
	verifier.Commitments = commitment
	return verifier, nil
}

func GenerateKeyPair(curve *curves.Curve) common.KeyPair {
	g := curve.NewGeneratorPoint()
	privateKey := curve.NewScalar().Random(rand.Reader)
	publicKey := g.Mul(privateKey)
	return common.KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}
func GenerateSecret(c *curves.Curve) curves.Scalar {
	secret := c.Scalar.Random(rand.Reader)
	return secret
}

func GenerateCommitmentAndShares(s curves.Scalar, k, n uint32, curve *curves.Curve) (*sharing.FeldmanVerifier, []sharing.ShamirShare, error) {
	f, err := sharing.NewFeldman(k, n, curve)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}

	feldcommit, shares, err := Split(s, f.Threshold, f.Limit, f.Curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}
	return feldcommit, shares, nil
}

func Split(secret curves.Scalar, threshold, limit uint32, curve *curves.Curve, reader io.Reader) (*sharing.FeldmanVerifier, []sharing.ShamirShare, error) {

	shares, poly := getPolyAndShares(secret, threshold, limit, curve, reader)
	verifier := new(sharing.FeldmanVerifier)
	verifier.Commitments = make([]curves.Point, threshold)
	for i := range verifier.Commitments {
		base, _ := sharing.CurveParams(curve.Name)
		verifier.Commitments[i] = base.Mul(poly.Coefficients[i])
	}
	return verifier, shares, nil
}

func getPolyAndShares(
	secret curves.Scalar,
	threshold, limit uint32,
	curve *curves.Curve,
	reader io.Reader) ([]sharing.ShamirShare, *kryptsharing.Polynomial) {
	poly := new(kryptsharing.Polynomial).Init(secret, threshold, reader)
	shares := make([]sharing.ShamirShare, limit)
	for i := range shares {
		x := curve.Scalar.New(i + 1)
		shares[i] = sharing.ShamirShare{
			Id:    uint32(i + 1),
			Value: poly.Evaluate(x).Bytes(),
		}
	}
	return shares, poly
}

func SharedKey(priv curves.Scalar, dealerPublicKey curves.Point) [32]byte {
	key := dealerPublicKey.Mul(priv)
	keyHash := sha256.Sum256(key.ToAffineCompressed())
	return keyHash
}

// Predicate verifies if the share fits the polynomial commitments
func Predicate(key []byte, cipher []byte, commits []byte, k int, curve *curves.Curve) (*sharing.ShamirShare, *sharing.FeldmanVerifier, bool) {

	shareBytes, err := decryptAES(key, cipher)
	if err != nil {
		log.Errorf("Error while decrypting share: err=%s", err)
		return nil, nil, false
	}
	share := sharing.ShamirShare{Id: binary.BigEndian.Uint32(shareBytes[:4]), Value: shareBytes[4:]}
	log.Debugf("share: id=%d, val=%v", share.Id, share.Value)
	verifier, err := verifierFromCommits(k, commits, curve)
	if err != nil {
		log.Errorf("Error while getting verifier from commits=%s", err)
		return nil, nil, false
	}

	if err = verifier.Verify(&share); err != nil {
		log.Errorf("Error while verifying share=%s", err)
		return nil, nil, false
	}
	return &share, verifier, true
}

func Encode(encoder *infectious.FEC, msg []byte) ([]infectious.Share, error) {
	shares := make([]infectious.Share, encoder.Total())
	output := func(s infectious.Share) {
		shares[s.Number] = s.DeepCopy()
	}

	paddedMsg, err := pkcs7Pad(msg, encoder.Required())
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(paddedMsg, output)
	if err != nil {
		return nil, err
	}

	return shares, nil
}

func Decode(f *infectious.FEC, s []infectious.Share) ([]byte, error) {
	result, err := f.Decode(nil, s)
	if err != nil {
		return nil, err
	}

	unpaddedMsg, err := pkcs7Unpad(result, f.Required())
	return unpaddedMsg, err
}

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}
