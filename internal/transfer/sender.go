package transfer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sanyamgarg/airpipe/internal/crypto"
)

const ChunkSize = 64 * 1024

type Sender struct {
	relayURL string
	token    string
	key      []byte
	conn     *websocket.Conn
}

func NewSender(relayURL, token string, key []byte) *Sender {
	return &Sender{relayURL: relayURL, token: token, key: key}
}

func (s *Sender) Connect() error {
	url := fmt.Sprintf("%s/ws/%s", s.relayURL, s.token)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}
	s.conn = conn

	versionMsg := NewVersionMessage()
	encryptedVersion, err := crypto.EncryptChunk(EncodeMessage(versionMsg), s.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt version message: %w", err)
	}
	if err := s.conn.WriteMessage(websocket.BinaryMessage, encryptedVersion); err != nil {
		return fmt.Errorf("failed to send version message: %w", err)
	}

	return nil
}

func (s *Sender) WaitForReceiver(timeout time.Duration) error {
	s.conn.SetReadDeadline(time.Now().Add(timeout))
	_, message, err := s.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("timeout waiting for receiver: %w", err)
	}
	
	decrypted, err := crypto.DecryptChunk(message, s.key)
	if err != nil {
		return fmt.Errorf("failed to decrypt ready message: %w", err)
	}
	
	msg, err := DecodeMessage(decrypted)
	if err != nil {
		return err
	}
	if msg.Type != MsgTypeReady {
		return fmt.Errorf("unexpected message type: %d", msg.Type)
	}
	s.conn.SetReadDeadline(time.Time{})
	return nil
}

func (s *Sender) SendFile(filePath string, progressFn func(sent, total int64)) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	filename := filepath.Base(filePath)
	fileSize := stat.Size()
	totalChunks := int((fileSize + ChunkSize - 1) / ChunkSize)

	metaMsg, err := NewMetadataMessage(filename, fileSize, totalChunks)
	if err != nil {
		return err
	}

	encryptedMeta, err := crypto.EncryptChunk(EncodeMessage(metaMsg), s.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt metadata: %w", err)
	}

	if err := s.conn.WriteMessage(websocket.BinaryMessage, encryptedMeta); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	buf := make([]byte, ChunkSize)
	var bytesSent int64

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		chunkMsg := NewChunkMessage(buf[:n])
		encryptedChunk, err := crypto.EncryptChunk(EncodeMessage(chunkMsg), s.key)
		if err != nil {
			return fmt.Errorf("failed to encrypt chunk: %w", err)
		}

		if err := s.conn.WriteMessage(websocket.BinaryMessage, encryptedChunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}

		bytesSent += int64(n)
		if progressFn != nil {
			progressFn(bytesSent, fileSize)
		}
	}

	completeMsg := NewCompleteMessage()
	encryptedComplete, err := crypto.EncryptChunk(EncodeMessage(completeMsg), s.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt complete message: %w", err)
	}

	if err := s.conn.WriteMessage(websocket.BinaryMessage, encryptedComplete); err != nil {
		return fmt.Errorf("failed to send complete message: %w", err)
	}

	return nil
}

func (s *Sender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
