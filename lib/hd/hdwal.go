package hd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/tyler-smith/go-bip39"
)

type MnemonicType int

const (
	Mnemonic12 MnemonicType = iota
	Mnemonic24
)

func (m MnemonicType) String() string {
	switch m {
	case Mnemonic12:
		return "12 mnemonics"
	case Mnemonic24:
		return "24 mnemonics"
	}

	panic("Unexpected MnemonicType")
}

func entropy(mt MnemonicType) ([]byte, error) {
	randomBytes := make([]byte, 0)
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memory, _ := mem.VirtualMemory()

	ioCounters, _ := net.IOCounters(true)
	netWork := strconv.Itoa(int(ioCounters[0].BytesSent + ioCounters[0].BytesRecv))

	cRandBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, cRandBytes); err != nil {
		return []byte{}, err
	}

	randomBytes = append(randomBytes, cRandBytes...)
	randomBytes = append(randomBytes, float64ToByte(cpuPercent[0])...)
	randomBytes = append(randomBytes, float64ToByte(memory.UsedPercent)...)
	randomBytes = append(randomBytes, []byte(netWork)...)

	random := sha256.Sum256(randomBytes)

	switch mt {
	case Mnemonic12:
		return random[:16], nil
	case Mnemonic24:
		return random[:32], nil
	default:
		return nil, fmt.Errorf("MnemonicType err: %d", mt)
	}
}

func NewMnemonic(mt MnemonicType) (string, error) {
	entropyBytes, err := entropy(mt)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropyBytes)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

func CheckMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

func GenerateSeedFromMnemonic(mnemonic, password string) ([]byte, error) {
	seedBytes, err := bip39.NewSeedWithErrorChecking(mnemonic, password)
	if err != nil {
		return []byte{}, err
	}

	return seedBytes, nil
}

func GetExtendSeedFromPath(path string, seed []byte) ([]byte, error) {
	extendedKey, err := NewMaster(seed)
	if err != nil {
		return nil, err
	}

	derivationPath, err := ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	for _, index := range derivationPath {
		childExtendedKey, err := extendedKey.Child(index)
		if err != nil {
			return nil, err
		}
		extendedKey = childExtendedKey
	}

	return extendedKey.key, nil
}

func float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
