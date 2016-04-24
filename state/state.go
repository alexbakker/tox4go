package state

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/Impyy/tox4go/crypto"
	"github.com/Impyy/tox4go/dht"
)

type (
	// UserStatus represents the user status.
	UserStatus byte
	// FriendStatus represents the status of a friend request. As you move down
	// the list the current friend status also assumes the previous ones.
	FriendStatus byte
)

const (
	// UserStatusNone indicates that this person didn't specify a status.
	UserStatusNone UserStatus = iota
	// UserStatusAway indicates that this person is away.
	UserStatusAway
	// UserStatusBusy indicates that this person is busy.
	UserStatusBusy
)

const (
	// FriendStatusNone indicates that this friend.
	FriendStatusNone FriendStatus = iota
	// FriendStatusAdded indicates that this friend has been added to the
	// friend list. However, no friend request has been sent yet.
	FriendStatusAdded
	// FriendStatusRequestSent indicates that a friend request has been sent to
	// this friend.
	FriendStatusRequestSent
	// FriendStatusConfirmed indicates that the friend request has been
	// accepted. This friend is now a confirmed friend.
	FriendStatusConfirmed
	// FriendStatusOnline indicates that this friend is currently online.
	FriendStatusOnline
)

const (
	cookieGlobal    = 0x15ED1B1F
	cookieDHTGlobal = 0x159000D
	cookieInner     = 0x01CE
	cookieDHTInner  = 0x11CE

	sectionTypeNospamKeys    = 1
	sectionTypeDHT           = 2
	sectionTypeFriends       = 3
	sectionTypeName          = 4
	sectionTypeStatusMessage = 5
	sectionTypeStatus        = 6
	sectionTypeTCPRelay      = 10
	sectionTypePathNode      = 11
	sectionTypeEnd           = 0xFF

	dhtSectionTypeNodes = 4

	maxNameSize           = 128
	maxStatusMessageSize  = 1007
	maxRequestMessageSize = 1024
)

// State represents a Tox state file.
type State struct {
	PublicKey *[crypto.PublicKeySize]byte
	SecretKey *[crypto.PublicKeySize]byte
	Nospam    uint32

	Name          string
	StatusMessage string
	Status        UserStatus

	Friends   []*Friend
	Nodes     []*dht.Node
	TCPRelays []*dht.Node
	PathNodes []*dht.Node
}

type sectionNospamKeys struct {
	PublicKey *[crypto.PublicKeySize]byte
	SecretKey *[crypto.PublicKeySize]byte
	Nospam    uint32
}

