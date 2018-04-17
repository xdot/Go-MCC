// Copyright 2017-2018 Andrew Goulas
// https://www.structinf.com
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package gomcc

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"os"
)

type LvlHeader struct {
	Version                          uint16
	Width, Height, Depth             uint16
	SpawnX, SpawnY, SpawnZ           uint16
	SpawnYaw, SpawnPitch             byte
	PermissionVisit, PermissionBuild byte
}

type LvlStorage struct {
	DirectoryPath string
}

func NewLvlStorage(directoryPath string) *LvlStorage {
	os.Mkdir(directoryPath, 0777)
	return &LvlStorage{directoryPath}
}

func (storage *LvlStorage) GetPath(name string) string {
	return storage.DirectoryPath + name + ".lvl"
}

func (storage *LvlStorage) Load(name string) (*Level, error) {
	file, err := os.Open(storage.GetPath(name))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	header := LvlHeader{}
	err = binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	if header.Version != 1874 {
		return nil, errors.New("lvlstorage: invalid format")
	}

	level := NewLevel(name, uint(header.Width), uint(header.Height), uint(header.Depth))
	level.Spawn.X = float64(header.SpawnX) / 32
	level.Spawn.Y = float64(header.SpawnY) / 32
	level.Spawn.Z = float64(header.SpawnZ) / 32
	level.Spawn.Yaw = float64(header.SpawnYaw) * 360 / 256
	level.Spawn.Pitch = float64(header.SpawnPitch) * 360 / 256

	for y := uint(0); y < level.Height; y++ {
		for z := uint(0); z < level.Depth; z++ {
			for x := uint(0); x < level.Width; x++ {
				block := make([]byte, 1)
				n, err := reader.Read(block)
				if n != len(block) && err != nil {
					return nil, err
				}

				level.SetBlock(x, y, z, BlockID(block[0]), false)
			}
		}
	}

	return level, nil
}

func (storage *LvlStorage) Save(level *Level) error {
	file, err := os.Create(storage.GetPath(level.Name))
	if err != nil {
		return err
	}

	writer := gzip.NewWriter(file)
	defer file.Close()
	defer writer.Close()

	err = binary.Write(writer, binary.BigEndian, &LvlHeader{
		1874,
		uint16(level.Width),
		uint16(level.Height),
		uint16(level.Depth),
		uint16(level.Spawn.X * 32),
		uint16(level.Spawn.Y * 32),
		uint16(level.Spawn.Z * 32),
		byte(level.Spawn.Yaw * 256 / 360),
		byte(level.Spawn.Pitch * 256 / 360),
		0, 0,
	})

	if err != nil {
		return err
	}

	for y := uint(0); y < level.Height; y++ {
		for z := uint(0); z < level.Depth; z++ {
			for x := uint(0); x < level.Width; x++ {
				block := make([]byte, 1)
				block[0] = byte(level.GetBlock(x, y, z))

				_, err = writer.Write(block)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
