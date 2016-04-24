package state

import (
	"bytes"
	"encoding/binary"

	"github.com/Impyy/tox4go/crypto"
)

// Friend represents the structure of friends that can be found inside a Tox
// state file.
type Friend struct {
	Status         FriendStatus
	UserStatus     UserStatus
	PublicKey      *[crypto.PublicKeySize]byte
	RequestMessage string
	Name           string
	StatusMessage  string
	Nospam         uint32
	LastSeen       uint64
}

type sectionFriends struct {
	Friends []*Friend
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *sectionFriends) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		friend := new(Friend)

		friendStatus, err := reader.ReadByte()
		if err != nil {
			return err
		}
		friend.Status = FriendStatus(friendStatus)

		friend.PublicKey = new([crypto.PublicKeySize]byte)
		_, err = reader.Read(friend.PublicKey[:])
		if err != nil {
			return err
		}

		reqMessage, err := readStringWithSize(reader, maxRequestMessageSize, 1)
		if err != nil {
			return err
		}
		friend.RequestMessage = reqMessage

		name, err := readStringWithSize(reader, maxNameSize, 0)
		if err != nil {
			return err
		}
		friend.Name = name

		statusMessage, err := readStringWithSize(reader, maxStatusMessageSize, 1)
		if err != nil {
			return err
		}
		friend.StatusMessage = statusMessage

		userStatus, err := reader.ReadByte()
		if err != nil {
			return err
		}
		friend.UserStatus = UserStatus(userStatus)

		//skip padding
		_, err = reader.Seek(3, 1)
		if err != nil {
			return err
		}

		err = binary.Read(reader, binary.LittleEndian, &friend.Nospam)
		if err != nil {
			return err
		}

		err = binary.Read(reader, binary.BigEndian, &friend.LastSeen)
		if err != nil {
			return err
		}
		s.Friends = append(s.Friends, friend)
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *sectionFriends) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	for _, friend := range s.Friends {
		err := buff.WriteByte(byte(friend.Status))
		if err != nil {
			return nil, err
		}

		_, err = buff.Write(friend.PublicKey[:])
		if err != nil {
			return nil, err
		}

		err = writeStringWithSize(buff, friend.RequestMessage, maxRequestMessageSize, 1)
		if err != nil {
			return nil, err
		}

		err = writeStringWithSize(buff, friend.Name, maxNameSize, 0)
		if err != nil {
			return nil, err
		}

		err = writeStringWithSize(buff, friend.StatusMessage, maxStatusMessageSize, 1)
		if err != nil {
			return nil, err
		}

		err = buff.WriteByte(byte(friend.UserStatus))
		if err != nil {
			return nil, err
		}

		//write 3 padding bytes
		_, err = buff.Write([]byte{0, 0, 0})
		if err != nil {
			return nil, err
		}

		err = binary.Write(buff, binary.LittleEndian, friend.Nospam)
		if err != nil {
			return nil, err
		}

		err = binary.Write(buff, binary.BigEndian, friend.LastSeen)
		if err != nil {
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