type sectionNodes struct {
	Nodes []*dht.Node
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *State) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)
	dhtBuff := new(bytes.Buffer)

	//write the first 4 zero bytes
	_, err := buff.Write([]byte{0, 0, 0, 0})
	if err != nil {
		return nil, err
	}

	//write the global cookie
	err = binary.Write(buff, binary.LittleEndian, uint32(cookieGlobal))
	if err != nil {
		return nil, err
	}

	//write sectionTypeNospamKeys
	nospamSection := sectionNospamKeys{
		PublicKey: s.PublicKey,
		SecretKey: s.SecretKey,
		Nospam:    s.Nospam,
	}
	bytes, err := nospamSection.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = writeSection(buff, sectionTypeNospamKeys, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeFriends
	friendsSection := sectionFriends{Friends: s.Friends}
	bytes, err = friendsSection.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = writeSection(buff, sectionTypeFriends, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypePathNode
	pathNodeSection := sectionNodes{Nodes: s.PathNodes}
	bytes, err = pathNodeSection.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = writeSection(buff, sectionTypePathNode, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeTCPRelay
	tcpRelaySection := sectionNodes{Nodes: s.TCPRelays}
	bytes, err = tcpRelaySection.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = writeSection(buff, sectionTypeTCPRelay, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeDHT and dhtSectionTypeNodes
	err = binary.Write(dhtBuff, binary.LittleEndian, uint32(cookieDHTGlobal))
	if err != nil {
		return nil, err
	}
	nodesSection := sectionNodes{Nodes: s.Nodes}
	bytes, err = nodesSection.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = writeSection(dhtBuff, dhtSectionTypeNodes, cookieDHTInner, bytes)
	if err != nil {
		return nil, err
	}
	err = writeSection(buff, sectionTypeDHT, cookieInner, dhtBuff.Bytes())
	if err != nil {
		return nil, err
	}

	//write sectionTypeName
	bytes = []byte(s.Name)
	err = writeSection(buff, sectionTypeName, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeStatusMessage
	bytes = []byte(s.StatusMessage)
	err = writeSection(buff, sectionTypeStatusMessage, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeStatus
	bytes = []byte{byte(s.Status)}
	err = writeSection(buff, sectionTypeStatus, cookieInner, bytes)
	if err != nil {
		return nil, err
	}

	//write sectionTypeEnd
	err = writeSection(buff, sectionTypeEnd, cookieInner, []byte{})
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), err
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *State) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	zeroes := make([]byte, 4)
	_, err := reader.Read(zeroes)
	if err != nil {
		return err
	}

	if !bytes.Equal(zeroes, []byte{0, 0, 0, 0}) {
		return errors.New("state file must start with 4 zeroes")
	}

	var cookie uint32
	err = binary.Read(reader, binary.LittleEndian, &cookie)
	if err != nil {
		return err
	} else if cookie != cookieGlobal {
		return GlobalCookieError{actual: cookie, expected: cookieGlobal}
	}

	for {
		sectionType, sectionBody, err := readSection(reader, cookieInner)
		if err != nil {
			return err
		} else if sectionType == sectionTypeEnd {
			return nil
		}

		switch sectionType {
		case sectionTypeNospamKeys:
			section := sectionNospamKeys{}
			err = section.UnmarshalBinary(sectionBody)
			if err != nil {
				return err
			}

			s.Nospam = section.Nospam
			s.PublicKey = section.PublicKey
			s.SecretKey = section.SecretKey
		case sectionTypeFriends:
			friendSection := new(sectionFriends)
			err = friendSection.UnmarshalBinary(sectionBody)
			if err != nil {
				return err
			}

			s.Friends = friendSection.Friends
		case sectionTypeName:
			s.Name = string(sectionBody)
		case sectionTypeStatusMessage:
			s.StatusMessage = string(sectionBody)
		case sectionTypeStatus:
			s.Status = UserStatus(sectionBody[0])
		case sectionTypeTCPRelay:
			section := sectionNodes{}
			err = section.UnmarshalBinary(sectionBody)
			if err != nil {
				return err
			}

			s.TCPRelays = section.Nodes
		case sectionTypePathNode:
			section := sectionNodes{}
			err = section.UnmarshalBinary(sectionBody)
			if err != nil {
				return err
			}

			s.PathNodes = section.Nodes
		case sectionTypeDHT:
			dhtReader := bytes.NewReader(sectionBody)
			var dhtCookie uint32

			err = binary.Read(dhtReader, binary.LittleEndian, &dhtCookie)
			if err != nil {
				return err
			} else if dhtCookie != cookieDHTGlobal {
				return GlobalCookieError{actual: dhtCookie, expected: cookieDHTGlobal}
			}

			for dhtReader.Len() > 0 {
				dhtSectionType, dhtSectionBody, err := readSection(dhtReader, cookieDHTInner)
				if err != nil {
					return err
				}

				switch dhtSectionType {
				case dhtSectionTypeNodes:
					section := sectionNodes{}
					err = section.UnmarshalBinary(dhtSectionBody)
					if err != nil {
						return err
					}

					s.Nodes = section.Nodes
				default:
					//unknown dht section type, ignore it
				}
			}
		default:
			//unknown section type, ignore it
		}
	}
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *sectionNospamKeys) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.LittleEndian, &s.Nospam)
	if err != nil {
		return err
	}

	s.PublicKey = new([crypto.PublicKeySize]byte)
	_, err = reader.Read(s.PublicKey[:])
	if err != nil {
		return err
	}

	s.SecretKey = new([crypto.SecretKeySize]byte)
	_, err = reader.Read(s.SecretKey[:])
	return err
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *sectionNospamKeys) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.LittleEndian, s.Nospam)
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(s.PublicKey[:])
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(s.SecretKey[:])
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *sectionNodes) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		var ipType byte
		var ipSize int

		node := new(dht.Node)

		err := binary.Read(reader, binary.BigEndian, &ipType)
		if err != nil {
			return err
		}

		switch ipType {
		case 2, 130: //ipv4
			ipSize = net.IPv4len
		case 10, 138: //ipv6
			ipSize = net.IPv6len
		default:
			return fmt.Errorf("unknown address family: %d", ipType)
		}

		nodeBytes := make([]byte, 1+ipSize+2+crypto.PublicKeySize)
		nodeBytes[0] = ipType
		_, err = reader.Read(nodeBytes[1:])
		if err != nil {
			return err
		}

		err = node.UnmarshalBinary(nodeBytes)
		if err != nil {
			return err
		}

		s.Nodes = append(s.Nodes, node)
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s *sectionNodes) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	for _, node := range s.Nodes {
		nodeBytes, err := node.MarshalBinary()
		if err != nil {
			return nil, err
		}

		_, err = buff.Write(nodeBytes)
		if err != nil {
			return nil, err
		}
	}

	return buff.Bytes(), nil
}

