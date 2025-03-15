package openpgp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/pkg/errors"

	"github.com/wal-g/wal-g/internal/crypto"
	"github.com/wal-g/wal-g/internal/crypto/envelope"
	"github.com/wal-g/wal-g/internal/ioextensions"
)

const (
	maxHeaderLenAllowed int = 4096 * 2
)

// Crypter incapsulates specific of cypher method
// Includes keys, infrastructure information etc
type Crypter struct {
	enveloper    envelope.Enveloper
	encryptedKey *envelope.EncryptedKey

	ArmoredKey      string
	IsUseArmoredKey bool

	ArmoredKeyPath      string
	IsUseArmoredKeyPath bool

	mutex sync.RWMutex
}

func (crypter *Crypter) Name() string {
	parts := []string{"Enveloped", crypter.enveloper.Name(), "Opengpg", "Crypter"}
	return strings.Join(parts, "/")
}

// Encrypt creates encryption writer from ordinary writer
func (crypter *Crypter) Encrypt(writer io.Writer) (io.WriteCloser, error) {
	err := crypter.setupEncryptedKey()
	if err != nil {
		return nil, err
	}
	key, err := crypter.enveloper.DecryptKey(crypter.encryptedKey)
	if err != nil {
		return nil, errors.Wrapf(err, "can't decrypt encryption key")
	}
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
	if err != nil {
		return nil, errors.Wrapf(err, "can't read decrypyed gpg key")
	}
	keyID, err := encodeKeyID(entityList)
	if err != nil {
		return nil, errors.Wrapf(err, "can't encode gpg key id")
	}

	// need write header at first, with length less than maxHeaderLenAllowed
	bufferedWriter := bufio.NewWriterSize(writer, maxHeaderLenAllowed)

	header := crypter.enveloper.SerializeEncryptedKey(crypter.encryptedKey.WithID(keyID))

	if len(header) > maxHeaderLenAllowed {
		return nil, errors.New("opengpg: invalid size of the encrypted key")
	}
	_, err = bufferedWriter.Write(header)
	if err != nil {
		return nil, errors.Wrapf(err, "can't write encryption key to buffer")
	}

	encryptedWriter, err := openpgp.Encrypt(bufferedWriter, entityList, nil, nil, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "opengpg encryption error")
	}

	return ioextensions.NewOnCloseFlusher(encryptedWriter, bufferedWriter), nil
}

// Decrypt creates decrypted reader from ordinary reader
func (crypter *Crypter) Decrypt(reader io.Reader) (io.Reader, error) {
	// need read header at first, with length less than maxHeaderLenAllowed
	bufferedReader := bufio.NewReaderSize(reader, maxHeaderLenAllowed)
	encryptedKey, err := crypter.enveloper.ReadEncryptedKey(bufferedReader)
	if err != nil {
		return nil, err
	}
	var key []byte
	key, err = crypter.enveloper.DecryptKey(encryptedKey)
	if err != nil {
		return nil, errors.Wrapf(err, "can't decrypt encryption key")
	}
	secretKey, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
	if err != nil {
		return nil, errors.Wrapf(err, "can't read decrypyed gpg key")
	}

	md, err := openpgp.ReadMessage(bufferedReader, secretKey, nil, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "opengpg decryption error")
	}

	return md.UnverifiedBody, nil
}

func (crypter *Crypter) setupEncryptedKey() error {
	crypter.mutex.RLock()
	if crypter.encryptedKey != nil {
		crypter.mutex.RUnlock()
		return nil
	}
	crypter.mutex.RUnlock()

	crypter.mutex.Lock()
	defer crypter.mutex.Unlock()
	if crypter.encryptedKey != nil {
		return nil
	}
	var (
		rawEncryptedKey []byte
		err             error
	)
	switch {
	case crypter.IsUseArmoredKey:
		rawEncryptedKey, err = readFromString(crypter.ArmoredKey)
		if err != nil {
			return err
		}
	case crypter.IsUseArmoredKeyPath:
		rawEncryptedKey, err = readFromFilePath(crypter.ArmoredKeyPath)
		if err != nil {
			return err
		}
	}
	crypter.encryptedKey = envelope.NewEncryptedKey("", rawEncryptedKey)
	return nil
}

// CrypterFromKey creates Crypter from encrypted armored key.
func CrypterFromKey(armoredKey string, enveloper envelope.Enveloper) crypto.Crypter {
	return &Crypter{
		ArmoredKey:      armoredKey,
		IsUseArmoredKey: true,
		enveloper:       enveloper,
	}
}

// CrypterFromKeyPath creates Crypter from encrypted armored key path.
func CrypterFromKeyPath(armoredKeyPath string, enveloper envelope.Enveloper) crypto.Crypter {
	return &Crypter{
		ArmoredKeyPath:      armoredKeyPath,
		IsUseArmoredKeyPath: true,
		enveloper:           enveloper,
	}
}

func readFromString(content string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(content)
}

func readFromFilePath(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	encryptedKey := make([]byte, base64.StdEncoding.DecodedLen(len(content)))
	var decodedLen int
	decodedLen, err = base64.StdEncoding.Decode(encryptedKey, content)
	if err != nil {
		return nil, err
	}
	// DecodedLen returns the maximum length in bytes of the decoded data
	// which Decode writes at most, that's why need to be sliced
	// with actually written length
	return encryptedKey[:decodedLen], nil
}

func encodeKeyID(entities openpgp.EntityList) (string, error) {
	parts := make([]string, len(entities))

	for id, entity := range entities {
		_, ok := entity.EncryptionKey(time.Now())
		if !ok {
			return "", errors.Errorf("opengpg: %ds key has no valid encryption keys", id)
		}
		parts[id] = strings.ToUpper(strconv.FormatUint(entity.PrimaryKey.KeyId, 16))
	}
	return strings.Join(parts, ","), nil
}