func writeSection(writer io.Writer, sectionType uint16, cookie uint16, body []byte) error {
	err := binary.Write(writer, binary.LittleEndian, uint32(len(body)))
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, sectionType)
	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.LittleEndian, cookie)
	if err != nil {
		return err
	}

	_, err = writer.Write(body)
	return err
}

func readSection(reader io.Reader, expectedCookie uint16) (uint16, []byte, error) {
	var length uint32
	err := binary.Read(reader, binary.LittleEndian, &length)
	if err != nil {
		return 0, nil, err
	}

	var sectionType uint16
	err = binary.Read(reader, binary.LittleEndian, &sectionType)
	if err != nil {
		return 0, nil, err
	} else if sectionType == sectionTypeEnd {
		//this means we've reached the end of the state file
		return sectionTypeEnd, nil, nil
	}

	var cookie uint16
	err = binary.Read(reader, binary.LittleEndian, &cookie)
	if err != nil {
		return 0, nil, err
	} else if cookie != expectedCookie {
		return 0, nil, InnerCookieError{actual: cookie, expected: expectedCookie}
	}

	sectionBody := make([]byte, length)
	_, err = reader.Read(sectionBody)
	if err != nil {
		return 0, nil, err
	}

	return sectionType, sectionBody, nil
}

func writeStringWithSize(writer io.Writer, toWrite string, size uint16, paddingSize int) error {
	bytesToWrite := []byte(toWrite)
	data := make([]byte, size)
	copy(data, bytesToWrite)

	_, err := writer.Write(data)
	if err != nil {
		return err
	}

	if paddingSize > 0 {
		padding := make([]byte, paddingSize)
		_, err = writer.Write(padding)

		if err != nil {
			return err
		}
	}

	err = binary.Write(writer, binary.BigEndian, uint16(len(bytesToWrite)))
	return err
}

func readStringWithSize(reader io.ReadSeeker, size uint16, paddingSize int) (string, error) {
	str := make([]byte, size)
	_, err := reader.Read(str)
	if err != nil {
		return "", err
	}

	//skip padding
	if paddingSize > 0 {
		_, err = reader.Seek(int64(paddingSize), 1)
		if err != nil {
			return "", err
		}
	}

	var strSize uint16
	err = binary.Read(reader, binary.BigEndian, &strSize)
	if err != nil {
		return "", err
	} else if int(strSize) > len(str) {
		return "", fmt.Errorf("invalid string size: %d > %d", strSize, len(str))
	}

	return string(str[:strSize]), nil
}
